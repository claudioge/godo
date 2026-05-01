package modals

import (
	"fmt"
	"godo/internal/taskstore"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SelectProjectModal struct {
	projects []taskstore.Project
	cursor   int
	width    int
	height   int
	taskID   int
}

func NewSelectProjectModal(taskID int, projects []taskstore.Project, width, height int) *SelectProjectModal {
	return &SelectProjectModal{
		projects: projects,
		cursor:   0,
		width:    width,
		height:   height,
		taskID:   taskID,
	}
}

func (m *SelectProjectModal) Init() tea.Cmd {
	return nil
}

func (m *SelectProjectModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg {
				return FormCancelledMsg{}
			}
		case "j", "down":
			if m.cursor < len(m.projects)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "enter":
			if m.cursor >= 0 && m.cursor < len(m.projects) {
				return m, func() tea.Msg {
					return FormSubmittedMsg{
						ModalType: "assign_task_to_project",
						Data:      map[string]string{},
						TaskID:    m.taskID,
						ProjectID: m.projects[m.cursor].ID,
					}
				}
			}
		}
	}
	return m, nil
}

func (m *SelectProjectModal) View() string {
	if len(m.projects) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("4")).
			Width(50).
			Padding(1)
		return emptyStyle.Render("No projects available\nPress ESC to cancel")
	}

	var items []string
	for i, project := range m.projects {
		isSelected := i == m.cursor
		text := fmt.Sprintf("  %s", project.Name)

		style := lipgloss.NewStyle()
		if isSelected {
			style = style.
				Background(lipgloss.Color("4")).
				Foreground(lipgloss.Color("0")).
				Bold(true)
			text = "➡️ " + project.Name
		}

		items = append(items, style.Render(text))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, items...)

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Width(50).
		Height(m.height - 4).
		Padding(1)

	helpStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color("8"))

	return style.Render(content) + "\n" + helpStyle.Render("j/k or ↓/↑: navigate • ENTER: select • ESC: cancel")
}
