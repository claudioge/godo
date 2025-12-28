package modals

import (
	"godo/internal/taskstore"
	"godo/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type EditProjectModal struct {
	form      *components.Form
	width     int
	height    int
	projectID int
}

func NewEditProjectModal(project *taskstore.Project, width, height int) *EditProjectModal {
	form := components.NewForm([]components.Field{
		{Label: "Name", Type: "text"},
		{Label: "Description", Type: "text"},
	})

	// Pre-fill the form with existing data
	form.SetValue("Name", project.Name)
	form.SetValue("Description", project.Description)

	return &EditProjectModal{
		form:      form,
		width:     width,
		height:    height,
		projectID: project.ID,
	}
}

func (m *EditProjectModal) Init() tea.Cmd {
	return nil
}

func (m *EditProjectModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
					ModalType: "edit_project",
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

func (m *EditProjectModal) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Width(45).
		Height(10).
		Padding(1)

	return style.Render(m.form.View())
}
