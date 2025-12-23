package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"godo/internal/taskstore"
)

type Mode int

const (
	ModeNormal Mode = iota
	ModeEdit
	ModeEditDescription
	ModeAddTask
)

type InteractiveList struct {
	tasks              []taskstore.Task
	cursor             int
	mode               Mode
	editBuffer         string
	message            string
	messageTimer       time.Time
	height             int
	width              int
	selectedTask       *taskstore.Task
	newTaskTitle       string
	newTaskDescription string
}

type tasksReloadedMsg struct {
	tasks []taskstore.Task
}

type taskStatusChangedMsg struct {
	id     int
	status taskstore.TaskStatus
}

type editCompleteMsg struct {
	id    int
	field string
	value string
}

type taskDeletedMsg struct {
	id int
}

func NewInteractiveList(tasks []taskstore.Task) *InteractiveList {
	return &InteractiveList{
		tasks:  tasks,
		cursor: 0,
		mode:   ModeNormal,
		height: 20,
		width:  80,
	}
}

func (m *InteractiveList) Init() tea.Cmd {
	return nil
}

func (m *InteractiveList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case taskStatusChangedMsg:
		m.setMessage(fmt.Sprintf("Task #%d status changed to %s", msg.id, msg.status))
		return m, m.reloadTasks()

	case editCompleteMsg:
		m.mode = ModeNormal
		m.editBuffer = ""
		m.setMessage(fmt.Sprintf("Task #%d %s updated", msg.id, msg.field))
		return m, m.reloadTasks()

	case tasksReloadedMsg:
		m.tasks = msg.tasks
		// Adjust cursor if it's out of bounds
		totalDisplayTasks := m.countDisplayTasks()
		if m.cursor >= totalDisplayTasks && totalDisplayTasks > 0 {
			m.cursor = totalDisplayTasks - 1
		} else if totalDisplayTasks == 0 {
			m.cursor = 0
		}
		return m, nil

	case taskDeletedMsg:
		m.setMessage(fmt.Sprintf("Task #%d deleted", msg.id))
		return m, m.reloadTasks()
	}

	return m, nil
}

func (m *InteractiveList) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeNormal:
		return m.handleNormalMode(msg)
	case ModeEdit:
		return m.handleEditMode(msg)
	case ModeEditDescription:
		return m.handleEditDescriptionMode(msg)
	case ModeAddTask:
		return m.handleAddTaskMode(msg)
	default:
		return m, nil
	}
}

func (m *InteractiveList) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "a":
		// Enter add task mode
		m.mode = ModeAddTask
		m.editBuffer = ""
		m.newTaskTitle = ""
		m.newTaskDescription = ""
		return m, nil

	// Vim-like navigation
	case "j", "down":
		maxTasks := m.countDisplayTasks()
		if m.cursor < maxTasks-1 {
			m.cursor++
		}

	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}

	case "g":
		m.cursor = 0

	case "G":
		maxTasks := m.countDisplayTasks()
		if maxTasks > 0 {
			m.cursor = maxTasks - 1
		}

	// Status changes
	case "t":
		return m, m.changeTaskStatus(taskstore.StatusTodo)

	case "p":
		return m, m.changeTaskStatus(taskstore.StatusInProgress)

	case "d":
		return m, m.changeTaskStatus(taskstore.StatusDone)

	case "s":
		return m, m.changeTaskStatus(taskstore.StatusPaused)

	// Edit operations
	case "e", "i":
		m.mode = ModeEdit
		if task := m.getCurrentTask(); task != nil {
			m.editBuffer = task.Title
		}

	case "o":
		m.mode = ModeEditDescription
		if task := m.getCurrentTask(); task != nil {
			m.editBuffer = task.Description
		}

	// Delete task
	case "x", "delete":
		return m, m.deleteCurrentTask()

	// Help
	case "?", "h":
		m.setMessage("j/k: move, t/p/d/s: status, e: edit title, o: edit desc (Ctrl+S to save), x: delete, q: quit")
	}

	return m, nil
}

