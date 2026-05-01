package modals

import tea "github.com/charmbracelet/bubbletea"

// Message types used by modals
type FormSubmittedMsg struct {
	ModalType string
	Data      map[string]string
	TaskID    int
	ProjectID int
}

type FormCancelledMsg struct{}

type StartAIJobMsg struct {
	TaskID   int
	Prompt   string
	Worktree string
	Branch   string
}

type AIJobStatusMsg struct {
	TaskID  int
	Status  string // "starting", "running", "done", "error"
	Message string
	Error   error
}

// Ensure compatibility with tea.Model
var _ tea.Msg = FormSubmittedMsg{}
var _ tea.Msg = FormCancelledMsg{}
var _ tea.Msg = StartAIJobMsg{}
var _ tea.Msg = AIJobStatusMsg{}
