package siphon

import (
	"encoding/binary"
	"io"
)

type MessageClient interface {
	Receive() (string, error)
	Publish(string) error
}

type NetworkClient struct {
	Conn io.ReadWriter
	q		Queue
}

func (q *NetworkClient) Publish(s string) error {
	data := messageToBytes(s)
	_, err := q.Conn.Write(data)
	return err
}

func (q *NetworkClient) Receive() (string, error) {
	length := make([]byte, 4)
	q.Conn.Write(length)
	_, err := io.ReadFull(q.Conn, length)
	if err != nil {
		return "", err
	}
	size := binary.BigEndian.Uint32(length)
	if size == 0 {
		return "", nil
	}
	message := make([]byte, size)
	_, err = io.ReadFull(q.Conn, message)
	if err != nil {
		return "", err
	}
	return string(message), nil
}
