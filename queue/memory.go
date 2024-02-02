package queue

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
)

type MemoryQueue struct {
	Buf     *bytes.Buffer
	size    int
	mu      *sync.RWMutex
	isEmpty *sync.Cond
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
func (q *MemoryQueue) Signal() {
	q.isEmpty.Signal()
}

func (q *MemoryQueue) Enqueue(items ...string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, item := range items {
		length := len(item)
		sizeHint := make([]byte, 4)
		binary.BigEndian.PutUint32(sizeHint, uint32(length))
		q.Buf.Write(sizeHint)
		q.Buf.WriteString(item)
		q.size++
	}
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
