package modals

import tea "github.com/charmbracelet/bubbletea"

// All modals implement this
type Modal interface {
	tea.Model
	View() string
}
