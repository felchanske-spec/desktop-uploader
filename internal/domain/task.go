package domain

import "time"

type TaskState string

const (
	StateQueued     TaskState = "QUEUED"
	StateValidating TaskState = "VALIDATING"
	StateSending    TaskState = "SENDING"
	StateSuccess    TaskState = "SUCCESS"
	StateFailed     TaskState = "FAILED"
)

type Task struct {
	ID        string
	FilePath  string
	Target    string
	CreatedAt time.Time
}

type TaskStatus struct {
	TaskID          string
	State           TaskState
	Message         string
	ProgressPercent int32
}
