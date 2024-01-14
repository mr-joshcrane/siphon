package siphon

import (
	"bytes"
)

type Queue interface {
	Enqueue(string)
}

type MemoryQueue struct {
	Buf *bytes.Buffer
}

func (q *MemoryQueue) Enqueue(s string) {
	q.Buf.WriteString(s)
}
