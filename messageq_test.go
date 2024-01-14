package siphon_test

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mr-joshcrane/siphon"
)

func TestEnqueue_OrderIsCorrect(t *testing.T) {
	t.Parallel()
	buf := new(bytes.Buffer)
	q := siphon.MemoryQueue{Buf: buf}
	q.Enqueue("a")
	q.Enqueue("b")
	q.Enqueue("c")

	want := []byte{0, 0, 0, 1, 'a', 0, 0, 0, 1, 'b', 0, 0, 0, 1, 'c'}

	if !cmp.Equal(buf.Bytes(), want) {
		t.Errorf("Expected queue to be %q, got %q", want, buf.Bytes())
	}
}

func TestEnqueue_SizeIncrementsCorrectly(t *testing.T) {
	t.Parallel()
	buf := new(bytes.Buffer)
	q := siphon.MemoryQueue{Buf: buf}
	if buf.Len() != 0 {
		t.Errorf("Expected queue to be empty, got %d", buf.Len())
	}
	q.Enqueue("a")
	if buf.Len() != 5 {
		t.Errorf("Expected 5 bytes (1 byte data, 4 byte prefix), got %d", buf.Len())
	}
	q.Enqueue("b")
	if buf.Len() != 10 {
		t.Errorf("Expected 10 bytes 2x (1 byte data, 4 byte prefix), got %d", buf.Len())
	}
}

func TestDequeue_IsFirstInFirstOut(t *testing.T) {
	t.Parallel()
	buf := helperTestBuffer()
	q := siphon.MemoryQueue{Buf: buf}

	item, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	if item != "a" {
		t.Errorf("Expected first item to be \"a\", got %q", item)
	}
	item, err = q.Dequeue()
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	if item != "b" {
		t.Errorf("Expected first item to be \"b\", got %q", item)
	}
	item, err = q.Dequeue()
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	if item != "c" {
		t.Errorf("Expected first item to be \"c\", got %q", item)
	}
}

func TestDequeue_SizeDecrementsCorrectly(t *testing.T) {
	t.Parallel()
	buf := helperTestBuffer()
	q := siphon.MemoryQueue{Buf: buf}

	_, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	if buf.Len() != 10 {
		t.Errorf("Expected 10 bytes 2x (1 byte data, 4 byte prefix), got %d", buf.Len())
	}
	_, err = q.Dequeue()
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	if buf.Len() != 5 {
		t.Errorf("Expected 5 bytes (1 byte data, 4 byte prefix), got %d", buf.Len())
	}
	_, err = q.Dequeue()
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	if buf.Len() != 0 {
		t.Errorf("Expected queue to have 0 bytes, got %d", buf.Len())
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
	return bytes.NewBuffer([]byte{0, 0, 0, 1, 'a', 0, 0, 0, 1, 'b', 0, 0, 0, 1, 'c'})
}
