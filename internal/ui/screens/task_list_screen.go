package screens

import (
	"fmt"
	"strings"
	"time"

	"godo/internal/taskstore"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Local message types
type (
	ShowAddTaskModalMsg       struct{}
	ShowAddProjectModalMsg    struct{}
	ShowEditProjectModalMsg   struct{ ProjectID int }
	ShowDeleteProjectModalMsg struct{ ProjectID int }
	ShowEditGitLinkModalMsg   struct{ TaskID int }
	ShowEditTitleModalMsg     struct{ TaskID int }
	ShowDeleteTaskModalMsg    struct{ TaskID int }
	ShowAddTaskToProjectMsg   struct{ ProjectID int }
	ShowEditProjectGitLinkMsg struct{ ProjectID int }
)

type DisplayItem struct {
	Type      string // "project" or "task"
	ProjectID int    // Set if Type == "project"
	Task      *taskstore.Task
}

type TaskListScreen struct {
	tasks        []taskstore.Task
	projects     []taskstore.Project
	displayItems []DisplayItem
	cursor       int
	width        int
	height       int
	message      string
	messageTimer time.Time
	editBuffer   string
	mode         Mode
}

type Mode int

const (
	ModeNormal Mode = iota
	ModeEdit
	ModeEditDescription
	ModeEditGitLink
	ModeAddTask
)

func NewTaskListScreen(tasks []taskstore.Task, projects []taskstore.Project) *TaskListScreen {
	screen := &TaskListScreen{
		tasks:    tasks,
		projects: projects,
		cursor:   0,
		width:    80,
		height:   20,
		mode:     ModeNormal,
	}
	screen.rebuildDisplayItems()
	return screen
}

func (s *TaskListScreen) Init() tea.Cmd {
	return nil
}

func (s *TaskListScreen) rebuildDisplayItems() {
	s.displayItems = []DisplayItem{}

	// Add all projects with their tasks
	for _, project := range s.projects {
		s.displayItems = append(s.displayItems, DisplayItem{
			Type:      "project",
			ProjectID: project.ID,
		})

		for _, task := range s.tasks {
			for _, projID := range task.ProjectIDs {
				if projID == project.ID {
					taskCopy := task
					s.displayItems = append(s.displayItems, DisplayItem{
						Type: "task",
						Task: &taskCopy,
					})
					break
				}
			}
		}
	}

	// Add unassigned tasks
	if len(s.getUnassignedTasks()) > 0 {
		s.displayItems = append(s.displayItems, DisplayItem{
			Type:      "project",
			ProjectID: 0,
		})

		for _, task := range s.getUnassignedTasks() {
			taskCopy := task
			s.displayItems = append(s.displayItems, DisplayItem{
				Type: "task",
				Task: &taskCopy,
			})
		}
	}

	if s.cursor >= len(s.displayItems) && len(s.displayItems) > 0 {
		s.cursor = len(s.displayItems) - 1
	}
}

func (s *TaskListScreen) getUnassignedTasks() []taskstore.Task {
	var unassigned []taskstore.Task
	for _, task := range s.tasks {
		if len(task.ProjectIDs) == 0 {
			unassigned = append(unassigned, task)
		}
	}
	return unassigned
}

func (s *TaskListScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		if windowMsg, ok := msg.(tea.WindowSizeMsg); ok {
			s.width = windowMsg.Width
			s.height = windowMsg.Height
		}
		return s, nil
	}

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
		if s.cursor < len(s.displayItems)-1 {
			s.cursor++
		}

	case "k", "up":
		if s.cursor > 0 {
			s.cursor--
		}

	case "a":
		// Add task to current project (if on a project) or create unassigned task
		if item := s.getCurrentDisplayItem(); item != nil && item.Type == "project" {
			return s, func() tea.Msg {
				return ShowAddTaskToProjectMsg{ProjectID: item.ProjectID}
			}
		}
		// Otherwise show regular add task modal
		return s, func() tea.Msg { return ShowAddTaskModalMsg{} }

	case "A":
		return s, func() tea.Msg { return ShowAddProjectModalMsg{} }

	case "e":
		if item := s.getCurrentDisplayItem(); item != nil && item.Type == "task" && item.Task != nil {
			s.mode = ModeEdit
			s.editBuffer = item.Task.Title
		}

	case "g":
		// Git links for both tasks and projects
		if item := s.getCurrentDisplayItem(); item != nil {
			if item.Type == "task" && item.Task != nil {
				return s, func() tea.Msg {
					return ShowEditGitLinkModalMsg{TaskID: item.Task.ID}
				}
			} else if item.Type == "project" && item.ProjectID != 0 {
				return s, func() tea.Msg {
					return ShowEditProjectGitLinkMsg{ProjectID: item.ProjectID}
				}
			}
		}

	case "o":
		if item := s.getCurrentDisplayItem(); item != nil && item.Type == "task" && item.Task != nil {
			s.mode = ModeEditDescription
			s.editBuffer = item.Task.Description
		}

	case "x":
		if item := s.getCurrentDisplayItem(); item != nil {
			if item.Type == "task" && item.Task != nil {
				return s, func() tea.Msg {
					return ShowDeleteTaskModalMsg{TaskID: item.Task.ID}
				}
			} else if item.Type == "project" {
				return s, func() tea.Msg {
					return ShowDeleteProjectModalMsg{ProjectID: item.ProjectID}
				}
			}
		}

	case "t":
		if item := s.getCurrentDisplayItem(); item != nil && item.Type == "task" && item.Task != nil {
			task := item.Task
			if task.Status == taskstore.StatusTodo {
				s.changeTaskStatus(task.ID, taskstore.StatusInProgress)
			} else {
				s.changeTaskStatus(task.ID, taskstore.StatusTodo)
			}
			s.rebuildDisplayItems()
		}

	case "d":
		if item := s.getCurrentDisplayItem(); item != nil && item.Type == "task" && item.Task != nil {
			s.changeTaskStatus(item.Task.ID, taskstore.StatusDone)
			s.rebuildDisplayItems()
		}

	case "p":
		if item := s.getCurrentDisplayItem(); item != nil && item.Type == "task" && item.Task != nil {
			s.changeTaskStatus(item.Task.ID, taskstore.StatusInProgress)
			s.rebuildDisplayItems()
		}

	case "s":
		if item := s.getCurrentDisplayItem(); item != nil && item.Type == "task" && item.Task != nil {
			s.changeTaskStatus(item.Task.ID, taskstore.StatusPaused)
			s.rebuildDisplayItems()
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
			for i := range s.tasks {
				if s.tasks[i].ID == task.ID {
					s.tasks[i].Title = s.editBuffer
					break
				}
			}
			s.rebuildDisplayItems()
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
			for i := range s.tasks {
				if s.tasks[i].ID == task.ID {
					s.tasks[i].Description = s.editBuffer
					break
				}
			}
			s.rebuildDisplayItems()
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
	if len(s.displayItems) == 0 {
		return "No tasks or projects. Press 'a' to add a task or 'A' to add a project."
	}

	var sections []string

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("4")).
		Render("📋 Projects & Tasks")
	sections = append(sections, header)

	for i, item := range s.displayItems {
		isSelected := i == s.cursor
		if item.Type == "project" {
			sections = append(sections, s.renderProjectHeader(item.ProjectID, isSelected))
		} else if item.Task != nil {
			sections = append(sections, s.renderTask(*item.Task, isSelected))
		}
	}

	if s.mode != ModeNormal {
		sections = append(sections, s.renderEditMode())
	}

	descriptionSection := ""
	if item := s.getCurrentDisplayItem(); item != nil && item.Type == "task" && item.Task != nil {
		descriptionSection = s.renderDescription(*item.Task)
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

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftSideWithBorder, descriptionWithBorder)

	var footerSections []string

	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("a: task/project-task • A: project • e: title • o: desc • g: git-link • t/p/d/s: status • x: delete • q: quit")
	footerSections = append(footerSections, helpText)

	if s.message != "" && time.Since(s.messageTimer) < 3*time.Second {
		msgStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Bold(true)
		footerSections = append(footerSections, msgStyle.Render(s.message))
	}

	footer := strings.Join(footerSections, "\n")

	return lipgloss.JoinVertical(lipgloss.Left, mainContent, footer)
}

