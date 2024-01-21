package siphon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

var ErrServerClosed = errors.New("siphon: server closed")

func ListenAndServe(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()
	q, err := NewAWSQueue("testEvents")
	if err != nil {
		return err
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go HandleConn(conn, q)
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

func messageToBytes(msg string) []byte {
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(len(msg)))
	return append(size, []byte(msg)...)
}

type MessagePasser interface {
	Receive() (string, error)
	Publish(string) error
}

type Client struct {
	Conn io.ReadWriter
}

type Worker struct {
	Conn io.ReadWriter
	Q    Queue
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

func (q *Client) Publish(s string) error {
	data := messageToBytes(s)
	_, err := q.Conn.Write(data)
	return err
}

func (q *Client) Receive() (string, error) {
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

func GetWorker(conn io.ReadWriter, q Queue) (*Worker, error) {
	return &Worker{
		Conn: conn,
		Q:    q,
	}, nil
}

func GetQueue(addr string, name string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Client{
		Conn: conn,
	}, nil
}

func NewMemoryQueue(buf *bytes.Buffer) *MemoryQueue {
	mu := &sync.RWMutex{}
	cond := sync.NewCond(mu)
	return &MemoryQueue{
		Buf:     buf,
		size:    len(buf.Bytes()),
		isEmpty: cond,
		mu:      mu,
	}
}

type Queue interface {
	Enqueue(string) error
	Dequeue() (string, error)
	Size() int
}

type AWSQueue struct {
	client    *sqs.SQS
	queueURL  *string
	QueueName string
}

func NewAWSQueue(queueName string) (*AWSQueue, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating session: %q", err)
	}
	client := sqs.New(sess)
	queueURL := os.Getenv("QUEUE_URL")
	if queueURL == "" {
		return nil, fmt.Errorf("error getting QUEUE_URL from environment")
	}
	return &AWSQueue{
		client:    client,
		queueURL:  &queueURL,
		QueueName: queueName,
	}, nil
}

func (q *AWSQueue) Enqueue(s string) error {
	_, err := q.client.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(s),
		QueueUrl:    q.queueURL,
	})
	if err != nil {
		return fmt.Errorf("error sending message: %q", err)
	}
	return nil
}

func (q *AWSQueue) Dequeue() (string, error) {
	for {
		received, err := q.client.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:        q.queueURL,
			WaitTimeSeconds: aws.Int64(20),
		})
		if err != nil {
			return "", fmt.Errorf("error receiving message: %q", err)
		}
		if len(received.Messages) == 0 {
			continue
		}
		message := received.Messages[0]
		_, err = q.client.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      q.queueURL,
			ReceiptHandle: message.ReceiptHandle,
		})
		if err != nil {
			return "", fmt.Errorf("error deleting message: %q", err)
		}
		return *message.Body, nil
	}
}

func (q *AWSQueue) Size() int {
	attr, err := q.client.GetQueueAttributes(&sqs.GetQueueAttributesInput{
		QueueUrl: q.queueURL,
		AttributeNames: []*string{
			aws.String("ApproximateNumberOfMessages"),
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting queue attributes: %q", err)
		return -1
	}
	length := *attr.Attributes["ApproximateNumberOfMessages"]
	l, err := strconv.Atoi(length)
	if err != nil {
		return -1
	}
	return l
}

type MemoryQueue struct {
	Buf     *bytes.Buffer
	size    int
	mu      *sync.RWMutex
	isEmpty *sync.Cond
}

func (q *MemoryQueue) Signal() {
	q.isEmpty.Signal()
}

func (q *MemoryQueue) Enqueue(s string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	length := len(s)
	sizeHint := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeHint, uint32(length))
	q.Buf.Write(sizeHint)
	q.Buf.WriteString(s)
	q.size++
	q.Signal()
	return nil
}

func (q *MemoryQueue) Dequeue() (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for q.Size() == 0 {
		q.isEmpty.Wait()
	}
	sizeHint := make([]byte, 4)
	_, err := q.Buf.Read(sizeHint)
	if err != nil {
		return "", fmt.Errorf("error reading size hint: %q", err)
	}
	messageSize := binary.BigEndian.Uint32(sizeHint)
	message := make([]byte, messageSize)
	_, err = q.Buf.Read(message)
	if err != nil {
		return "", fmt.Errorf("error reading message: %q", err)
	}
	q.size--
	return string(message), nil
}

func (q *MemoryQueue) Size() int {
	return q.size
}
