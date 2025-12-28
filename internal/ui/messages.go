package ui

import "godo/internal/taskstore"

// Form events
type FormSubmittedMsg struct {
	ModalType string
	Data      map[string]string
}

type FormCancelledMsg struct{}

// Data updates
type TasksReloadedMsg struct {
	Tasks []taskstore.Task
}
