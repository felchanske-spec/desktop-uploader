package usecase

import (
	"desktop-uploader/internal/domain"
	"errors"
	"fmt"
	"os"
	"time"
)

var (
	ErrFileNotFound        = errors.New("File not found on disk")
	ErrFileTooLarge        = errors.New("File exceeds the 50MB limit")
	ErrCannotSendDirectory = errors.New("Cannot send a directory, please select a file")
)

type TaskQueue interface {
	Push(task domain.Task) error
}

type UploadInteractor struct {
	queue TaskQueue
}

func NewUploadInteractor(q TaskQueue) *UploadInteractor {
	return &UploadInteractor{queue: q}
}

func (u *UploadInteractor) ValidateAndQueue(filePath, target string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return fmt.Errorf("file access error: %w", err)
	}

	if info.IsDir() {
		return ErrCannotSendDirectory
	}

	maxSize := int64(50 * 1024 * 1024)
	size := info.Size()

	fmt.Printf("[Validator] Checking file size: %d MB\n", size/1024/1024)

	if size > maxSize {
		return ErrFileTooLarge
	}

	task := domain.Task{
		ID:        fmt.Sprintf("task_%d", time.Now().UnixNano()),
		FilePath:  filePath,
		Target:    target,
		CreatedAt: time.Now(),
	}

	err = u.queue.Push(task)
	if err != nil {
		return fmt.Errorf("Failed to queue the task: %w", err)
	}
	return nil
}
