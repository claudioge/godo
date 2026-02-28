package modals

import (
	"godo/internal/git"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SelectRepoModal struct {
	repos     []git.Repo
	cursor    int
	width     int
	height    int
	projectID int
}

func NewSelectRepoModal(projectID int, width, height int) *SelectRepoModal {
	repos, _ := git.ScanGitHubFolder()

	return &SelectRepoModal{
		repos:     repos,
		cursor:    0,
		width:     width,
		height:    height,
		projectID: projectID,
	}
}

func (m *SelectRepoModal) Init() tea.Cmd {
	return nil
}

func (m *SelectRepoModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "escape":
			return m, func() tea.Msg {
				return FormCancelledMsg{}
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.repos)+1 {
				m.cursor++
			}
		case "enter":
			if m.cursor == len(m.repos) {
				return m, func() tea.Msg {
					return FormCancelledMsg{}
				}
			}
			if m.cursor < len(m.repos) {
				repo := m.repos[m.cursor]
				return m, func() tea.Msg {
					return FormSubmittedMsg{
						ModalType: "edit_project_git_link",
						Data: map[string]string{
							"Name":            repo.Name,
							"Repository Path": repo.Path,
							"Git Link":        repo.URL,
							"Branch":          repo.Branch,
						},
						ProjectID: m.projectID,
					}
				}
			}
		}
	}
	return m, nil
}

func (m *SelectRepoModal) View() string {
	borderColor := "5"

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Width(60).
		Height(min(15, 7+len(m.repos))).
		Padding(1, 2)

	var lines []string

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(borderColor))
	lines = append(lines, headerStyle.Render("Select Git Repository"))

	subtitleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	lines = append(lines, subtitleStyle.Render("Choose a repo to link to this project"))
	lines = append(lines, "")

	if len(m.repos) == 0 {
		lines = append(lines, subtitleStyle.Render("No repositories found in ~/Documents/GitHub"))
	} else {
		for i, repo := range m.repos {
			isSelected := i == m.cursor
			if isSelected {
				lines = append(lines,
					lipgloss.NewStyle().
						Background(lipgloss.Color(borderColor)).
						Foreground(lipgloss.Color("0")).
						Render(" ▶ "+repo.Name+" ("+repo.Branch+")"),
				)
			} else {
				lines = append(lines, "  "+repo.Name+" ("+repo.Branch+")")
			}
		}
	}

	lines = append(lines, "")
	isManualSelected := len(m.repos) == m.cursor
	if isManualSelected {
		lines = append(lines,
			lipgloss.NewStyle().
				Background(lipgloss.Color(borderColor)).
				Foreground(lipgloss.Color("0")).
				Render(" ▶ [Manual entry]"),
		)
	} else {
		lines = append(lines, "  [Manual entry]")
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).MarginTop(1)
	lines = append(lines, helpStyle.Render("↑↓: navigate • Enter: select • Esc: cancel"))

	return style.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}