func (s *TaskListScreen) renderProjectHeader(projectID int, isSelected bool) string {
	var projectName string

	if projectID == 0 {
		projectName = "📁 Unassigned"
	} else {
		for _, proj := range s.projects {
			if proj.ID == projectID {
				projectName = fmt.Sprintf("📁 %s", proj.Name)
				break
			}
		}
	}

	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("4")).
		Margin(1, 0, 0, 0)

	if isSelected {
		style = style.Background(lipgloss.Color("4")).Foreground(lipgloss.Color("0"))
	}

	return style.Render(projectName)
}

func (s *TaskListScreen) renderTask(task taskstore.Task, isSelected bool) string {
	statusEmoji := s.getStatusEmoji(task.Status)
	taskText := fmt.Sprintf("  %s %s", statusEmoji, task.Title)

	style := lipgloss.NewStyle().Padding(0, 1)
	if isSelected {
		style = style.Background(lipgloss.Color("4")).Foreground(lipgloss.Color("0")).Bold(true)
	}

	return style.Render(taskText)
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

func (s *TaskListScreen) renderDescription(task taskstore.Task) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("4")).
		MarginBottom(1)

	titleText := titleStyle.Render(task.Title)

	statusEmoji := s.getStatusEmoji(task.Status)
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		MarginBottom(1)

	statusText := statusStyle.Render(fmt.Sprintf("%s %s", statusEmoji, strings.ToUpper(string(task.Status))))

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

	var projectText string
	if len(task.ProjectIDs) > 0 {
		projHeaderStyle := lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("8")).
			MarginTop(1)

		projectText = projHeaderStyle.Render("Projects")

		projStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).
			MarginLeft(2)

		for _, projID := range task.ProjectIDs {
			for _, proj := range s.projects {
				if proj.ID == projID {
					projectText += "\n" + projStyle.Render("→ "+proj.Name)
					break
				}
			}
		}
	}

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

	var result string
	parts := []string{lipgloss.JoinHorizontal(lipgloss.Top, titleText, " ", statusText)}

	if descText != "" {
		parts = append(parts, descText)
	}
	if projectText != "" {
		parts = append(parts, projectText)
	}
	if gitText != "" {
		parts = append(parts, gitText)
	}

	result = strings.Join(parts, "\n\n")
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

func (s *TaskListScreen) getCurrentDisplayItem() *DisplayItem {
	if s.cursor >= 0 && s.cursor < len(s.displayItems) {
		return &s.displayItems[s.cursor]
	}
	return nil
}

func (s *TaskListScreen) getCurrentTask() *taskstore.Task {
	if item := s.getCurrentDisplayItem(); item != nil && item.Type == "task" && item.Task != nil {
		for i := range s.tasks {
			if s.tasks[i].ID == item.Task.ID {
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
	s.rebuildDisplayItems()
}

func (s *TaskListScreen) UpdateProjects(projects []taskstore.Project) {
	s.projects = projects
	s.rebuildDisplayItems()
}
