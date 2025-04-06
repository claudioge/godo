/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"godo/internal/taskstore"
	"godo/internal/ui"
	"time"

	"github.com/spf13/cobra"
)

// pauseCmd represents the pause command
var pauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause the current task",
	Long:  `Pause the currently running task and save the elapsed time`,
	Run: func(cmd *cobra.Command, args []string) {
		taskID, selectedTask, err := ui.GetTaskID(args)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if selectedTask.Status != taskstore.StatusInProgress {
			fmt.Println("Task is not in progress")
			return
		}

		// Calculate total time
		elapsed := time.Since(*selectedTask.StartedAt)
		totalTime := selectedTask.TotalTime + elapsed

		updates := map[string]any{
			"status":     taskstore.StatusPaused,
			"total_time": totalTime,
			"started_at": nil,
		}

		if err := taskstore.UpdateTask(taskID, updates); err != nil {
			fmt.Println("Error updating task:", err)
			return
		}

		fmt.Println("Task paused. Total time:", totalTime.Round(time.Second))

	},
}

func init() {
	rootCmd.AddCommand(pauseCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pauseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pauseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
