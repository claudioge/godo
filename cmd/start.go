package cmd

import (
	"fmt"
	"godo/internal/taskstore"
	"godo/internal/ui"
	"time"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a task",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		taskId, selectedTask, err := ui.GetTaskID(args)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if selectedTask.Status == taskstore.StatusInProgress {
			fmt.Println("Task is already in progress")
			return
		}
		now := time.Now()
		updates := map[string]any{
			"status":     taskstore.StatusInProgress,
			"started_at": now,
		}
		if err := taskstore.UpdateTask(taskId, updates); err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Task started")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
