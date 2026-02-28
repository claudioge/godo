package taskstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"
)

func getStoragePath() (string, error) {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	configDir := filepath.Join(homeDir, ".godo")
	errMkdir := os.MkdirAll(configDir, 0755)
	if errMkdir != nil {
		return "", fmt.Errorf("failed to create config directory: %w", errMkdir)
	}
	return filepath.Join(configDir, "tasks.json"), nil
}

func loadTasks() (*TaskStore, error) {
	filePath, err := getStoragePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &TaskStore{
				Tasks:         []Task{},
				Projects:      []Project{},
				NextID:        1,
				NextProjectID: 1,
			}, nil
		}
		return nil, fmt.Errorf("could not read tasks file %s: %w", filePath, err)
	}

	if len(data) == 0 {
		return &TaskStore{
			Tasks:         []Task{},
			Projects:      []Project{},
			NextID:        1,
			NextProjectID: 1,
		}, nil
	}

	var store TaskStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("could not parse tasks file %s:%w", filePath, err)
	}

	if store.Tasks == nil {
		store.Tasks = []Task{}
	}

	if store.Projects == nil {
		store.Projects = []Project{}
	}

	if store.NextID == 0 {
		store.NextID = 1
	}

	if store.NextProjectID == 0 {
		store.NextProjectID = 1
	}

	return &store, nil
}

func saveTaskStore(store *TaskStore) error {
	filePath, err := getStoragePath()
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(store)
	if err != nil {
		return fmt.Errorf("could not marshal tasks to JSON: %w", err)
	}

	return os.WriteFile(filePath, jsonData, 0640)
}

// Task operations
func AddTask(title string, description string) error {
	store, err := loadTasks()
	if err != nil {
		return err
	}

	task := Task{
		ID:          store.NextID,
		Title:       title,
		Description: description,
		Status:      StatusTodo,
		CreatedAt:   time.Now(),
		ProjectIDs:  []int{},
		GitLinks:    []GitLink{},
	}

	store.Tasks = append(store.Tasks, task)
	store.NextID++

	return saveTaskStore(store)
}

func GetTasks() ([]Task, error) {
	store, err := loadTasks()
	if err != nil {
		return nil, err
	}
	return store.Tasks, nil
}

func DeleteTask(id int) error {
	store, err := loadTasks()
	if err != nil {
		return err
	}

	for i, task := range store.Tasks {
		if task.ID == id {
			store.Tasks = slices.Delete(store.Tasks, i, i+1)
			return saveTaskStore(store)
		}
	}

	return fmt.Errorf("task with ID %d not found", id)
}

func UpdateTask(id int, updates map[string]any) error {
	store, err := loadTasks()
	if err != nil {
		return err
	}

	for i, task := range store.Tasks {
		if task.ID == id {
			if title, ok := updates["title"].(string); ok {
				store.Tasks[i].Title = title
			}
			if description, ok := updates["description"].(string); ok {
				store.Tasks[i].Description = description
			}
			if status, ok := updates["status"].(TaskStatus); ok {
				store.Tasks[i].Status = status
			}
			if startedAt, ok := updates["started_at"].(time.Time); ok {
				store.Tasks[i].StartedAt = &startedAt
			}

			return saveTaskStore(store)
		}
	}

	return fmt.Errorf("task with ID %d not found", id)
}

func LinkTaskToGit(id int, localPath string, name string, link string, branch string) error {
	store, err := loadTasks()
	if err != nil {
		return err
	}

	gitLink := GitLink{
		LocalPath: localPath,
		Name:      name,
		Link:      link,
		Branch:    branch,
	}

	for i, task := range store.Tasks {
		if task.ID == id {
			store.Tasks[i].GitLinks = append(store.Tasks[i].GitLinks, gitLink)
			return saveTaskStore(store)
		}
	}

	return fmt.Errorf("task with ID %d not found", id)
}

// Project operations
func AddProject(name string, description string) error {
	store, err := loadTasks()
	if err != nil {
		return err
	}

	project := Project{
		ID:          store.NextProjectID,
		Name:        name,
		Description: description,
		GitLinks:    []GitLink{},
		CreatedAt:   time.Now(),
	}

	store.Projects = append(store.Projects, project)
	store.NextProjectID++

	return saveTaskStore(store)
}