func (m *InteractiveList) handleEditMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "escape":
		m.mode = ModeNormal
		m.editBuffer = ""

	case "enter":
		if task := m.getCurrentTask(); task != nil && m.editBuffer != "" {
			cmd := m.updateTaskTitle(task.ID, m.editBuffer)
			return m, cmd
		}
		m.mode = ModeNormal
		m.editBuffer = ""

	case "backspace":
		if len(m.editBuffer) > 0 {
			m.editBuffer = m.editBuffer[:len(m.editBuffer)-1]
		}

	default:
		// Handle all printable characters including spaces
		switch msg.Type {
		case tea.KeyRunes:
			m.editBuffer += string(msg.Runes)
		case tea.KeySpace:
			m.editBuffer += " "
		}
	}

	return m, nil
}

func (m *InteractiveList) handleEditDescriptionMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "escape":
		m.mode = ModeNormal
		m.editBuffer = ""

	case "enter":
		// For description, enter adds a newline
		m.editBuffer += "\n"

	case "ctrl+s":
		// Ctrl+S to save description
		if task := m.getCurrentTask(); task != nil {
			cmd := m.updateTaskDescription(task.ID, m.editBuffer)
			return m, cmd
		}
		m.mode = ModeNormal
		m.editBuffer = ""

	case "backspace":
		if len(m.editBuffer) > 0 {
			m.editBuffer = m.editBuffer[:len(m.editBuffer)-1]
		}

	case "ctrl+u":
		// Clear entire line
		m.editBuffer = ""

	default:
		// Handle all printable characters including spaces
		switch msg.Type {
		case tea.KeyRunes:
			m.editBuffer += string(msg.Runes)
		case tea.KeySpace:
			m.editBuffer += " "
		}
	}

	return m, nil
}

func (m *InteractiveList) View() string {
	// Handle add task mode separately
	if m.mode == ModeAddTask {
		return m.renderAddTaskMode()
	}

	if len(m.tasks) == 0 {
		return m.renderEmpty()
	}

	var sections []string

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("4")).
		Render("📋 Interactive Task List")
	sections = append(sections, header)

	// Create ordered task list matching static list ordering
	orderedTasks := []taskstore.Task{}

	// Status sections in order
	statusSections := []struct {
		status taskstore.TaskStatus
		emoji  string
		color  lipgloss.Color
	}{
		{taskstore.StatusInProgress, "🔄", "4"},
		{taskstore.StatusTodo, "⭕", "7"},
		{taskstore.StatusPaused, "⏸️", "3"},
		{taskstore.StatusDone, "✅", "2"},
	}

	// Group tasks by status
	tasksByStatus := make(map[taskstore.TaskStatus][]taskstore.Task)
	for _, task := range m.tasks {
		tasksByStatus[task.Status] = append(tasksByStatus[task.Status], task)
	}

	currentIndex := 0
	var currentTask taskstore.Task

	for _, section := range statusSections {
		if tasks := tasksByStatus[section.status]; len(tasks) > 0 {
			sectionTitle := lipgloss.NewStyle().
				Bold(true).
				Foreground(section.color).
				Render(fmt.Sprintf("\n%s %s", section.emoji, strings.ToUpper(string(section.status))))
			sections = append(sections, sectionTitle)

			for _, task := range tasks {
				orderedTasks = append(orderedTasks, task)
				sections = append(sections, m.renderTask(task, currentIndex == m.cursor))
				if currentIndex == m.cursor {
					currentTask = task
				}
				currentIndex++
			}
		}
	}

	// Mode indicator and edit buffer
	if m.mode != ModeNormal {
		sections = append(sections, m.renderEditMode())
	}

	// Help text
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("\nj/k: move • a: add task • t/p/d/s: change status • e: edit title • o: edit desc (Ctrl+S) • x: delete • q: quit")

	descriptionSection := m.renderDescription(currentTask, currentIndex == m.cursor)

	sections = append(sections, helpText)

	// Message
	if m.message != "" && time.Since(m.messageTimer) < 3*time.Second {
		msgStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Bold(true)
		sections = append(sections, msgStyle.Render(m.message))
	}

	leftSide := strings.Join(sections, "\n")

	var split []string
	split = append(split, leftSide)
	split = append(split, descriptionSection)

	return lipgloss.JoinHorizontal(lipgloss.Top, split...)
}

