package ui

import (
	"fmt"

	"godo/internal/taskstore"

	tea "github.com/charmbracelet/bubbletea"
)

// RunInteractiveList loads tasks and runs the interactive TUI app
func RunInteractiveList() error {
	tasks, err := taskstore.GetTasks()
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	app := NewApp(tasks)
	p := tea.NewProgram(app, tea.WithAltScreen())

	_, err = p.Run()
	return err
}
