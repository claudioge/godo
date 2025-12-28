package modals

import (
	"godo/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AddTaskModal struct {
	form   *components.Form
	width  int
	height int
}

func NewAddTaskModal(width, height int) *AddTaskModal {
	form := components.NewForm([]components.Field{
		{Label: "Title", Type: "text"},
		{Label: "Description", Type: "textarea"},
	})

	return &AddTaskModal{
		form:   form,
		width:  width,
		height: height,
	}
}

func (m *AddTaskModal) Init() tea.Cmd {
	return nil
}

func (m *AddTaskModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
					ModalType: "add_task",
					Data:      m.form.GetValues(),
				}
			}
		default:
			m.form.Update(msg)
		}
	}
	return m, nil
}

func (m *AddTaskModal) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Width(45).
		Height(10).
		Padding(1)

	return style.Render(m.form.View())
}
