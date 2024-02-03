package siphon

import (
	"encoding/binary"
	"io"
	"net"
	"strconv"
)

type Worker struct {
	Conn io.ReadWriter
	Q    Queue
}

func GetWorker(conn net.Conn, q Queue) (*Worker, error) {
	return &Worker{
		Conn: conn,
		Q:    q,
	}, nil
}
func (w *Worker) Publish(s string) error {
	size, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	message := make([]byte, size)
	_, err = io.ReadFull(w.Conn, message)
	if err != nil {
		return err
	}
	msg := string(message)
	err = w.Q.Enqueue(msg)
	if err != nil {
		return err
	}
	return nil
}

func (w *Worker) Receive() (string, error) {
	msg, err := w.Q.Dequeue()
	if err != nil {
		return "", err
	}
	_, err = w.Conn.Write(messageToBytes(msg))
	if err != nil {
		return "", err
	}
	return msg, nil

}

func (w *Worker) ClientResponse() error {
	var err error
	length := make([]byte, 4)
	_, err = io.ReadFull(w.Conn, length)
	if err != nil {
		return err
	}
	size := binary.BigEndian.Uint32(length)
	if size == 0 {
		_, err = w.Receive()
		if err != nil {
			return err
		}
		return nil
	}
	return w.Publish(strconv.Itoa(int(size)))
}
