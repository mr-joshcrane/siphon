package siphon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
)

type Queue interface {
	Enqueue(string)
	Dequeue() string
	Size() int
}

func NewMemoryQueue(buf *bytes.Buffer) *MemoryQueue {
	return &MemoryQueue{
		Buf:  buf,
		size: 0,
		mu:   &sync.RWMutex{},
	}
}

type MemoryQueue struct {
	Buf  *bytes.Buffer
	size int
	mu   *sync.RWMutex
}

func (q *MemoryQueue) Enqueue(s string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	length := len(s)
	sizeHint := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeHint, uint32(length))
	q.Buf.Write(sizeHint)
	q.Buf.WriteString(s)
	q.size++
}

func (q *MemoryQueue) Dequeue() (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.Buf.Bytes()) == 0 {
		// empty queue is not an error
		return "", nil
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
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.size
}
