package screens

import (
	"fmt"
	"strings"
	"time"

	"godo/internal/taskstore"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Local message types to avoid import cycles
type (
	ShowAddTaskModalMsg     struct{}
	ShowEditGitLinkModalMsg struct{ TaskID int }
	ShowEditTitleModalMsg   struct{ TaskID int }
	ShowDeleteTaskModalMsg  struct{ TaskID int }
)

type TaskListScreen struct {
	tasks              []taskstore.Task
	orderedTasks       []taskstore.Task // The display order (by status groups)
	cursor             int
	width              int
	height             int
	message            string
	messageTimer       time.Time
	editBuffer         string
	mode               Mode
	taskStatusChanging bool
}

type Mode int

const (
	ModeNormal Mode = iota
	ModeEdit
	ModeEditDescription
	ModeEditGitLink
	ModeAddTask
)

func NewTaskListScreen(tasks []taskstore.Task) *TaskListScreen {
	screen := &TaskListScreen{
		tasks:  tasks,
		cursor: 0,
		width:  80,
		height: 20,
		mode:   ModeNormal,
	}
	// Build the ordered tasks list
	screen.rebuildOrderedTasks()
	return screen
}

// rebuildOrderedTasks rebuilds the ordered task list based on status groups
func (s *TaskListScreen) rebuildOrderedTasks() {
	s.orderedTasks = []taskstore.Task{}

	statusSections := []struct {
		status taskstore.TaskStatus
	}{
		{taskstore.StatusInProgress},
		{taskstore.StatusTodo},
		{taskstore.StatusPaused},
		{taskstore.StatusDone},
	}

	// Group tasks by status
	tasksByStatus := make(map[taskstore.TaskStatus][]taskstore.Task)
	for _, task := range s.tasks {
		tasksByStatus[task.Status] = append(tasksByStatus[task.Status], task)
	}

	for _, section := range statusSections {
		if tasks, ok := tasksByStatus[section.status]; ok {
			s.orderedTasks = append(s.orderedTasks, tasks...)
		}
	}

	// Ensure cursor is still valid
	if s.cursor >= len(s.orderedTasks) && len(s.orderedTasks) > 0 {
		s.cursor = len(s.orderedTasks) - 1
	}
}

func (s *TaskListScreen) Init() tea.Cmd {
	return nil
}

func (s *TaskListScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		// Handle other message types like WindowSizeMsg
		if windowMsg, ok := msg.(tea.WindowSizeMsg); ok {
			s.width = windowMsg.Width
			s.height = windowMsg.Height
		}
		return s, nil
	}

	// Handle different modes
	switch s.mode {
	case ModeNormal:
		return s.handleNormalMode(keyMsg)
	case ModeEdit:
		return s.handleEditMode(keyMsg)
	case ModeEditDescription:
		return s.handleEditDescriptionMode(keyMsg)
	case ModeEditGitLink:
		return s.handleEditGitLinkMode(keyMsg)
	}

	return s, nil
}

func (s *TaskListScreen) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if s.cursor < len(s.orderedTasks)-1 {
			s.cursor++
		}

	case "k", "up":
		if s.cursor > 0 {
			s.cursor--
		}

	case "a":
		return s, func() tea.Msg { return ShowAddTaskModalMsg{} }

	case "e":
		// Edit task title
		s.mode = ModeEdit
		if task := s.getCurrentTask(); task != nil {
			s.editBuffer = task.Title
		}

	case "g":
		// Edit git link
		if task := s.getCurrentTask(); task != nil {
			return s, func() tea.Msg {
				return ShowEditGitLinkModalMsg{TaskID: task.ID}
			}
		}

	case "o":
		// Edit description
		s.mode = ModeEditDescription
		if task := s.getCurrentTask(); task != nil {
			s.editBuffer = task.Description
		}

	case "x":
		// Delete task
		if task := s.getCurrentTask(); task != nil {
			return s, func() tea.Msg {
				return ShowDeleteTaskModalMsg{TaskID: task.ID}
			}
		}

	// Status changes
	case "t":
		if task := s.getCurrentTask(); task != nil {
			if task.Status == taskstore.StatusTodo {
				s.changeTaskStatus(task.ID, taskstore.StatusInProgress)
			} else {
				s.changeTaskStatus(task.ID, taskstore.StatusTodo)
			}
			s.rebuildOrderedTasks()
		}

	case "d":
		if task := s.getCurrentTask(); task != nil {
			s.changeTaskStatus(task.ID, taskstore.StatusDone)
			s.rebuildOrderedTasks()
		}

	case "p":
		if task := s.getCurrentTask(); task != nil {
			s.changeTaskStatus(task.ID, taskstore.StatusInProgress)
			s.rebuildOrderedTasks()
		}

	case "s":
		if task := s.getCurrentTask(); task != nil {
			s.changeTaskStatus(task.ID, taskstore.StatusPaused)
			s.rebuildOrderedTasks()
		}

	case "ctrl+c", "q":
		return s, tea.Quit
	}

	return s, nil
}

