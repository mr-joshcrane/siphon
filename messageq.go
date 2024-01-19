package siphon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
)

var ErrServerClosed = errors.New("siphon: server closed")

func ListenAndServe(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()
	q := NewMemoryQueue(new(bytes.Buffer))
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go handleConn(conn, *q)
	}
	return ErrServerClosed
}

func handleConn(conn net.Conn, q MemoryQueue) {
	defer conn.Close()
	for {
		length := make([]byte, 4)
		_, err := io.ReadFull(conn, length)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		size := binary.BigEndian.Uint32(length)
		if size == 0 {
			msg, err := q.Dequeue()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
			length := make([]byte, 4)
			binary.BigEndian.PutUint32(length, uint32(len(msg)))
			message := append(length, []byte(msg)...)
			_, err = conn.Write(message)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
			continue
		}
		message := make([]byte, size)
		_, err = io.ReadFull(conn, message)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		err = q.Enqueue(string(message))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		q.Signal()
	}
}

type QueueClient struct {
	conn *MemoryQueue
}

func (q *QueueClient) Publish(s string) error {
	err := q.conn.Enqueue(s)
	return err
}

func (q *QueueClient) Receive() (string, error) {
	msg, err := q.conn.Dequeue()
	if err != nil {
		return "", err
	}
	return string(msg), nil
}

func GetQueue(q *MemoryQueue, name string) (*QueueClient, error) {
	return &QueueClient{
		conn: q,
	}, nil
}

func NewMemoryQueue(buf *bytes.Buffer) *MemoryQueue {
	mu := &sync.RWMutex{}
	cond := sync.NewCond(mu)
	return &MemoryQueue{
		Buf:     buf,
		size:    len(buf.Bytes()),
		isEmpty: cond,
		mu:      mu,
	}
}

type MemoryQueue struct {
	Buf     *bytes.Buffer
	size    int
	mu      *sync.RWMutex
	isEmpty *sync.Cond
}

func (q *MemoryQueue) Signal() {
	q.isEmpty.Signal()
}

func (q *MemoryQueue) Enqueue(s string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	length := len(s)
	sizeHint := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeHint, uint32(length))
	q.Buf.Write(sizeHint)
	q.Buf.WriteString(s)
	q.size++
	q.Signal()
	return nil
}

func (q *MemoryQueue) Dequeue() (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for q.Size() == 0 {
		q.isEmpty.Wait()
	}
	sizeHint := make([]byte, 4)
	_, err := q.Buf.Read(sizeHint)
	if err != nil {
		return "", fmt.Errorf("error reading size hint: %q", err)
	}
	messageSize := binary.BigEndian.Uint32(sizeHint)
	message := make([]byte, messageSize)
	_, err = q.Buf.Read(message)
	if err != nil {
		return "", fmt.Errorf("error reading message: %q", err)
	}
	q.size--
	return string(message), nil
}

func (q *MemoryQueue) Size() int {
	return q.size
}
