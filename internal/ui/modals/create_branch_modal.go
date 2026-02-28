package modals

import (
	"godo/internal/taskstore"
	"godo/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CreateBranchModal struct {
	form    *components.Form
	width   int
	height  int
	task    *taskstore.Task
	project *taskstore.Project
	gitLink taskstore.GitLink
}

func NewCreateBranchModal(task *taskstore.Task, project *taskstore.Project, gitLink taskstore.GitLink, width, height int) *CreateBranchModal {
	defaultBranch := "task-" + formatBranchName(task.Title)

	form := components.NewForm([]components.Field{
		{Label: "Branch Name", Value: defaultBranch},
	})

	return &CreateBranchModal{
		form:    form,
		width:   width,
		height:  height,
		task:    task,
		project: project,
		gitLink: gitLink,
	}
}

func (m *CreateBranchModal) Init() tea.Cmd {
	return nil
}

func (m *CreateBranchModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		submit, cancel := m.form.Update(msg)
		if cancel {
			return m, func() tea.Msg {
				return FormCancelledMsg{}
			}
		}
		if submit {
			values := m.form.GetValues()
			branchName := values["Branch Name"]
			if branchName == "" {
				return m, nil
			}
			return m, func() tea.Msg {
				return BranchCreatedMsg{
					TaskID:     m.task.ID,
					ProjectID:  m.project.ID,
					GitLink:    m.gitLink,
					BranchName: branchName,
				}
			}
		}
	}
	return m, nil
}

func (m *CreateBranchModal) View() string {
	formView := m.form.View()

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	repoInfo := infoStyle.Render("Repository: " + m.gitLink.Name)
	pathInfo := infoStyle.Render("Path: " + m.gitLink.LocalPath)

	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("←→: move cursor • Enter: create branch • Esc: cancel")

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("5")).
		Width(55).
		Padding(1, 2)

	content := lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Render("Create Branch for Task"),
		"",
		infoStyle.Render("Task: "+m.task.Title),
		repoInfo,
		pathInfo,
		"",
		formView,
		"",
		helpText,
	)

	return style.Render(content)
}

type BranchCreatedMsg struct {
	TaskID     int
	ProjectID  int
	GitLink    taskstore.GitLink
	BranchName string
}

func formatBranchName(name string) string {
	var result []rune
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result = append(result, r)
		} else if r == ' ' {
			result = append(result, '-')
		}
	}
	return string(result)
}
