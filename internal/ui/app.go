package ui

import (
	"fmt"
	"godo/internal/git"
	"godo/internal/taskstore"
	"godo/internal/ui/modals"
	"godo/internal/ui/screens"
	"slices"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type App struct {
	taskList *screens.TaskListScreen
	modal    modals.Modal
	width    int
	height   int
	tasks    []taskstore.Task
	projects []taskstore.Project
}

func NewApp(tasks []taskstore.Task, projects []taskstore.Project) *App {
	return &App{
		taskList: screens.NewTaskListScreen(tasks, projects),
		tasks:    tasks,
		projects: projects,
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

	case screens.ShowAddTaskModalMsg:
		a.modal = modals.NewAddTaskModal(a.width, a.height)
		return a, nil

	case screens.ShowAddProjectModalMsg:
		a.modal = modals.NewAddProjectModal(a.width, a.height)
		return a, nil

	case screens.ShowEditProjectModalMsg:
		if project := a.findProject(msg.ProjectID); project != nil {
			a.modal = modals.NewEditProjectModal(project, a.width, a.height)
		}
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

	case screens.ShowEditDescriptionModalMsg:
		if task := a.findTask(msg.TaskID); task != nil {
			a.modal = modals.NewEditDescriptionModal(msg.TaskID, task.Description, a.width, a.height)
		}
		return a, nil

	case screens.ShowDeleteTaskModalMsg:
		for i, task := range a.tasks {
			if task.ID == msg.TaskID {
				a.tasks = append(a.tasks[:i], a.tasks[i+1:]...)
				a.taskList.UpdateTasks(a.tasks)
				if err := taskstore.DeleteTask(msg.TaskID); err != nil {
					fmt.Printf("Error deleting task: %v\n", err)
				}
				break
			}
		}
		return a, nil

	case screens.ShowDeleteProjectModalMsg:
		for i, project := range a.projects {
			if project.ID == msg.ProjectID {
				a.projects = append(a.projects[:i], a.projects[i+1:]...)
				a.taskList.UpdateProjects(a.projects)
				if err := taskstore.DeleteProject(msg.ProjectID); err != nil {
					fmt.Printf("Error deleting project: %v\n", err)
				}
				break
			}
		}
		return a, nil

	case screens.ShowAddTaskToProjectMsg:
		a.modal = modals.NewAddTaskToProjectModal(msg.ProjectID, a.width, a.height)
		return a, nil

	case screens.ShowSelectProjectModalMsg:
		a.modal = modals.NewSelectProjectModal(msg.TaskID, a.projects, a.width, a.height)
		return a, nil

	case screens.ShowRemoveTaskFromProjectMsg:
		if err := taskstore.RemoveTaskFromProject(msg.TaskID, msg.ProjectID); err != nil {
			fmt.Printf("Error removing task from project: %v\n", err)
		} else {
			for i := range a.tasks {
				if a.tasks[i].ID == msg.TaskID {
					a.tasks[i].ProjectIDs = slices.DeleteFunc(a.tasks[i].ProjectIDs, func(id int) bool {
						return id == msg.ProjectID
					})
					break
				}
			}
			a.taskList.UpdateTasks(a.tasks)
		}
		return a, nil

	case screens.ShowEditProjectGitLinkMsg:
		if project := a.findProject(msg.ProjectID); project != nil {
			a.modal = modals.NewEditProjectGitLinkModal(project, a.width, a.height)
		}
		return a, nil

	case screens.ShowSelectRepoModalMsg:
		a.modal = modals.NewSelectRepoModal(msg.ProjectID, a.width, a.height)
		return a, nil

	case screens.ShowCreateBranchModalMsg:
		task := a.findTask(msg.TaskID)
		project := a.findProject(msg.ProjectID)
		if task != nil && project != nil && len(project.GitLinks) > 0 {
			a.modal = modals.NewCreateBranchModal(task, project, project.GitLinks[0], a.width, a.height)
		}
		return a, nil

	case screens.ShowAIModalMsg:
		task := a.findTask(msg.TaskID)
		if task != nil && len(task.ProjectIDs) > 0 {
			projectID := task.ProjectIDs[0]
			project := a.findProject(projectID)
			if project != nil && len(project.GitLinks) > 0 && project.GitLinks[0].LocalPath != "" {
				a.modal = modals.NewAIModal(task, project, project.GitLinks[0], a.width, a.height)
			}
		}
		return a, nil

	case modals.BranchCreatedMsg:
		project := a.findProject(msg.ProjectID)
		if project != nil && len(project.GitLinks) > 0 {
			gitLink := project.GitLinks[0]
			if err := git.CreateBranch(gitLink.LocalPath, msg.BranchName); err != nil {
				fmt.Printf("Error creating branch: %v\n", err)
			} else {
				for i := range a.tasks {
					if a.tasks[i].ID == msg.TaskID {
						a.tasks[i].Status = taskstore.StatusInProgress
						break
					}
				}
				a.taskList.UpdateTasks(a.tasks)
				if err := taskstore.UpdateTask(msg.TaskID, map[string]any{
					"status": taskstore.StatusInProgress,
				}); err != nil {
					fmt.Printf("Error updating task status: %v\n", err)
				}
			}
		}
		a.modal = nil
		return a, nil

	case screens.ShowDeleteTaskGitLinkMsg:
		if err := taskstore.RemoveGitLinkFromTask(msg.TaskID, msg.LinkName); err != nil {
			fmt.Printf("Error removing git link from task: %v\n", err)
		} else {
			for i := range a.tasks {
				if a.tasks[i].ID == msg.TaskID {
					a.tasks[i].GitLinks = slices.DeleteFunc(a.tasks[i].GitLinks, func(gl taskstore.GitLink) bool {
						return gl.Name == msg.LinkName
					})
					break
				}
			}
			a.taskList.UpdateTasks(a.tasks)
		}
		return a, nil

	case screens.ShowDeleteProjectGitLinkMsg:
		if err := taskstore.RemoveGitLinkFromProject(msg.ProjectID, msg.LinkName); err != nil {
			fmt.Printf("Error removing git link from project: %v\n", err)
		} else {
			for i := range a.projects {
				if a.projects[i].ID == msg.ProjectID {
					a.projects[i].GitLinks = slices.DeleteFunc(a.projects[i].GitLinks, func(gl taskstore.GitLink) bool {
						return gl.Name == msg.LinkName
					})
					break
				}
			}
			a.taskList.UpdateProjects(a.projects)
		}
		return a, nil

	case modals.FormCancelledMsg:
		a.modal = nil
		return a, nil

	case modals.FormSubmittedMsg:
		return a.handleFormSubmission(msg)

	case tea.KeyMsg:
		if a.modal != nil {
			model, cmd := a.modal.Update(msg)
			a.modal = model.(modals.Modal)
			return a, cmd
		}
		model, cmd := a.taskList.Update(msg)
		a.taskList = model.(*screens.TaskListScreen)
		return a, cmd
	}

	return a, nil
}

func (a *App) View() string {
	base := a.taskList.View()

	if a.modal != nil {
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
				CreatedAt:   time.Now(),
				ProjectIDs:  []int{},
				GitLinks:    []taskstore.GitLink{},
			}
			a.tasks = append(a.tasks, newTask)
			a.taskList.UpdateTasks(a.tasks)
			if err := taskstore.AddTask(title, description); err != nil {
				fmt.Printf("Error saving task: %v\n", err)
			}
		}

	case "add_project":
		name := msg.Data["Name"]
		description := msg.Data["Description"]

		if name != "" {
			newProject := taskstore.Project{
				ID:          a.getNextProjectID(),
				Name:        name,
				Description: description,
				GitLinks:    []taskstore.GitLink{},
				CreatedAt:   time.Now(),
			}
			a.projects = append(a.projects, newProject)
			a.taskList.UpdateProjects(a.projects)
			if err := taskstore.AddProject(name, description); err != nil {
				fmt.Printf("Error saving project: %v\n", err)
			}
		}

	case "edit_git_link":
		name := msg.Data["Name"]
		path := msg.Data["Repository Path"]
		link := msg.Data["Git Link"]
		branch := msg.Data["Branch"]

		if task := a.findTask(msg.TaskID); task != nil {
			for i := range a.tasks {
				if a.tasks[i].ID == msg.TaskID {
					a.tasks[i].GitLinks = append(a.tasks[i].GitLinks, taskstore.GitLink{
						Name:      name,
						LocalPath: path,
						Link:      link,
						Branch:    branch,
					})
					break
				}
			}
			a.taskList.UpdateTasks(a.tasks)
			if err := taskstore.LinkTaskToGit(msg.TaskID, path, name, link, branch); err != nil {
				fmt.Printf("Error saving git link: %v\n", err)
			}
		}

	case "edit_project_git_link":
		name := msg.Data["Name"]
		path := msg.Data["Repository Path"]
		link := msg.Data["Git Link"]
		branch := msg.Data["Branch"]

		if project := a.findProject(msg.ProjectID); project != nil {
			for i := range a.projects {
				if a.projects[i].ID == msg.ProjectID {
					a.projects[i].GitLinks = append(a.projects[i].GitLinks, taskstore.GitLink{
						Name:      name,
						LocalPath: path,
						Link:      link,
						Branch:    branch,
					})
					break
				}
			}
			a.taskList.UpdateProjects(a.projects)
			if err := taskstore.AddGitLinkToProject(msg.ProjectID, taskstore.GitLink{
				Name:      name,
				LocalPath: path,
				Link:      link,
				Branch:    branch,
			}); err != nil {
				fmt.Printf("Error saving git link to project: %v\n", err)
			}
		}

	case "add_task_to_project":
		title := msg.Data["Title"]
		description := msg.Data["Description"]

		if title != "" {
			newTask := taskstore.Task{
				ID:          a.getNextTaskID(),
				Title:       title,
				Description: description,
				Status:      taskstore.StatusTodo,
				CreatedAt:   time.Now(),
				ProjectIDs:  []int{msg.ProjectID},
				GitLinks:    []taskstore.GitLink{},
			}
			a.tasks = append(a.tasks, newTask)
			a.taskList.UpdateTasks(a.tasks)
			if err := taskstore.AddTaskToProject(newTask.ID, msg.ProjectID); err != nil {
				fmt.Printf("Error linking task to project: %v\n", err)
			}
		}

	case "edit_title":
		title := msg.Data["Title"]

		if title != "" {
			for i := range a.tasks {
				if a.tasks[i].ID == msg.TaskID {
					a.tasks[i].Title = title
					if err := taskstore.UpdateTask(msg.TaskID, map[string]any{
						"title": title,
					}); err != nil {
						fmt.Printf("Error saving task title: %v\n", err)
					}
					break
				}
			}
			a.taskList.UpdateTasks(a.tasks)
		}

	case "edit_description":
		description := msg.Data["Description"]

		if description != "" || true {
			for i := range a.tasks {
				if a.tasks[i].ID == msg.TaskID {
					a.tasks[i].Description = description
					if err := taskstore.UpdateTask(msg.TaskID, map[string]any{
						"description": description,
					}); err != nil {
						fmt.Printf("Error saving task description: %v\n", err)
					}
					break
				}
			}
			a.taskList.UpdateTasks(a.tasks)
		}

	case "edit_project":
		name := msg.Data["Name"]
		description := msg.Data["Description"]

		if project := a.findProject(msg.ProjectID); project != nil {
			for i := range a.projects {
				if a.projects[i].ID == msg.ProjectID {
					a.projects[i].Name = name
					a.projects[i].Description = description
					break
				}
			}
			a.taskList.UpdateProjects(a.projects)
			if err := taskstore.UpdateProject(msg.ProjectID, map[string]any{
				"name":        name,
				"description": description,
			}); err != nil {
				fmt.Printf("Error saving project: %v\n", err)
			}
		}

	case "assign_task_to_project":
		// Assign an existing task to an existing project
		if msg.TaskID > 0 && msg.ProjectID > 0 {
			if err := taskstore.AddTaskToProject(msg.TaskID, msg.ProjectID); err != nil {
				fmt.Printf("Error assigning task to project: %v\n", err)
			} else {
				// Update the in-memory task
				for i := range a.tasks {
					if a.tasks[i].ID == msg.TaskID {
						// Check if task is already in project
						alreadyAssigned := false
						for _, projID := range a.tasks[i].ProjectIDs {
							if projID == msg.ProjectID {
								alreadyAssigned = true
								break
							}
						}
						if !alreadyAssigned {
							a.tasks[i].ProjectIDs = append(a.tasks[i].ProjectIDs, msg.ProjectID)
						}
						break
					}
				}
				a.taskList.UpdateTasks(a.tasks)
			}
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

func (a *App) findProject(id int) *taskstore.Project {
	for i := range a.projects {
		if a.projects[i].ID == id {
			return &a.projects[i]
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

func (a *App) getNextProjectID() int {
	maxID := 0
	for _, project := range a.projects {
		if project.ID > maxID {
			maxID = project.ID
		}
	}
	return maxID + 1
}
