package modals

import (
	"godo/internal/taskstore"
	"godo/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type EditTitleModal struct {
	form   *components.Form
	width  int
	height int
	task   *taskstore.Task
}

func NewEditTitleModal(task *taskstore.Task, width, height int) *EditTitleModal {
	form := components.NewForm([]components.Field{
		{Label: "Title", Type: "text", Value: task.Title},
	})

	return &EditTitleModal{
		form:   form,
		width:  width,
		height: height,
		task:   task,
	}
}

func (m *EditTitleModal) Init() tea.Cmd {
	return nil
}

func (m *EditTitleModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
					ModalType: "edit_title",
					Data:      m.form.GetValues(),
				}
			}
		default:
			m.form.Update(msg)
		}
	}
	return m, nil
}

func (m *EditTitleModal) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Width(45).
		Height(8).
		Padding(1)

	return style.Render(m.form.View())
}
