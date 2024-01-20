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

type command int

const (
	Publish command = iota
	Receive
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
		go HandleConn(conn, q)
	}
	return ErrServerClosed
}

func HandleConn(conn net.Conn, q *MemoryQueue) {
	worker, err := GetWorker(conn)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	for {
		err := worker.ClientResponse(q)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
	}

}

func messageToBytes(msg string) []byte {
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(len(msg)))
	return append(size, []byte(msg)...)
}

type MessagePasser interface {
	Receive() (string, error)
	Publish(string) error
}

type Client struct {
	Conn io.ReadWriter
}

type Worker struct {
	Conn io.ReadWriter
	Client
}

func (w *Worker) ClientResponse(q *MemoryQueue) error {
	length := make([]byte, 4)
	_, err := io.ReadFull(w.Conn, length)
	if err != nil {
		return err
	}
	size := binary.BigEndian.Uint32(length)
	if size == 0 {
		if err != nil {
			return err
		}
		msg, err := q.Dequeue()
		if err != nil {
			return err
		}
		_, err = w.Conn.Write(messageToBytes(msg))
		if err != nil {
			return err
		}
	} else {
		message := make([]byte, size)
		_, err = io.ReadFull(w.Conn, message)
		if err != nil {
			return err
		}
		msg := string(message)

		err = q.Enqueue(msg)
		if err != nil {
			return err
		}
	}
	return nil
}
func (q *Client) Publish(s string) error {
	data := messageToBytes(s)
	_, err := q.Conn.Write(data)
	return err
}

func (q *Client) Receive() (string, error) {
	length := make([]byte, 4)
	q.Conn.Write(length)
	_, err := io.ReadFull(q.Conn, length)
	if err != nil {
		return "", err
	}
	size := binary.BigEndian.Uint32(length)
	if size == 0 {
		return "", nil
	}
	message := make([]byte, size)
	_, err = io.ReadFull(q.Conn, message)
	if err != nil {
		return "", err
	}
	return string(message), nil
}

func GetWorker(conn io.ReadWriter) (*Worker, error) {
	return &Worker{
		Conn: conn,
	}, nil
}

func GetQueue(addr string, name string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Client{
		Conn: conn,
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
