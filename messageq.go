package siphon

import (
	"encoding/binary"
	"errors"
	"net"
)

var ErrServerClosed = errors.New("siphon: server closed")

type Queue interface {
	Enqueue(...string) error
	Dequeue() (string, error)
	Size() int
}

func GetQueue(addr string, name string) (*NetworkClient, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &NetworkClient{
		Conn: conn,
	}, nil
}

func messageToBytes(msg string) []byte {
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(len(msg)))
	return append(size, []byte(msg)...)
}
