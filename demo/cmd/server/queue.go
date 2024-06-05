package main

import (
	"sync"
)


type TaskQueue interface {
	Enqueue(task *TaskInstance) error
	Dequeue() (*TaskInstance, error)
}

type TaskQueueFactory func() TaskQueue

type MuxTaskQueue struct {
	tasks []*TaskInstance
	mux sync.Mutex
}

func NewMuxTaskQueueFactory(initialSize int) TaskQueueFactory {
	return func() TaskQueue {
		return NewMuxTaskQueue(initialSize)
	}
}

func NewMuxTaskQueue(initialSize int) *MuxTaskQueue {
	return &MuxTaskQueue{
		tasks: make([]*TaskInstance, 0, initialSize),
	}
}

func (q *MuxTaskQueue) Enqueue(task *TaskInstance) error {
	q.mux.Lock()
	defer q.mux.Unlock()

	q.tasks = append(q.tasks, task)
	return nil
}

func (q *MuxTaskQueue) Dequeue() (*TaskInstance, error) {
	q.mux.Lock()
	defer q.mux.Unlock()

	if len(q.tasks) == 0 {
		return nil, nil
	}

	task := q.tasks[0]
	q.tasks = q.tasks[1:]
	return task, nil
}	