func (m *InteractiveList) renderEmpty() string {
	emptyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("3")).
		Bold(true)
	return emptyStyle.Render("📭 No tasks found! Press 'a' to add a task or 'q' to quit.")
}

func (m *InteractiveList) renderDescription(task taskstore.Task, selected bool) string {
	style := lipgloss.NewStyle()
	result := style.Render("Details:")
	if task.Description != "" {
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		if selected {
			descStyle = descStyle.Background(lipgloss.Color("8")).Foreground(lipgloss.Color("7"))
		}

		// Handle multi-line descriptions with better formatting
		descLines := strings.Split(task.Description, "\n")
		for i, line := range descLines {
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine != "" {
				prefix := "     "
				if i > 0 {
					prefix = "     │ " // Visual continuation for multi-line
				}
				result += "\n" + descStyle.Render(fmt.Sprintf("%s%s", prefix, trimmedLine))
			} else if i > 0 && i < len(descLines)-1 {
				// Show empty lines in multi-line descriptions
				result += "\n" + descStyle.Render("     │")
			}
		}
	}
	return result

}

func (m *InteractiveList) renderTask(task taskstore.Task, selected bool) string {
	var style lipgloss.Style
	if selected {
		style = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			Bold(true)
	} else {
		style = lipgloss.NewStyle()
	}

	timeInfo := ""
	if task.Status == taskstore.StatusInProgress && task.StartedAt != nil {
		timeInfo = fmt.Sprintf(" ⏱ %s", formatDuration(time.Since(*task.StartedAt)))
	} else if task.TotalTime > 0 {
		timeInfo = fmt.Sprintf(" ⌛ %s", formatDuration(task.TotalTime))
	}

	taskLine := fmt.Sprintf("  #%d %s%s", task.ID, task.Title, timeInfo)
	if selected {
		taskLine = "► " + taskLine[2:]
	}

	result := style.Render(taskLine)

	/*
		if task.Description != "" {
			descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
			if selected {
				descStyle = descStyle.Background(lipgloss.Color("8")).Foreground(lipgloss.Color("7"))
			}

			// Handle multi-line descriptions with better formatting
			descLines := strings.Split(task.Description, "\n")
			for i, line := range descLines {
				trimmedLine := strings.TrimSpace(line)
				if trimmedLine != "" {
					prefix := "     "
					if i > 0 {
						prefix = "     │ " // Visual continuation for multi-line
					}
					result += "\n" + descStyle.Render(fmt.Sprintf("%s%s", prefix, trimmedLine))
				} else if i > 0 && i < len(descLines)-1 {
					// Show empty lines in multi-line descriptions
					result += "\n" + descStyle.Render("     │")
				}
			}
		}
	*/

	return result
}

func (m *InteractiveList) renderEditMode() string {
	var modeText string
	switch m.mode {
	case ModeEdit:
		modeText = "EDIT TITLE"
	case ModeEditDescription:
		modeText = "EDIT DESCRIPTION"
	}

	editStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("5")).
		Foreground(lipgloss.Color("0")).
		Bold(true).
		Padding(0, 1).
		MarginLeft(1)

	var helpText string
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true).
		MarginLeft(1)

	if m.mode == ModeEditDescription {
		helpText = helpStyle.Render("ESC: cancel • ENTER: new line • Ctrl+S: save • Ctrl+U: clear")
		return m.renderMultiLineEditor(editStyle.Render(modeText), helpText)
	} else {
		helpText = helpStyle.Render("ESC: cancel • ENTER: save • Ctrl+U: clear")
		return m.renderSingleLineEditor(editStyle.Render(modeText), helpText)
	}
}

