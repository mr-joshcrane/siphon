package siphon

import (
	"bytes"
	"fmt"
	"net"
	"os"

	"github.com/mr-joshcrane/siphon/queue"
)

type QueueServer struct {
	Addr string
	Q    Queue
}

type Options func(*QueueServer) *QueueServer

func WithQueue(q Queue) Options {
	return func(s *QueueServer) *QueueServer {
		s.Q = q
		return s
	}
}

func NewServer(addr string, opt ...Options) *QueueServer {
	q := queue.NewMemoryQueue(new(bytes.Buffer))
	s := &QueueServer{
		Addr: addr,
		Q:    q,
	}
	for _, o := range opt {
		s = o(s)
	}
	return s
}

func (s *QueueServer) ListenAndServe() error {
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go HandleConn(conn, s.Q)
	}
	return ErrServerClosed
}

func HandleConn(conn net.Conn, q Queue) {
	worker, err := GetWorker(conn, q)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	for {
		err := worker.ClientResponse()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
	}
}
