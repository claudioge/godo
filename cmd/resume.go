package cmd

import (
	"fmt"
	"godo/internal/taskstore"
	"godo/internal/ui"
	"time"

	"github.com/spf13/cobra"
)

// resumeCmd represents the resume command
var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume a paused task",
	Long:  `Resume a previously paused task by setting its status back to in-progress and updating the start time.`,
	Run: func(cmd *cobra.Command, args []string) {
		_, selectedTask, err := ui.GetTaskID(args)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if selectedTask.Status != taskstore.StatusPaused {
			fmt.Println("Task must be paused to resume")
			return
		}

		now := time.Now()
		updates := map[string]any{
			"status":     taskstore.StatusInProgress,
			"started_at": now,
		}

		if err := taskstore.UpdateTask(selectedTask.ID, updates); err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Task resumed")

		fmt.Println("Task not found")
	},
}

func init() {
	rootCmd.AddCommand(resumeCmd)
}