func (m *InteractiveList) renderSingleLineEditor(header, help string) string {
	maxWidth := m.width - 4
	if maxWidth < 30 {
		maxWidth = 30
	}
	if maxWidth > 60 {
		maxWidth = 60
	}

	inputStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("0")).
		Foreground(lipgloss.Color("7")).
		Padding(0, 1).
		Width(maxWidth)

	// Add blinking cursor effect
	cursorChar := "│"
	if time.Since(m.messageTimer)%time.Second < 500*time.Millisecond {
		cursorChar = "█"
	}
	displayText := m.editBuffer + cursorChar

	return fmt.Sprintf("\n%s\n%s\n%s",
		header,
		inputStyle.Render(displayText),
		help)
}

func (m *InteractiveList) renderMultiLineEditor(header, help string) string {
	// Create a bordered box for multi-line editing
	// Calculate appropriate width and height based on terminal size
	maxWidth := m.width - 6
	if maxWidth < 40 {
		maxWidth = 40
	}
	if maxWidth > 80 {
		maxWidth = 80
	}

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Padding(1).
		Width(maxWidth).
		Height(8).
		MarginLeft(1)

	contentStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("0")).
		Foreground(lipgloss.Color("7")).
		Padding(0, 1)

	// Process the edit buffer for display with line wrapping
	lines := strings.Split(m.editBuffer, "\n")
	displayLines := make([]string, 0)

	maxLineWidth := maxWidth - 6 // Account for padding and borders

	for i, line := range lines {
		if len(line) <= maxLineWidth {
			if i == len(lines)-1 {
				// Add blinking cursor to last line
				cursorChar := "│"
				if time.Since(m.messageTimer)%time.Second < 500*time.Millisecond {
					cursorChar = "█"
				}
				displayLines = append(displayLines, line+cursorChar)
			} else {
				displayLines = append(displayLines, line)
			}
		} else {
			// Wrap long lines
			for j := 0; j < len(line); j += maxLineWidth {
				end := j + maxLineWidth
				if end > len(line) {
					end = len(line)
				}
				wrappedLine := line[j:end]
				if i == len(lines)-1 && end == len(line) {
					cursorChar := "│"
					if time.Since(m.messageTimer)%time.Second < 500*time.Millisecond {
						cursorChar = "█"
					}
					wrappedLine += cursorChar
				}
				displayLines = append(displayLines, wrappedLine)
			}
		}
	}

	// If the buffer is empty, show cursor on first line
	if m.editBuffer == "" {
		cursorChar := "│"
		if time.Since(m.messageTimer)%time.Second < 500*time.Millisecond {
			cursorChar = "█"
		}
		displayLines = []string{cursorChar}
	}

	content := strings.Join(displayLines, "\n")

	// Show line and character count info
	lineCount := len(strings.Split(m.editBuffer, "\n"))
	charCount := len(m.editBuffer)
	if m.editBuffer == "" {
		lineCount = 0
		charCount = 0
	}

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)

	lineInfo := infoStyle.Render(fmt.Sprintf("📏 Lines: %d • Characters: %d", lineCount, charCount))

	// Add a title above the editor box
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")).
		Bold(true).
		MarginLeft(1)

	editorTitle := titleStyle.Render("📝 Description Editor")

	return fmt.Sprintf("\n%s\n%s\n%s\n%s\n%s",
		header,
		editorTitle,
		borderStyle.Render(contentStyle.Render(content)),
		lineInfo,
		help)
}

