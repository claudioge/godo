package modals

import (
	"godo/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AddProjectModal struct {
	form   *components.Form
	width  int
	height int
}

func NewAddProjectModal(width, height int) *AddProjectModal {
	form := components.NewForm([]components.Field{
		{Label: "Name", Type: "text"},
		{Label: "Description", Type: "text"},
	})

	return &AddProjectModal{
		form:   form,
		width:  width,
		height: height,
	}
}

func (m *AddProjectModal) Init() tea.Cmd {
	return nil
}

func (m *AddProjectModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
					ModalType: "add_project",
					Data:      m.form.GetValues(),
				}
			}
		default:
			m.form.Update(msg)
		}
	}
	return m, nil
}

func (m *AddProjectModal) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Width(45).
		Height(10).
		Padding(1)

	return style.Render(m.form.View())
}
