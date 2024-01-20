package siphon_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/mr-joshcrane/siphon"
)

func TestEnqueue_OrderIsCorrect(t *testing.T) {
	t.Parallel()
	buf := new(bytes.Buffer)
	q := siphon.NewMemoryQueue(buf)
	_ = q.Enqueue("a")
	_ = q.Enqueue("b")
	_ = q.Enqueue("c")

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
	_ = q.Enqueue("a")
	if buf.Len() != 5 {
		t.Errorf("Expected 5 bytes (1 byte data, 4 byte prefix), got %d", buf.Len())
	}
	_ = q.Enqueue("b")
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
	buf := new(bytes.Buffer)
	q := siphon.NewMemoryQueue(buf)
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = q.Enqueue("a")
	}()
	item, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	if item != "a" {
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
	_ = q.Enqueue("first string")
	if q.Size() != 1 {
		t.Errorf("Expected queue to have 1 item, got %d", q.Size())
	}
	_ = q.Enqueue("second string")
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
}

func TestQueueClient_Enqueue(t *testing.T) {
	t.Parallel()
	c := helperTestClient()
	err := c.Publish("a")
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	err = c.Publish("b")
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	err = c.Publish("c")
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	got, err := io.ReadAll(c.Conn)
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	want := []byte{0, 0, 0, 1, 'a', 0, 0, 0, 1, 'b', 0, 0, 0, 1, 'c'}

	if !cmp.Equal(got, want) {
		t.Errorf("Expected queue to be %q, got %q", want, got)
	}
}

func TestQueueClient_Dequeue(t *testing.T) {
	t.Parallel()
	c := helperTestClient("a", "b", "c")
	got, err := c.Receive()
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	if got != "a" {
		t.Errorf("Expected first item to be \"a\", got %q", got)
	}
	got, err = c.Receive()
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	if got != "b" {
		t.Errorf("Expected first item to be \"b\", got %q", got)
	}
	got, err = c.Receive()
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	if got != "c" {
		t.Errorf("Expected first item to be \"c\", got %q", got)
	}
}

func TestHandleConn_Publish(t *testing.T) {
	addr, q, cleanup := helperTCPServerWithConn(t)
	defer cleanup()
	c, err := siphon.GetQueue(addr, "test")
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	err = c.Publish("a")
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	err = c.Publish("b")
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	waitForQueueSize(q, 2)
	got := q.Buf.Bytes()
	want := []byte{0, 0, 0, 1, 'a', 0, 0, 0, 1, 'b'}
	if !cmp.Equal(got, want) {
		t.Errorf("Expected queue to be %q, got %q", want, got)
	}
}

func TestHandleConn_Receive(t *testing.T) {
	t.Parallel()
	addr, q, cleanup := helperTCPServerWithConn(t)
	defer cleanup()
	c, err := siphon.GetQueue(addr, "test")
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	_ = q.Enqueue("a")
	_ = q.Enqueue("b")
	_ = q.Enqueue("c")
	waitForQueueSize(q, 3)
	got, err := c.Receive()
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	if got != "a" {
		t.Errorf("Expected first item to be \"a\", got %q", got)
	}
	got, err = c.Receive()
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	if got != "b" {
		t.Errorf("Expected second item to be \"b\", got %q", got)
	}
	got, err = c.Receive()
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	if got != "c" {
		t.Errorf("Expected third item to be \"c\", got %q", got)
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
			_ = q.Enqueue(fmt.Sprintf("string %d", i))
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
		_ = q.Enqueue("a")
		_, _ = q.Dequeue()
	}
}

func helperTestBuffer() *bytes.Buffer {
	return bytes.NewBuffer([]byte{0, 0, 0, 1, 'a', 0, 0, 0, 1, 'b', 0, 0, 0, 1, 'c'})
}

func helperTestClient(items ...string) *siphon.Client {
	c := &siphon.Client{
		Conn: new(bytes.Buffer),
	}
	for _, item := range items {
		_ = c.Publish(item)
	}
	return c

}

func helperTCPServerWithConn(t *testing.T) (addr string, q *siphon.MemoryQueue, cleanup func()) {
	t.Helper()
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Expected no error, got %q", err)
	}
	q = siphon.NewMemoryQueue(new(bytes.Buffer))
	go func() {
		conn, _ := l.Accept()
		siphon.HandleConn(conn, q)
	}()
	return l.Addr().String(), q, func() {
		l.Close()
	}
}

func waitForQueueSize(q *siphon.MemoryQueue, size int) {
	ctx, cancel := context.WithTimeout(context.TODO(), 2000*time.Millisecond)
	defer cancel()
	for ctx.Err() == nil {
		if q.Size() == size {
			cancel()
		}
	}
}