func (m *InteractiveList) changeTaskStatus(status taskstore.TaskStatus) tea.Cmd {
	if len(m.tasks) == 0 {
		return nil
	}

	task := m.getCurrentTask()
	if task == nil {
		return nil
	}

	taskID := task.ID
	return func() tea.Msg {
		updates := map[string]any{"status": status}

		if status == taskstore.StatusInProgress {
			updates["started_at"] = time.Now()
		}

		err := taskstore.UpdateTask(taskID, updates)
		if err != nil {
			return taskStatusChangedMsg{id: taskID, status: "error"}
		}
		return taskStatusChangedMsg{id: taskID, status: status}
	}
}

func (m *InteractiveList) getCurrentTask() *taskstore.Task {
	if len(m.tasks) == 0 {
		return nil
	}

	// Create ordered task list to match display order
	orderedTasks := []taskstore.Task{}
	statusOrder := []taskstore.TaskStatus{
		taskstore.StatusInProgress,
		taskstore.StatusTodo,
		taskstore.StatusPaused,
		taskstore.StatusDone,
	}

	tasksByStatus := make(map[taskstore.TaskStatus][]taskstore.Task)
	for _, task := range m.tasks {
		tasksByStatus[task.Status] = append(tasksByStatus[task.Status], task)
	}

	for _, status := range statusOrder {
		if tasks := tasksByStatus[status]; len(tasks) > 0 {
			orderedTasks = append(orderedTasks, tasks...)
		}
	}

	if m.cursor >= 0 && m.cursor < len(orderedTasks) {
		return &orderedTasks[m.cursor]
	}

	return nil
}

func (m *InteractiveList) deleteCurrentTask() tea.Cmd {
	task := m.getCurrentTask()
	if task == nil {
		return nil
	}

	taskID := task.ID
	return func() tea.Msg {
		err := taskstore.DeleteTask(taskID)
		if err != nil {
			return taskDeletedMsg{id: -1} // Error indicator
		}
		return taskDeletedMsg{id: taskID}
	}
}

func (m *InteractiveList) updateTaskTitle(id int, title string) tea.Cmd {
	return func() tea.Msg {
		updates := map[string]any{"title": title}
		err := taskstore.UpdateTask(id, updates)
		if err != nil {
			return editCompleteMsg{id: id, field: "title", value: "error"}
		}
		return editCompleteMsg{id: id, field: "title", value: title}
	}
}

func (m *InteractiveList) updateTaskDescription(id int, description string) tea.Cmd {
	return func() tea.Msg {
		updates := map[string]any{"description": description}
		err := taskstore.UpdateTask(id, updates)
		if err != nil {
			return editCompleteMsg{id: id, field: "description", value: "error"}
		}
		return editCompleteMsg{id: id, field: "description", value: description}
	}
}

func (m *InteractiveList) reloadTasks() tea.Cmd {
	return func() tea.Msg {
		tasks, err := taskstore.GetTasks()
		if err != nil {
			return tasksReloadedMsg{tasks: []taskstore.Task{}}
		}

		return tasksReloadedMsg{tasks: tasks}
	}
}

func (m *InteractiveList) countDisplayTasks() int {
	statusOrder := []taskstore.TaskStatus{
		taskstore.StatusInProgress,
		taskstore.StatusTodo,
		taskstore.StatusPaused,
		taskstore.StatusDone,
	}

	tasksByStatus := make(map[taskstore.TaskStatus][]taskstore.Task)
	for _, task := range m.tasks {
		tasksByStatus[task.Status] = append(tasksByStatus[task.Status], task)
	}

	count := 0
	for _, status := range statusOrder {
		count += len(tasksByStatus[status])
	}

	return count
}

