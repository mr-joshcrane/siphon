package siphon

type Worker struct {
	Queue Queue
}

func NewWorker(q Queue) *Worker {
	return &Worker{
		Queue: q,
	}
}
