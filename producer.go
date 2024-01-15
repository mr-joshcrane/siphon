package siphon

import (
	"bytes"
	"fmt"
	"sync"
	"time"
)

type Producer struct {
	MessageQueueTarget QueueTarget
	SendInterval       time.Duration
	buf                *bytes.Buffer
	mu                 *sync.Mutex
	done               chan error
}

func NewProducer(q QueueTarget) *Producer {
	return &Producer{
		MessageQueueTarget: q,
		buf:                new(bytes.Buffer),
		SendInterval:       1 * time.Second,
		mu:                 &sync.Mutex{},
		done:               make(chan error, 1),
	}
}

func (p *Producer) Write(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	n, err := p.buf.Write(b)
	if err != nil {
		p.done <- err
	}
	return n, err
}

func (p *Producer) Done() {
	p.done <- nil
}

func (p *Producer) ProduceStream() error {
	ticker := time.NewTicker(p.SendInterval)
	defer ticker.Stop()
	for {
		select {
		case <-p.done:
			p.Send()
			return nil
		case <-ticker.C:
			p.Send()
		}
	}
}

func (p *Producer) Send() {
	p.mu.Lock()
	if p.buf.Len() > 0 {
		err := p.MessageQueueTarget.Enqueue(p.buf.String())
		if err != nil {
			fmt.Println(err)
			p.done <- err
		}
		p.buf.Reset()
	}
}
