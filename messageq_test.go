package siphon_test

import (
	"bytes"
	"fmt"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mr-joshcrane/siphon"
)

func TestEnqueue_OrderIsCorrect(t *testing.T) {
	t.Parallel()
	buf := new(bytes.Buffer)
	q := siphon.NewMemoryQueue(buf)
	q.Enqueue("a")
	q.Enqueue("b")
	q.Enqueue("c")

	want := []byte{0, 0, 0, 1, 'a', 0, 0, 0, 1, 'b', 0, 0, 0, 1, 'c'}

	if !cmp.Equal(buf.Bytes(), want) {
		t.Errorf("Expected queue to be %q, got %q", want, buf.Bytes())
	}
}

func TestEnqueue_BytesIncrease(t *testing.T) {
	t.Parallel()
	buf := new(bytes.Buffer)
	q := siphon.NewMemoryQueue(buf)
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
	q := siphon.NewMemoryQueue(buf)
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

func TestDequeue_BytesDecrease(t *testing.T) {
	t.Parallel()
	buf := helperTestBuffer()
	q := siphon.NewMemoryQueue(buf)
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
	buf := new(bytes.Buffer)
	q := siphon.NewMemoryQueue(buf)
	item, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	if item != "" {
		t.Errorf("Expected empty string, got %q", item)
	}
}

func TestSize(t *testing.T) {
	t.Parallel()
	buf := new(bytes.Buffer)
	q := siphon.NewMemoryQueue(buf)
	if q.Size() != 0 {
		t.Errorf("Expected queue to be empty, got %d", q.Size())
	}
	q.Enqueue("first string")
	if q.Size() != 1 {
		t.Errorf("Expected queue to have 1 item, got %d", q.Size())
	}
	q.Enqueue("second string")
	if q.Size() != 2 {
		t.Errorf("Expected queue to have 2 items, got %d", q.Size())
	}
	_, _ = q.Dequeue()
	_, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	if q.Size() != 0 {
		t.Errorf("Expected queue to have 0 item, got %d", q.Size())
	}
	_, _ = q.Dequeue()
	if q.Size() != 0 {
		t.Errorf("Expected queue to be empty after empty dequeue, got %d", q.Size())
	}
}

func TestConcurrency_IsThreadSafe(t *testing.T) {
	t.Parallel()
	buf := new(bytes.Buffer)
	q := siphon.NewMemoryQueue(buf)
	wg := &sync.WaitGroup{}
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func(i int) {
			q.Enqueue(fmt.Sprintf("string %d", i))
			wg.Done()
		}(i)
	}
	wg.Wait()
	if q.Size() != 100 {
		t.Errorf("expected queue to have 100 items, got %d", q.Size())
	}
}

func BenchmarkEnqueue(b *testing.B) {
	buf := new(bytes.Buffer)
	q := siphon.NewMemoryQueue(buf)
	for i := 0; i < b.N; i++ {
		q.Enqueue("a")
		q.Dequeue()
	}
}

//
// func BenchmarkEnqueueConcurrent(b *testing.B) {
// 	buf := new(bytes.Buffer)
// 	q := siphon.NewMemoryQueue(buf)
// 	for i := 0; i < b.N; i++ {
// 		go func() {
// 			q.Enqueue("a")
// 			q.Dequeue()
// 		}()
//
//	}
//}

func helperTestBuffer() *bytes.Buffer {
	return bytes.NewBuffer([]byte{0, 0, 0, 1, 'a', 0, 0, 0, 1, 'b', 0, 0, 0, 1, 'c'})
}
