package siphon

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Queue interface {
	Enqueue(string)
	Dequeue() string
}

type MemoryQueue struct {
	Buf *bytes.Buffer
}

func (q *MemoryQueue) Enqueue(s string) {
	length := len(s)
	sizeHint := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeHint, uint32(length))
	q.Buf.Write(sizeHint)
	q.Buf.WriteString(s)
}

func (q *MemoryQueue) Dequeue() (string, error) {
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
	return string(message), nil
}
