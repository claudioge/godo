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

type AIModal struct {
	task         *taskstore.Task
	project      *taskstore.Project
	gitLink      taskstore.GitLink
	aiClient     *ai.Client
	width        int
	height       int
	state        AIState
	response     string
	backupName   string
	worktreePath string
	branchName   string
	error        string
}

type AIState int

const (
	AIStateConfirm AIState = iota
	AIStateLoading
	AIStateResult
	AIStateError
)

func NewAIModal(task *taskstore.Task, project *taskstore.Project, gitLink taskstore.GitLink, width, height int) *AIModal {
	cfg, _ := config.Load()
	aiClient := ai.NewClient(&cfg.AI)

	branchName := "task-" + strings.ToLower(strings.ReplaceAll(task.Title, " ", "-"))

	return &AIModal{
		task:       task,
		project:    project,
		gitLink:    gitLink,
		aiClient:   aiClient,
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
		case "escape":
			return m, func() tea.Msg {
				return FormCancelledMsg{}
			}

		case "enter":
			if m.state == AIStateConfirm {
				return m, m.runAIImplementation
			}

		case "y":
			if m.state == AIStateConfirm {
				return m, m.runAIImplementation
			}

		case "u":
			if m.state == AIStateResult {
				return m, m.undoChanges
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

	worktreePath := filepath.Join(filepath.Dir(m.gitLink.LocalPath), m.branchName)

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
		m.error = fmt.Sprintf("AI Error: %v\n\nMake sure opencode is installed and configured.\nRun: opencode models", err)
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
		m.error = fmt.Sprintf("Could not parse AI response.\n\nThis usually means opencode needs configuration.\nResponse preview:\n%s", truncate(response, 500))
		return nil
	}

	var files []string
	for _, c := range changes {
		files = append(files, c.Path)
	}

	if err := git.CreateBackup(m.worktreePath, m.backupName, files); err != nil {
		m.error = fmt.Sprintf("Warning: failed to create backup: %v", err)
	}

	if err := git.ApplyChanges(m.worktreePath, changes); err != nil {
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
		content = append(content, repoStyle.Render("Worktree: "+filepath.Join(filepath.Dir(m.gitLink.LocalPath), m.branchName)))
		content = append(content, "")

		warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
		content = append(content, warnStyle.Render("This will create a new branch and worktree."))
		content = append(content, warnStyle.Render("Use 'u' to undo changes after."))
		content = append(content, "")

		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		content = append(content, helpStyle.Render("Press Enter or 'y' to start • Esc to cancel"))

	case AIStateLoading:
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(borderColor))
		content = append(content, headerStyle.Render("Generating code..."))
		content = append(content, "")

		spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		loadingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
		content = append(content, loadingStyle.Render(spinner[len(spinner)/2%len(spinner)]+" Connecting to AI and generating code..."))

	case AIStateResult:
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
		content = append(content, headerStyle.Render("✓ Implementation Complete"))
		content = append(content, "")

		successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
		content = append(content, successStyle.Render("Worktree created at: "+m.worktreePath))
		content = append(content, successStyle.Render("Branch: "+m.branchName))
		content = append(content, "")

		if m.response != "" && !strings.Contains(m.response, "NO_CHANGES_NEEDED") {
			content = append(content, "Generated code:")
			content = append(content, "")

			respStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
			respLines := strings.Split(m.response, "\n")
			for i, line := range respLines {
				if i > 15 {
					content = append(content, respStyle.Render("... (truncated)"))
					break
				}
				content = append(content, respStyle.Render(line))
			}
			content = append(content, "")
		}

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
