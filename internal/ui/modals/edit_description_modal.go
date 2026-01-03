package modals

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type EditDescriptionModal struct {
	taskID      int
	description string
	lines       []string
	cursorLine  int
	cursorCol   int
	width       int
	height      int
	viewOffset  int
}

func NewEditDescriptionModal(taskID int, description string, width, height int) *EditDescriptionModal {
	lines := strings.Split(description, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}

	return &EditDescriptionModal{
		taskID:      taskID,
		description: description,
		lines:       lines,
		cursorLine:  0,
		cursorCol:   len(lines[0]),
		width:       width,
		height:      height,
		viewOffset:  0,
	}
}

func (m *EditDescriptionModal) Init() tea.Cmd {
	return nil
}

func (m *EditDescriptionModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "escape":
			return m, func() tea.Msg {
				return FormCancelledMsg{}
			}

		case "ctrl+s":
			return m, func() tea.Msg {
				return FormSubmittedMsg{
					ModalType: "edit_description",
					Data: map[string]string{
						"Description": strings.Join(m.lines, "\n"),
					},
					TaskID: m.taskID,
				}
			}

		case "enter":
			if m.cursorLine < len(m.lines) {
				line := m.lines[m.cursorLine]
				before := line[:m.cursorCol]
				after := line[m.cursorCol:]

				m.lines[m.cursorLine] = before
				m.lines = append(m.lines[:m.cursorLine+1], append([]string{after}, m.lines[m.cursorLine+1:]...)...)

				m.cursorLine++
				m.cursorCol = 0
			}

		case "backspace":
			if m.cursorCol > 0 {
				line := m.lines[m.cursorLine]
				m.lines[m.cursorLine] = line[:m.cursorCol-1] + line[m.cursorCol:]
				m.cursorCol--
			} else if m.cursorLine > 0 {
				prevLine := m.lines[m.cursorLine-1]
				m.cursorCol = len(prevLine)
				m.lines[m.cursorLine-1] = prevLine + m.lines[m.cursorLine]
				m.lines = append(m.lines[:m.cursorLine], m.lines[m.cursorLine+1:]...)
				m.cursorLine--
			}

		case "delete":
			if m.cursorCol < len(m.lines[m.cursorLine]) {
				line := m.lines[m.cursorLine]
				m.lines[m.cursorLine] = line[:m.cursorCol] + line[m.cursorCol+1:]
			} else if m.cursorLine < len(m.lines)-1 {
				m.lines[m.cursorLine] = m.lines[m.cursorLine] + m.lines[m.cursorLine+1]
				m.lines = append(m.lines[:m.cursorLine+1], m.lines[m.cursorLine+2:]...)
			}

		case "up":
			if m.cursorLine > 0 {
				m.cursorLine--
				if m.cursorCol > len(m.lines[m.cursorLine]) {
					m.cursorCol = len(m.lines[m.cursorLine])
				}
			}

		case "down":
			if m.cursorLine < len(m.lines)-1 {
				m.cursorLine++
				if m.cursorCol > len(m.lines[m.cursorLine]) {
					m.cursorCol = len(m.lines[m.cursorLine])
				}
			}

		case "left":
			if m.cursorCol > 0 {
				m.cursorCol--
			} else if m.cursorLine > 0 {
				m.cursorLine--
				m.cursorCol = len(m.lines[m.cursorLine])
			}

		case "right":
			if m.cursorCol < len(m.lines[m.cursorLine]) {
				m.cursorCol++
			} else if m.cursorLine < len(m.lines)-1 {
				m.cursorLine++
				m.cursorCol = 0
			}

		case "home":
			m.cursorCol = 0

		case "end":
			m.cursorCol = len(m.lines[m.cursorLine])

		case "ctrl+home":
			m.cursorLine = 0
			m.cursorCol = 0

		case "ctrl+end":
			m.cursorLine = len(m.lines) - 1
			m.cursorCol = len(m.lines[m.cursorLine])

		default:
			if msg.Type == tea.KeyRunes {
				line := m.lines[m.cursorLine]
				m.lines[m.cursorLine] = line[:m.cursorCol] + string(msg.Runes) + line[m.cursorCol:]
				m.cursorCol++
			} else if msg.Type == tea.KeySpace {
				line := m.lines[m.cursorLine]
				m.lines[m.cursorLine] = line[:m.cursorCol] + " " + line[m.cursorCol:]
				m.cursorCol++
			}
		}

		viewHeight := m.height - 8
		if m.cursorLine < m.viewOffset {
			m.viewOffset = m.cursorLine
		} else if m.cursorLine >= m.viewOffset+viewHeight {
			m.viewOffset = m.cursorLine - viewHeight + 1
		}
	}

	return m, nil
}

func (m *EditDescriptionModal) View() string {
	contentWidth := m.width - 6
	viewHeight := m.height - 8

	var editorLines []string
	maxLines := len(m.lines)
	if maxLines > viewHeight {
		maxLines = viewHeight
	}

	for i := 0; i < maxLines; i++ {
		lineNum := m.viewOffset + i
		if lineNum < len(m.lines) {
			line := m.lines[lineNum]
			if lineNum == m.cursorLine {
				if m.cursorCol <= len(line) {
					displayLine := line[:m.cursorCol] + "│" + line[m.cursorCol:]
					if len(displayLine) > contentWidth {
						displayLine = displayLine[:contentWidth]
					}
					editorLines = append(editorLines, displayLine)
				} else {
					editorLines = append(editorLines, line+"│")
				}
			} else {
				if len(line) > contentWidth {
					editorLines = append(editorLines, line[:contentWidth])
				} else {
					editorLines = append(editorLines, line)
				}
			}
		}
	}

	for i := len(editorLines); i < viewHeight; i++ {
		editorLines = append(editorLines, "")
	}

	editorContent := strings.Join(editorLines, "\n")

	editorStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Width(m.width - 4).
		Height(viewHeight).
		Padding(1)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("4")).
		Margin(0, 0, 1, 0).
		Render("✏️ Edit Description")

	helpStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color("8")).
		Margin(1, 0, 0, 0)

	help := helpStyle.Render("Ctrl+S: save • ESC: cancel • Arrow keys: move • Enter: new line")

	statsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true).
		Margin(1, 0, 0, 0)

	stats := statsStyle.Render(fmt.Sprintf("Line %d/%d • Col %d", m.cursorLine+1, len(m.lines), m.cursorCol))

	return title + "\n" + editorStyle.Render(editorContent) + "\n" + help + "\n" + stats
}