func GetProjects() ([]Project, error) {
	store, err := loadTasks()
	if err != nil {
		return nil, err
	}
	return store.Projects, nil
}

func DeleteProject(id int) error {
	store, err := loadTasks()
	if err != nil {
		return err
	}

	// Remove the project
	for i, project := range store.Projects {
		if project.ID == id {
			store.Projects = slices.Delete(store.Projects, i, i+1)

			// Also remove project ID from all tasks that reference it
			for j := range store.Tasks {
				store.Tasks[j].ProjectIDs = slices.DeleteFunc(store.Tasks[j].ProjectIDs, func(projID int) bool {
					return projID == id
				})
			}

			return saveTaskStore(store)
		}
	}

	return fmt.Errorf("project with ID %d not found", id)
}

func UpdateProject(id int, updates map[string]any) error {
	store, err := loadTasks()
	if err != nil {
		return err
	}

	for i, project := range store.Projects {
		if project.ID == id {
			if name, ok := updates["name"].(string); ok {
				store.Projects[i].Name = name
			}
			if description, ok := updates["description"].(string); ok {
				store.Projects[i].Description = description
			}
			if gitLinks, ok := updates["git_links"].([]GitLink); ok {
				store.Projects[i].GitLinks = gitLinks
			}

			return saveTaskStore(store)
		}
	}

	return fmt.Errorf("project with ID %d not found", id)
}

func AddTaskToProject(taskID int, projectID int) error {
	store, err := loadTasks()
	if err != nil {
		return err
	}

	// Find the task
	taskIndex := -1
	for i, task := range store.Tasks {
		if task.ID == taskID {
			taskIndex = i
			break
		}
	}

	if taskIndex == -1 {
		return fmt.Errorf("task with ID %d not found", taskID)
	}

	// Check if project exists
	projectExists := false
	for _, project := range store.Projects {
		if project.ID == projectID {
			projectExists = true
			break
		}
	}

	if !projectExists {
		return fmt.Errorf("project with ID %d not found", projectID)
	}

	// Check if task is already in project
	for _, projID := range store.Tasks[taskIndex].ProjectIDs {
		if projID == projectID {
			return nil // Already in project
		}
	}

	store.Tasks[taskIndex].ProjectIDs = append(store.Tasks[taskIndex].ProjectIDs, projectID)
	return saveTaskStore(store)
}

func RemoveTaskFromProject(taskID int, projectID int) error {
	store, err := loadTasks()
	if err != nil {
		return err
	}

	for i, task := range store.Tasks {
		if task.ID == taskID {
			store.Tasks[i].ProjectIDs = slices.DeleteFunc(store.Tasks[i].ProjectIDs, func(projID int) bool {
				return projID == projectID
			})
			return saveTaskStore(store)
		}
	}

	return fmt.Errorf("task with ID %d not found", taskID)
}

func AddGitLinkToProject(projectID int, gitLink GitLink) error {
	store, err := loadTasks()
	if err != nil {
		return err
	}

	for i, project := range store.Projects {
		if project.ID == projectID {
			store.Projects[i].GitLinks = append(store.Projects[i].GitLinks, gitLink)
			return saveTaskStore(store)
		}
	}

	return fmt.Errorf("project with ID %d not found", projectID)
}

func RemoveGitLinkFromProject(projectID int, gitLinkName string) error {
	store, err := loadTasks()
	if err != nil {
		return err
	}

	for i, project := range store.Projects {
		if project.ID == projectID {
			store.Projects[i].GitLinks = slices.DeleteFunc(store.Projects[i].GitLinks, func(gl GitLink) bool {
				return gl.Name == gitLinkName
			})
			return saveTaskStore(store)
		}
	}

	return fmt.Errorf("project with ID %d not found", projectID)
}

func RemoveGitLinkFromTask(taskID int, gitLinkName string) error {
	store, err := loadTasks()
	if err != nil {
		return err
	}

	for i, task := range store.Tasks {
		if task.ID == taskID {
			store.Tasks[i].GitLinks = slices.DeleteFunc(store.Tasks[i].GitLinks, func(gl GitLink) bool {
				return gl.Name == gitLinkName
			})
			return saveTaskStore(store)
		}
	}

	return fmt.Errorf("task with ID %d not found", taskID)
}
