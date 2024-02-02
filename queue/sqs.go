package queue


import (
	"os"
	"fmt"
	"strconv"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)
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

func (q *AWSQueue) Enqueue(s ...string) error {
	for _, msg := range s {
		if msg == "" {
			continue
		}
		_, err := q.client.SendMessage(&sqs.SendMessageInput{
			MessageBody: aws.String(msg),
			QueueUrl:    q.queueURL,
		})
		if err != nil {
			return fmt.Errorf("error sending message: %q", err)
		}
	}
	return nil
}

func (q *AWSQueue) Dequeue() (string, error) {
	for {
		received, err := q.client.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            q.queueURL,
			WaitTimeSeconds:     aws.Int64(20),
			MaxNumberOfMessages: aws.Int64(10),
		})
		if err != nil {
			return "", fmt.Errorf("error receiving message: %q", err)
		}
		if len(received.Messages) == 0 {
			continue
		}
		var msg string
		for _, message := range received.Messages {
			msg += *message.Body
		}

		go func() {
			entries := make([]*sqs.DeleteMessageBatchRequestEntry, len(received.Messages))
			for i, message := range received.Messages {
				entries[i] = &sqs.DeleteMessageBatchRequestEntry{
					Id:            message.MessageId,
					ReceiptHandle: message.ReceiptHandle,
				}
			}

			_, err = q.client.DeleteMessageBatch(&sqs.DeleteMessageBatchInput{
				QueueUrl: q.queueURL,
				Entries:  entries,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "error deleting message: %q", err)
			}
		}()
		return msg, nil
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


