package modals

import (
	"godo/internal/taskstore"
	"godo/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Message types - defined here to avoid import cycles
type (
	FormSubmittedMsg struct {
		ModalType string
		Data      map[string]string
		TaskID    int
	}

	FormCancelledMsg struct{}
)

type EditGitLinkModal struct {
	form   *components.Form
	width  int
	height int
	taskID int
}

func NewEditGitLinkModal(task *taskstore.Task, width, height int) *EditGitLinkModal {
	form := components.NewForm([]components.Field{
		{Label: "Name", Type: "text"},
		{Label: "Repository Path", Type: "text"},
		{Label: "Git Link", Type: "text"},
		{Label: "Branch", Type: "text"},
	})

	return &EditGitLinkModal{
		form:   form,
		width:  width,
		height: height,
		taskID: task.ID,
	}
}

func (m *EditGitLinkModal) Init() tea.Cmd {
	return nil
}

func (m *EditGitLinkModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "escape":
			return m, func() tea.Msg {
				return FormCancelledMsg{}
			}
		case "enter":
			return m, func() tea.Msg {
				return FormSubmittedMsg{
					ModalType: "edit_git_link",
					Data:      m.form.GetValues(),
					TaskID:    m.taskID,
				}
			}
		default:
			m.form.Update(msg)
		}
	}
	return m, nil
}

func (m *EditGitLinkModal) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Width(45).
		Height(12).
		Padding(1)

	return style.Render(m.form.View())
}
