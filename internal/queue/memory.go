package queue

import (
	"context"
	"errors"
	"log"
	"sync"

	"desktop-uploader/internal/domain"
)

type TaskChannelQueue struct {
	ch chan domain.Task
}

func NewTaskChannelQueue(bufferSize int) *TaskChannelQueue {
	return &TaskChannelQueue{
		ch: make(chan domain.Task, bufferSize),
	}
}

func (q *TaskChannelQueue) Push(task domain.Task) error {
	select {
	case q.ch <- task:
		return nil
	default:
		return errors.New("internal task queue is full")
	}
}

func (q *TaskChannelQueue) Close() {
	close(q.ch)
}

func (q *TaskChannelQueue) StartWorkers(ctx context.Context, count int, wg *sync.WaitGroup) {
	for i := 1; i <= count; i++ {
		workerID := i
		wg.Add(1)
		go q.worker(ctx, workerID, wg)
	}
}

func (q *TaskChannelQueue) worker(ctx context.Context, id int, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("[Worker %d] Active and awaiting tasks...", id)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Worker %d] Starting graceful shutdown...", id)
			return

		case task, ok := <-q.ch:
			if !ok {
				log.Printf("[Worker %d] Task channel closed. Exiting.", id)
				return
			}

			log.Printf("[Worker %d] Processing task %s for file: %s", id, task.ID, task.FilePath)

			log.Printf("[Worker %d] Task %s processed.", id, task.ID)
		}
	}
}
