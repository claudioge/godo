package ui

import (
	"fmt"
	"godo/internal/taskstore"
	"strconv"
)

func GetTaskID(args []string) (int, *taskstore.Task, error) {
	tasks, err := taskstore.GetTasks()
	if err != nil {
		return 0, nil, fmt.Errorf("error getting tasks: %w", err)
	}

	if len(tasks) == 0 {
		return 0, nil, fmt.Errorf("no tasks available")
	}

	if len(args) == 0 {
		// No task ID provided, show selection
		selectedTask, err := SelectTask(tasks)
		if err != nil {
			return 0, nil, fmt.Errorf("error selecting task: %w", err)
		}
		if selectedTask == nil {
			return 0, nil, fmt.Errorf("no task selected")
		}
		return selectedTask.ID, selectedTask, nil
	}

	// Task ID provided as argument
	taskID, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, nil, fmt.Errorf("invalid task ID")
	}

	// Find the task with the given ID
	var selectedTask *taskstore.Task
	for i := range tasks {
		if tasks[i].ID == taskID {
			selectedTask = &tasks[i]
			break
		}
	}

	if selectedTask == nil {
		return 0, nil, fmt.Errorf("task with ID %d not found", taskID)
	}

	return taskID, selectedTask, nil
}
