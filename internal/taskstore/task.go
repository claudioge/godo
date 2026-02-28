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

type Project struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	GitLinks    []GitLink `json:"git_links,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

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
	ProjectIDs  []int         `json:"project_ids,omitempty"`
	GitLinks    []GitLink     `json:"git_links,omitempty"`
}

type GitLink struct {
	LocalPath string `json:"local_path"`
	Name      string `json:"name"`
	Link      string `json:"link"`
	Branch    string `json:"branch"`
}

type TaskStore struct {
	Tasks         []Task    `json:"tasks"`
	Projects      []Project `json:"projects"`
	NextID        int       `json:"next_id"`
	NextProjectID int       `json:"next_project_id"`
	ActiveTaskID  *int      `json:"active_task_id,omitempty"`
}

func GetTaskStatus(status string) (TaskStatus, error) {
	taskStatus := TaskStatus(status)
	switch taskStatus {
	case StatusTodo, StatusInProgress, StatusDone, StatusPaused:
		return taskStatus, nil
	default:
		return "", fmt.Errorf("invalid task status: %s", status)
	}
}
