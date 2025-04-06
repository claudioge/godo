package ui

import (
	"godo/internal/taskstore"

	"github.com/manifoldco/promptui"
)

func SelectTask(tasks []taskstore.Task) (*taskstore.Task, error) {
	if len(tasks) == 0 {
		return nil, nil
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "➡️ {{ .Title | cyan }} ({{ .Status }})",
		Inactive: "  {{ .Title }} ({{ .Status }})",
		Selected: "✅ {{ .Title }}",
	}

	prompt := promptui.Select{
		Label:     "Select Task",
		Items:     tasks,
		Templates: templates,
	}

	index, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return &tasks[index], nil
}
