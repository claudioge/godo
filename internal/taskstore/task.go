package taskstore

import (
	"fmt"
	"time"
)

type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in-progress"
	StatusDone       TaskStatus = "done"
	StatusPaused     TaskStatus = "paused"
)

type Task struct {
	ID          int           `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description,omitempty"`
	Status      TaskStatus    `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	StartedAt   *time.Time    `json:"started_at,omitempty"`
	TotalTime   time.Duration `json:"total_time"`
	ParentID    *int          `json:"parent_id,omitempty"`
	SubtaskIDs  []int         `json:"subtask_ids,omitempty"`
}

// Maybe a struct to hold all tasks and metadata
type TaskStore struct {
	Tasks        []Task `json:"tasks"`
	NextID       int    `json:"next_id"`
	ActiveTaskID *int   `json:"active_task_id,omitempty"`
}

func GetTaskStatus(status string) (TaskStatus, error) {
	taskStatus := TaskStatus(status)
	switch taskStatus {
	case StatusTodo, StatusInProgress, StatusDone, StatusPaused:
		fmt.Println("status is valid")
		return taskStatus, nil
	default:
		return "", fmt.Errorf("invalid task status: %s", status)
	}
}
