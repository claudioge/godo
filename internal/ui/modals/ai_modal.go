package modals

import (
	"fmt"
	"path/filepath"
	"strings"

	"godo/internal/ai"
	"godo/internal/config"
	"godo/internal/git"
	"godo/internal/taskstore"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AIState int

const (
	AIStateConfirm AIState = iota
	AIStateLoading
	AIStateReview
	AIStateResult
	AIStateError
)

type AIModal struct {
	task         *taskstore.Task
	project      *taskstore.Project
	gitLink      taskstore.GitLink
	aiClient     *ai.Client
	cfg          *config.Config
	width        int
	height       int
	state        AIState
	response     string
	changes      []git.FileChange
	backupName   string
	worktreePath string
	branchName   string
	error        string
}

func NewAIModal(task *taskstore.Task, project *taskstore.Project, gitLink taskstore.GitLink, width, height int) *AIModal {
	cfg, _ := config.Load()
	aiClient := ai.NewClient(&cfg.AI)

	branchName := "task-" + strings.ToLower(strings.ReplaceAll(task.Title, " ", "-"))

	return &AIModal{
		task:       task,
		project:    project,
		gitLink:    gitLink,
		aiClient:   aiClient,
		cfg:        cfg,
		width:      width,
		height:     height,
		state:      AIStateConfirm,
		branchName: branchName,
	}
}

func (m *AIModal) Init() tea.Cmd {
	return nil
}

func (m *AIModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, func() tea.Msg {
				return FormCancelledMsg{}
			}

		case "enter", "y":
			if m.state == AIStateConfirm {
				worktreePath := ""
				if m.cfg.Worktree.BasePath != "" {
					worktreePath = filepath.Join(m.cfg.Worktree.BasePath, m.project.Name, m.branchName)
				} else {
					worktreePath = filepath.Join(filepath.Dir(m.gitLink.LocalPath), m.branchName)
				}

				prompt := ai.BuildPrompt(m.task.Title, m.task.Description, m.gitLink.LocalPath)

				return m, func() tea.Msg {
					return StartAIJobMsg{
						TaskID:   m.task.ID,
						Prompt:   prompt,
						Worktree: worktreePath,
						Branch:   m.branchName,
					}
				}
			}

		case "c":
			if m.state == AIStateResult {
				return m, func() tea.Msg {
					return FormCancelledMsg{}
				}
			}
		}
	}
	return m, nil
}

func (m *AIModal) runAIImplementation() tea.Msg {
	m.state = AIStateLoading

	worktreePath := ""
	if m.cfg.Worktree.BasePath != "" {
		worktreePath = filepath.Join(m.cfg.Worktree.BasePath, m.project.Name, m.branchName)
	} else {
		worktreePath = filepath.Join(filepath.Dir(m.gitLink.LocalPath), m.branchName)
	}

	if err := git.CreateWorktree(m.gitLink.LocalPath, m.branchName, worktreePath); err != nil {
		m.state = AIStateError
		m.error = fmt.Sprintf("Failed to create worktree: %v", err)
		return nil
	}

	m.worktreePath = worktreePath
	m.aiClient.SetPaths(m.gitLink.LocalPath, worktreePath)

	prompt := ai.BuildPrompt(m.task.Title, m.task.Description, m.gitLink.LocalPath)
	m.backupName = fmt.Sprintf("%s-%d", m.branchName, m.task.ID)

	response, err := m.aiClient.Generate(prompt)
	if err != nil {
		m.state = AIStateError
		m.error = fmt.Sprintf("AI Error: %v", err)
		return nil
	}

	m.response = response

	if strings.Contains(response, "NO_CHANGES_NEEDED") || strings.Contains(response, "no changes needed") {
		m.state = AIStateResult
		return nil
	}

	changes, err := git.ParseAIResponse(response)
	if err != nil {
		m.state = AIStateError
		m.error = fmt.Sprintf("Could not parse AI response: %v", err)
		return nil
	}

	m.changes = changes
	m.state = AIStateReview
	return nil
}

