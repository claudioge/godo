package modals

import (
	"godo/internal/taskstore"
	"godo/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type EditProjectGitLinkModal struct {
	form      *components.Form
	width     int
	height    int
	projectID int
}

func NewEditProjectGitLinkModal(project *taskstore.Project, width, height int) *EditProjectGitLinkModal {
	form := components.NewForm([]components.Field{
		{Label: "Name", Type: "text"},
		{Label: "Repository Path", Type: "text"},
		{Label: "Git Link", Type: "text"},
		{Label: "Branch", Type: "text"},
	})

	return &EditProjectGitLinkModal{
		form:      form,
		width:     width,
		height:    height,
		projectID: project.ID,
	}
}

func (m *EditProjectGitLinkModal) Init() tea.Cmd {
	return nil
}

func (m *EditProjectGitLinkModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
					ModalType: "edit_project_git_link",
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

func (m *EditProjectGitLinkModal) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Width(45).
		Height(12).
		Padding(1)

	return style.Render(m.form.View())
}
