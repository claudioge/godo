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

// Ensure compatibility with tea.Model
var _ tea.Msg = FormSubmittedMsg{}
var _ tea.Msg = FormCancelledMsg{}
