/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"godo/internal/taskstore"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

// pauseCmd represents the pause command
var pauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause the current task",
	Long:  `Pause the currently running task and save the elapsed time`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]

		// Get task ID as int
		taskID, err := strconv.Atoi(id)
		if err != nil {
			fmt.Println("Invalid task ID")
			return
		}

		tasks, err := taskstore.GetTasks()
		if err != nil {
			fmt.Println("Error getting tasks:", err)
			return
		}

		for _, task := range tasks {
			if task.ID == taskID {
				if task.Status != taskstore.StatusInProgress {
					fmt.Println("Task is not in progress")
					return
				}

				// Calculate total time
				elapsed := time.Since(*task.StartedAt)
				totalTime := task.TotalTime + elapsed

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
				return
			}
		}

		fmt.Println("Task not found")
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