func (m *InteractiveList) setMessage(msg string) {
	m.message = msg
	m.messageTimer = time.Now()
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// handleAddTaskMode handles keyboard input when adding a new task
func (m *InteractiveList) handleAddTaskMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel adding task
		m.mode = ModeNormal
		m.newTaskTitle = ""
		m.newTaskDescription = ""
		m.editBuffer = ""
		return m, nil

	case "tab":
		// Switch between title and description fields
		if m.newTaskTitle == "" {
			// Currently editing title, switch to description
			m.newTaskTitle = m.editBuffer
			m.editBuffer = m.newTaskDescription
		} else {
			// Currently editing description, switch to title
			m.newTaskDescription = m.editBuffer
			m.editBuffer = m.newTaskTitle
			m.newTaskTitle = ""
		}
		return m, nil

	case "enter":
		if m.newTaskTitle == "" {
			// Still editing title, move to description
			m.newTaskTitle = m.editBuffer
			m.editBuffer = ""
			return m, nil
		} else {
			// Save the new task
			m.newTaskDescription = m.editBuffer
			if m.newTaskTitle != "" {
				err := taskstore.AddTask(m.newTaskTitle, m.newTaskDescription)
				if err != nil {
					m.setMessage(fmt.Sprintf("Error adding task: %v", err))
				} else {
					m.setMessage("Task added successfully")
					m.mode = ModeNormal
					m.newTaskTitle = ""
					m.newTaskDescription = ""
					m.editBuffer = ""
					return m, m.reloadTasks()
				}
			}
			return m, nil
		}

	case "backspace":
		if len(m.editBuffer) > 0 {
			m.editBuffer = m.editBuffer[:len(m.editBuffer)-1]
		}
		return m, nil

	default:
		// Add character to buffer
		if len(msg.String()) == 1 {
			m.editBuffer += msg.String()
		}
		return m, nil
	}
}

// renderAddTaskMode renders the view for adding a new task
func (m *InteractiveList) renderAddTaskMode() string {
	var b strings.Builder

	// Header
	header := " ➕ Add New Task "
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("40")).
		Background(lipgloss.Color("235")).
		Width(m.width).
		Align(lipgloss.Center)
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n\n")

	// Title field
	titleLabel := "Title: "
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		MarginLeft(2)

	b.WriteString(titleStyle.Render(titleLabel))

	titleContent := m.editBuffer
	if m.newTaskTitle != "" {
		titleContent = m.newTaskTitle
	}

	titleBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(func() string {
			if m.newTaskTitle == "" {
				return "39" // Active field
			}
			return "240"
		}())).
		Width(m.width-12).
		Padding(0, 1)

	if m.newTaskTitle == "" {
		titleContent += "█" // Cursor
	}

	b.WriteString(titleBoxStyle.Render(titleContent))
	b.WriteString("\n\n")

	// Description field
	descLabel := "Description (optional): "
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		MarginLeft(2)

	b.WriteString(descStyle.Render(descLabel))

	descContent := m.newTaskDescription
	if m.newTaskTitle != "" {
		descContent = m.editBuffer
		if descContent == "" {
			descContent = " "
		}
	}

	descBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(func() string {
			if m.newTaskTitle != "" {
				return "39" // Active field
			}
			return "240"
		}())).
		Width(m.width-12).
		Height(5).
		Padding(0, 1)

	if m.newTaskTitle != "" {
		lines := strings.Split(descContent, "\n")
		if len(lines) > 0 {
			lines[len(lines)-1] += "█" // Cursor
		}
		descContent = strings.Join(lines, "\n")
	}

	b.WriteString(descBoxStyle.Render(descContent))
	b.WriteString("\n\n")

	// Help text
	helpText := " tab: switch fields | enter: next/save | esc: cancel "
	helpStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("240")).
		Width(m.width).
		Align(lipgloss.Center)

	// Position help at bottom
	currentLines := strings.Count(b.String(), "\n") + 1
	for i := currentLines; i < m.height-1; i++ {
		b.WriteString("\n")
	}
	b.WriteString(helpStyle.Render(helpText))

	return b.String()
}

// RunInteractiveList starts the interactive list TUI
func RunInteractiveList() error {
	tasks, err := taskstore.GetTasks()
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	model := NewInteractiveList(tasks)
	p := tea.NewProgram(model, tea.WithAltScreen())

	_, err = p.Run()
	return err
}
