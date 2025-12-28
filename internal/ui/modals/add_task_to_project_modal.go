package modals

import (
	"godo/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AddTaskToProjectModal struct {
	form      *components.Form
	width     int
	height    int
	projectID int
}

func NewAddTaskToProjectModal(projectID int, width, height int) *AddTaskToProjectModal {
	form := components.NewForm([]components.Field{
		{Label: "Title", Type: "text"},
		{Label: "Description", Type: "textarea"},
	})

	return &AddTaskToProjectModal{
		form:      form,
		width:     width,
		height:    height,
		projectID: projectID,
	}
}

func (m *AddTaskToProjectModal) Init() tea.Cmd {
	return nil
}

func (m *AddTaskToProjectModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
					ModalType: "add_task_to_project",
					Data:      m.form.GetValues(),
					ProjectID: m.projectID,
				}
			}
		default:
			m.form.Update(msg)
		}
	}
	return m, nil
}

func (m *AddTaskToProjectModal) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Width(45).
		Height(10).
		Padding(1)

	return style.Render(m.form.View())
}