func (m *AIModal) applyChanges() tea.Msg {
	var files []string
	for _, c := range m.changes {
		files = append(files, c.Path)
	}

	if err := git.CreateBackup(m.worktreePath, m.backupName, files); err != nil {
		m.error = fmt.Sprintf("Warning: failed to create backup: %v", err)
	}

	if err := git.ApplyChanges(m.worktreePath, m.changes); err != nil {
		m.state = AIStateError
		m.error = fmt.Sprintf("Failed to apply changes: %v", err)
		return nil
	}

	m.state = AIStateResult
	return nil
}

func (m *AIModal) undoChanges() tea.Msg {
	if m.backupName != "" {
		if err := git.RestoreBackup(m.worktreePath, m.backupName); err != nil {
			m.error = fmt.Sprintf("Undo failed: %v", err)
			return nil
		}
	}

	if m.branchName != "" {
		git.DeleteWorktree(m.gitLink.LocalPath, m.worktreePath)
		git.DeleteBranch(m.gitLink.LocalPath, m.branchName)
	}

	return FormCancelledMsg{}
}

func (m *AIModal) View() string {
	borderColor := "5"

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Width(min(70, m.width-10)).
		Padding(1, 2)

	var content []string

	switch m.state {
	case AIStateConfirm:
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(borderColor))
		content = append(content, headerStyle.Render("AI Implementation"))
		content = append(content, "")

		taskStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
		content = append(content, taskStyle.Render("Task: "+m.task.Title))
		content = append(content, "")

		repoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		content = append(content, repoStyle.Render("Repository: "+m.gitLink.Name))

		worktreePath := ""
		if m.cfg.Worktree.BasePath != "" {
			worktreePath = filepath.Join(m.cfg.Worktree.BasePath, m.project.Name, m.branchName)
		} else {
			worktreePath = filepath.Join(filepath.Dir(m.gitLink.LocalPath), m.branchName)
		}
		content = append(content, repoStyle.Render("Worktree: "+worktreePath))
		content = append(content, "")

		warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
		content = append(content, warnStyle.Render("This will create a new branch and worktree."))
		content = append(content, "")

		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		content = append(content, helpStyle.Render("Press Enter or 'y' to start • Esc to cancel"))

	case AIStateLoading:
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(borderColor))
		content = append(content, headerStyle.Render("Generating code..."))
		content = append(content, "")

		spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		loadingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
		content = append(content, loadingStyle.Render(spinner[0]+" Connecting to AI and generating code..."))

	case AIStateReview:
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
		content = append(content, headerStyle.Render("Review Changes"))
		content = append(content, "")

		content = append(content, "The AI wants to modify/create the following files:")
		content = append(content, "")

		fileStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
		for i, c := range m.changes {
			if i > 10 {
				content = append(content, "  ... and more")
				break
			}
			content = append(content, "  • "+fileStyle.Render(c.Path))
		}
		content = append(content, "")

		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		content = append(content, helpStyle.Render("'a' or Enter: apply • 'c' or Esc: cancel"))

	case AIStateResult:
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
		content = append(content, headerStyle.Render("✓ Implementation Complete"))
		content = append(content, "")

		successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
		content = append(content, successStyle.Render("Worktree created at: "+m.worktreePath))
		content = append(content, successStyle.Render("Branch: "+m.branchName))
		content = append(content, "")

		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		content = append(content, helpStyle.Render("'u': undo changes • 'c': close"))

	case AIStateError:
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1"))
		content = append(content, headerStyle.Render("✗ Error"))
		content = append(content, "")

		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
		content = append(content, errStyle.Render(m.error))
		content = append(content, "")

		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		content = append(content, helpStyle.Render("Press Esc to close"))
	}

	return style.Render(lipgloss.JoinVertical(lipgloss.Left, content...))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
