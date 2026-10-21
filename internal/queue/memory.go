package queue

import (
	"context"
	"errors"
	"log"
	"strconv"
	"sync"

	"desktop-uploader/internal/domain"
)

var ErrQueueFull = errors.New("internal task queue is full")

type TelegramClient interface {
	SendFile(chatID int64, filePath string) error
}

type TaskChannelQueue struct {
	ch       chan domain.Task
	tgClient TelegramClient
}

func NewTaskChannelQueue(bufferSize int, tgClient TelegramClient) *TaskChannelQueue {
	return &TaskChannelQueue{
		ch:       make(chan domain.Task, bufferSize),
		tgClient: tgClient,
	}
}

func (q *TaskChannelQueue) Push(task domain.Task) error {
	select {
	case q.ch <- task:
		return nil
	default:
		return ErrQueueFull
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

			chatID, err := strconv.ParseInt(task.Target, 10, 64)
			if err != nil {
				log.Printf("[Worker %d] Error parsing chat ID: %v", id, err)
				continue
			}

			err = q.tgClient.SendFile(chatID, task.FilePath)
			if err != nil {
				log.Printf("[Worker %d] Failed to send file %s: %v", id, task.ID, err)
				continue
			}

			log.Printf("[Worker %d] Task %s processed.", id, task.ID)
		}
	}
}
