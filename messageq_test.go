package siphon_test

import (
	"bytes"
	"testing"

	"github.com/mr-joshcrane/siphon"
)

func TestEnqueue_OrderIsCorrect(t *testing.T) {
	t.Parallel()
	buf := new(bytes.Buffer)
	q := siphon.MemoryQueue{Buf: buf}
	q.Enqueue("a")
	q.Enqueue("b")
	q.Enqueue("c")


	if buf.String() != "abc" {
		t.Errorf("Expected queue to be \"abc\", got %q", buf.String())
	}
}

func TestEnqueue_SizeIncrementsCorrectly(t *testing.T) {
	t.Parallel()
	buf := new(bytes.Buffer)
	q := siphon.MemoryQueue{Buf: buf}
	if buf.Len() != 0 {
		t.Errorf("Expected queue to be empty, got %q", buf.String())
	}
	q.Enqueue("a")
	if buf.Len() != 1 {
		t.Errorf("Expected queue to have 1 element, got %q", buf.String())
	}
	q.Enqueue("b")
	if buf.Len() != 2 {
		t.Errorf("Expected queue to have 2 elements, got %q", buf.String())
	}
}

func TestDequeue_IsFirstInFirstOut(t *testing.T) {
	t.Parallel()
	buf := helperTestBuffer()
	q := siphon.MemoryQueue{Buf: buf}

	item := q.Dequeue()
	if item != "a" {
		t.Errorf("Expected first item to be \"a\", got %q", item)
	}
	item = q.Dequeue()
	if item != "b" {
		t.Errorf("Expected first item to be \"b\", got %q", item)
	}
	item = q.Dequeue()
	if item != "c" {
		t.Errorf("Expected first item to be \"c\", got %q", item)
	}
}

func TestDequeue_SizeDecrementsCorrectly(t *testing.T) {
	t.Parallel()
	buf := helperTestBuffer()
	q := siphon.MemoryQueue{Buf: buf}

	_ = q.Dequeue()
	if buf.Len() != 2 {
		t.Errorf("Expected queue to have 2 elements, got %d", buf.Len())
	}
	_ = q.Dequeue()
	if buf.Len() != 1 {
		t.Errorf("Expected queue to have 1 element, got %d", buf.Len())
	}
	_ = q.Dequeue()
	if buf.Len() != 0 {
		t.Errorf("Expected queue to have 0 elements, got %d", buf.Len())
	}
}


func TestEmptyQueueDequeue(t *testing.T) {
	t.Parallel()
}

func TestSize(t *testing.T) {
	t.Parallel()
}

func TestOrdering(t *testing.T) {
	t.Parallel()
}

func TestConcurrency(t *testing.T) {
	t.Parallel()
}

func helperTestBuffer() *bytes.Buffer {
	buf := new(bytes.Buffer)
	buf.WriteString("a")
	buf.WriteString("b")
	buf.WriteString("c")
	return buf
}
