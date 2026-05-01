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
		{Label: "Title", Placeholder: "Enter task title..."},
		{Label: "Description", Type: "textarea", Placeholder: "Enter description (optional)..."},
	})
	form.SetViewWidth(46)

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
		if msg.String() == "esc" {
			return m, func() tea.Msg {
				return FormCancelledMsg{}
			}
		}
		submit, cancel := m.form.Update(msg)
		if cancel {
			return m, func() tea.Msg {
				return FormCancelledMsg{}
			}
		}
		if submit {
			return m, func() tea.Msg {
				return FormSubmittedMsg{
					ModalType: "add_task_to_project",
					Data:      m.form.GetValues(),
					ProjectID: m.projectID,
				}
			}
		}
	}
	return m, nil
}

func (m *AddTaskToProjectModal) View() string {
	formView := m.form.View()

	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("Tab: next field • ←→: move cursor • Enter: new line • Ctrl+Enter: save • Esc: cancel")

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("5")).
		Width(50).
		Padding(1, 2)

	content := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")).Render("Add Task to Project"),
		"",
		formView,
		"",
		helpText,
	)

	return style.Render(content)
}
