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

func TestDequeue(t *testing.T) {
	t.Parallel()
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