func (s *TaskListScreen) handleEditMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "escape":
		s.mode = ModeNormal
		s.editBuffer = ""

	case "enter":
		if task := s.getCurrentTask(); task != nil && s.editBuffer != "" {
			// Update the task in the main tasks array
			for i := range s.tasks {
				if s.tasks[i].ID == task.ID {
					s.tasks[i].Title = s.editBuffer
					break
				}
			}
			s.rebuildOrderedTasks()
			s.mode = ModeNormal
			s.editBuffer = ""
			s.setMessage(fmt.Sprintf("Task #%d title updated", task.ID))
		}

	case "backspace":
		if len(s.editBuffer) > 0 {
			s.editBuffer = s.editBuffer[:len(s.editBuffer)-1]
		}

	default:
		if msg.Type == tea.KeyRunes {
			s.editBuffer += string(msg.Runes)
		} else if msg.Type == tea.KeySpace {
			s.editBuffer += " "
		}
	}

	return s, nil
}

func (s *TaskListScreen) handleEditDescriptionMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "escape":
		s.mode = ModeNormal
		s.editBuffer = ""

	case "ctrl+s":
		if task := s.getCurrentTask(); task != nil {
			// Update the task in the main tasks array
			for i := range s.tasks {
				if s.tasks[i].ID == task.ID {
					s.tasks[i].Description = s.editBuffer
					break
				}
			}
			s.rebuildOrderedTasks()
			s.mode = ModeNormal
			s.editBuffer = ""
			s.setMessage(fmt.Sprintf("Task #%d description updated", task.ID))
		}

	case "ctrl+u":
		s.editBuffer = ""

	case "backspace":
		if len(s.editBuffer) > 0 {
			s.editBuffer = s.editBuffer[:len(s.editBuffer)-1]
		}

	default:
		if msg.Type == tea.KeyRunes {
			s.editBuffer += string(msg.Runes)
		} else if msg.Type == tea.KeySpace {
			s.editBuffer += " "
		}
	}

	return s, nil
}

func (s *TaskListScreen) handleEditGitLinkMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "escape":
		s.mode = ModeNormal
		s.editBuffer = ""

	case "enter":
		// TODO: Save git link
		s.mode = ModeNormal
		s.editBuffer = ""

	case "backspace":
		if len(s.editBuffer) > 0 {
			s.editBuffer = s.editBuffer[:len(s.editBuffer)-1]
		}

	default:
		if msg.Type == tea.KeyRunes {
			s.editBuffer += string(msg.Runes)
		} else if msg.Type == tea.KeySpace {
			s.editBuffer += " "
		}
	}

	return s, nil
}

func (s *TaskListScreen) View() string {
	// If we're in a modal mode, don't render the task list
	if len(s.orderedTasks) == 0 {
		return "No tasks. Press 'a' to add one."
	}

	var sections []string

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("4")).
		Render("📋 Tasks")
	sections = append(sections, header)

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
	for _, task := range s.orderedTasks {
		tasksByStatus[task.Status] = append(tasksByStatus[task.Status], task)
	}

	currentIndex := 0
	var currentTask *taskstore.Task

	for _, section := range statusSections {
		if tasks := tasksByStatus[section.status]; len(tasks) > 0 {
			sectionTitle := lipgloss.NewStyle().
				Bold(true).
				Foreground(section.color).
				Margin(1, 0, 0, 0).
				Render(fmt.Sprintf("%s %s", section.emoji, strings.ToUpper(string(section.status))))
			sections = append(sections, sectionTitle)

			for i, task := range tasks {
				isSelected := currentIndex == s.cursor
				sections = append(sections, s.renderTask(task, isSelected))

				// Keep a pointer to the selected task
				if isSelected {
					currentTask = &tasks[i]
				}
				currentIndex++
			}
		}
	}

	// Mode indicator and edit buffer
	if s.mode != ModeNormal {
		sections = append(sections, s.renderEditMode())
	}

	// Render description - use currentTask which is guaranteed to be the selected one
	descriptionSection := ""
	if currentTask != nil {
		descriptionSection = s.renderDescription(*currentTask)
	} else {
		descriptionSection = "Select a task to view details"
	}

	leftSide := strings.Join(sections, "\n")

	leftSideWidth := 40
	leftSideStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(leftSideWidth).
		Height(s.height - 4)
	leftSideWithBorder := leftSideStyle.Render(leftSide)

	descriptionWidth := s.width - leftSideWidth - 8
	if descriptionWidth < 20 {
		descriptionWidth = 20
	}

	descriptionStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Width(descriptionWidth).
		Height(s.height - 4)
	descriptionWithBorder := descriptionStyle.Render(descriptionSection)

	// Combine left and right panels
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftSideWithBorder, descriptionWithBorder)

	// Build footer with help text and message
	var footerSections []string

	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("a: add • t/p/d/s: status • e: title • o: desc • g: git • x: delete • q: quit")
	footerSections = append(footerSections, helpText)

	// Message
	if s.message != "" && time.Since(s.messageTimer) < 3*time.Second {
		msgStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Bold(true)
		footerSections = append(footerSections, msgStyle.Render(s.message))
	}

	footer := strings.Join(footerSections, "\n")

	// Combine main content and footer vertically
	return lipgloss.JoinVertical(lipgloss.Left, mainContent, footer)
}

