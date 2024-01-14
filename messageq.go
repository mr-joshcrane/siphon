package siphon

import (
	"bytes"
)

type Queue interface {
	Enqueue(string)
	Dequeue() string
}

type MemoryQueue struct {
	Buf *bytes.Buffer
}

func (q *MemoryQueue) Enqueue(s string) {
	q.Buf.WriteString(s)
}

func (q *MemoryQueue) Dequeue() string {
	b := q.Buf.Bytes()
	if len(b) == 0 {
		return ""
	}
	item := string(b[0])
	q.Buf = bytes.NewBuffer(b[1:])
	return item
}
