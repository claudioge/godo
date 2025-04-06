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
			return &TaskStore{Tasks: []Task{}, NextID: 1}, nil
		}
		return nil, fmt.Errorf("could not read tasks file %s: %w", filePath, err)
	}

	if len(data) == 0 {
		return &TaskStore{Tasks: []Task{}, NextID: 1}, nil
	}

	var store TaskStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("could not parse tasks file %s:%w", filePath, err)
	}

	if store.Tasks == nil {
		store.Tasks = []Task{}
	}

	if store.NextID == 0 {
		store.NextID = 1
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
	}

	store.Tasks = append(store.Tasks, task)
	store.NextID++

	return saveTaskStore(store)
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

	return fmt.Errorf("Something went wrong deleting the task")
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
			return saveTaskStore(store)
		}
	}

	return fmt.Errorf("task with ID %d not found", id)
}

func GetTasks() ([]Task, error) {
	store, err := loadTasks()
	if err != nil {
		return nil, err
	}
	return store.Tasks, nil
}
