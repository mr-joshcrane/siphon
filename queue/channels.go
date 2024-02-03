package queue

type ChannelQueue struct {
	queue chan string
}

func NewChannelQueue(size int) *ChannelQueue {
	c := make(chan string, size)
	return &ChannelQueue{
		queue: c,
	}
}

func (q *ChannelQueue) Enqueue(items ...string) error {
	for _, msg := range items {
		q.queue <- msg
	}
	return nil
}
func (q *ChannelQueue) Dequeue() (string, error) {
	return <-q.queue, nil

}
func (q *ChannelQueue) Size() int {
	return len(q.queue)
}