func (s *TaskListScreen) getStatusEmoji(status taskstore.TaskStatus) string {
	switch status {
	case taskstore.StatusInProgress:
		return "🔄"
	case taskstore.StatusTodo:
		return "⭕"
	case taskstore.StatusPaused:
		return "⏸️"
	case taskstore.StatusDone:
		return "✅"
	default:
		return "❓"
	}
}

func (s *TaskListScreen) renderTask(task taskstore.Task, isSelected bool) string {
	style := lipgloss.NewStyle().Padding(0, 1)
	if isSelected {
		style = style.Background(lipgloss.Color("4")).Foreground(lipgloss.Color("0")).Bold(true)
	}

	return style.Render(fmt.Sprintf("%s", task.Title))
}

func (s *TaskListScreen) renderDescription(task taskstore.Task) string {
	// Title section - top
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("4")).
		MarginBottom(1)

	titleText := titleStyle.Render(task.Title)

	// Status section with emoji
	statusEmoji := s.getStatusEmoji(task.Status)
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		MarginBottom(1)

	statusText := statusStyle.Render(fmt.Sprintf("%s %s", statusEmoji, strings.ToUpper(string(task.Status))))

	// Description section - middle
	var descText string
	if task.Description != "" {
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")).
			MarginBottom(1)

		descHeaderStyle := lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("8"))

		descText = descHeaderStyle.Render("Description") + "\n" + descStyle.Render(task.Description)
	}

	// Git links section - bottom
	var gitText string
	if len(task.GitLinks) > 0 {
		gitHeaderStyle := lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("8")).
			MarginTop(1)

		gitText = gitHeaderStyle.Render("Git Links")

		linkStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).
			MarginLeft(2)

		for _, link := range task.GitLinks {
			if link.Name == "" {
				continue
			}
			gitText += "\n" + linkStyle.Render("→ "+link.Name)
		}
	}

	// Combine all sections
	var result string
	if descText != "" && gitText != "" {
		result = lipgloss.JoinHorizontal(lipgloss.Top, titleText, " ", statusText) + "\n\n" + descText + "\n\n" + gitText
	} else if descText != "" {
		result = lipgloss.JoinHorizontal(lipgloss.Top, titleText, " ", statusText) + "\n\n" + descText
	} else if gitText != "" {
		result = lipgloss.JoinHorizontal(lipgloss.Top, titleText, " ", statusText) + "\n\n" + gitText
	} else {
		result = lipgloss.JoinHorizontal(lipgloss.Top, titleText, " ", statusText)
	}

	return result
}

func (s *TaskListScreen) renderEditMode() string {
	var modeText string
	switch s.mode {
	case ModeEdit:
		modeText = "EDIT TITLE"
	case ModeEditDescription:
		modeText = "EDIT DESCRIPTION"
	case ModeEditGitLink:
		modeText = "EDIT GIT LINK"
	default:
		return ""
	}

	editStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		MarginLeft(1)

	var helpText string
	helpStyle := lipgloss.NewStyle().
		Italic(true).
		MarginLeft(1)

	if s.mode == ModeEditDescription {
		helpText = helpStyle.Render("ESC: cancel • Ctrl+S: save • Ctrl+U: clear")
	} else {
		helpText = helpStyle.Render("ESC: cancel • ENTER: save")
	}

	inputStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("0")).
		Foreground(lipgloss.Color("7")).
		Padding(0, 1).
		Width(35)

	displayText := s.editBuffer + "│"

	return fmt.Sprintf("\n%s\n%s\n%s",
		editStyle.Render(modeText),
		inputStyle.Render(displayText),
		helpText)
}

func (s *TaskListScreen) getCurrentTask() *taskstore.Task {
	if s.cursor >= 0 && s.cursor < len(s.orderedTasks) {
		// Find the task in the original tasks array by ID
		orderedTask := s.orderedTasks[s.cursor]
		for i := range s.tasks {
			if s.tasks[i].ID == orderedTask.ID {
				return &s.tasks[i]
			}
		}
	}
	return nil
}

func (s *TaskListScreen) changeTaskStatus(taskID int, newStatus taskstore.TaskStatus) {
	for i := range s.tasks {
		if s.tasks[i].ID == taskID {
			s.tasks[i].Status = newStatus
			s.setMessage(fmt.Sprintf("Task #%d status changed to %s", taskID, newStatus))
			break
		}
	}
}

func (s *TaskListScreen) setMessage(msg string) {
	s.message = msg
	s.messageTimer = time.Now()
}

func (s *TaskListScreen) UpdateTasks(tasks []taskstore.Task) {
	s.tasks = tasks
	s.rebuildOrderedTasks()
}
