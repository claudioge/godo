package ui

import (
	"godo/internal/taskstore"
	"godo/internal/ui/modals"
	"godo/internal/ui/screens"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type App struct {
	taskList *screens.TaskListScreen
	modal    modals.Modal // Currently open modal (if any)
	width    int
	height   int
	tasks    []taskstore.Task
}

func NewApp(tasks []taskstore.Task) *App {
	return &App{
		taskList: screens.NewTaskListScreen(tasks),
		tasks:    tasks,
	}
}

func (a *App) Init() tea.Cmd {
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.taskList.Update(msg)
		return a, nil

	// Navigation: open modals
	case screens.ShowAddTaskModalMsg:
		a.modal = modals.NewAddTaskModal(a.width, a.height)
		return a, nil

	case screens.ShowEditGitLinkModalMsg:
		if task := a.findTask(msg.TaskID); task != nil {
			a.modal = modals.NewEditGitLinkModal(task, a.width, a.height)
		}
		return a, nil

	case screens.ShowEditTitleModalMsg:
		if task := a.findTask(msg.TaskID); task != nil {
			a.modal = modals.NewEditTitleModal(task, a.width, a.height)
		}
		return a, nil

	// Form events
	case modals.FormCancelledMsg:
		a.modal = nil // Close modal
		return a, nil

	case modals.FormSubmittedMsg:
		// Handle different form types
		return a.handleFormSubmission(msg)

	// Keyboard input
	case tea.KeyMsg:
		if a.modal != nil {
			// Modal has focus
			model, cmd := a.modal.Update(msg)
			a.modal = model.(modals.Modal)
			return a, cmd
		}
		// Main screen has focus
		model, cmd := a.taskList.Update(msg)
		a.taskList = model.(*screens.TaskListScreen)
		return a, cmd
	}

	return a, nil
}

func (a *App) View() string {
	base := a.taskList.View()

	if a.modal != nil {
		// Render modal centered on top of base
		return lipgloss.Place(
			a.width, a.height,
			lipgloss.Center, lipgloss.Center,
			a.modal.View(),
			lipgloss.WithWhitespaceChars(" "),
		)
	}

	return base
}

func (a *App) handleFormSubmission(msg modals.FormSubmittedMsg) (tea.Model, tea.Cmd) {
	switch msg.ModalType {
	case "add_task":
		title := msg.Data["Title"]
		description := msg.Data["Description"]

		if title != "" {
			newTask := taskstore.Task{
				ID:          a.getNextTaskID(),
				Title:       title,
				Description: description,
				Status:      taskstore.StatusTodo,
			}
			a.tasks = append(a.tasks, newTask)
			a.taskList.UpdateTasks(a.tasks)
		}

	case "edit_git_link":
		name := msg.Data["Name"]
		path := msg.Data["Repository Path"]
		link := msg.Data["Git Link"]
		branch := msg.Data["Branch"]

		if task := a.findTask(msg.TaskID); task != nil {
			task.GitLinks = append(task.GitLinks, taskstore.GitLink{
				Name:      name,
				LocalPath: path,
				Link:      link,
				Branch:    branch,
			})
			a.taskList.UpdateTasks(a.tasks)
		}
	}

	a.modal = nil
	return a, nil
}

func (a *App) findTask(id int) *taskstore.Task {
	for i := range a.tasks {
		if a.tasks[i].ID == id {
			return &a.tasks[i]
		}
	}
	return nil
}

func (a *App) getNextTaskID() int {
	maxID := 0
	for _, task := range a.tasks {
		if task.ID > maxID {
			maxID = task.ID
		}
	}
	return maxID + 1
}
