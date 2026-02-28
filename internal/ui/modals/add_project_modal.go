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
		{Label: "Name", Placeholder: "Enter project name..."},
		{Label: "Description", Type: "textarea", Placeholder: "Enter description (optional)..."},
	})
	form.SetViewWidth(46)

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
		submit, cancel := m.form.Update(msg)
		if cancel {
			return m, func() tea.Msg {
				return FormCancelledMsg{}
			}
		}
		if submit {
			return m, func() tea.Msg {
				return FormSubmittedMsg{
					ModalType: "add_project",
					Data:      m.form.GetValues(),
				}
			}
		}
	}
	return m, nil
}

func (m *AddProjectModal) View() string {
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
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")).Render("Add New Project"),
		"",
		formView,
		"",
		helpText,
	)

	return style.Render(content)
}
