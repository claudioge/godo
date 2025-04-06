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

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop working on a task",
	Long:  `Sets a task's status to done and calculates the total time spent`,
	Run: func(cmd *cobra.Command, args []string) {
		_, selectedTask, err := ui.GetTaskID(args)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if selectedTask.Status != taskstore.StatusInProgress {
			fmt.Println("Task is not in progress")
			return
		}

		totalTime := selectedTask.TotalTime
		if selectedTask.StartedAt != nil {
			totalTime += time.Since(*selectedTask.StartedAt)
		}

		updates := map[string]any{
			"status":     taskstore.StatusDone,
			"total_time": totalTime,
			"started_at": nil,
		}

		if err := taskstore.UpdateTask(selectedTask.ID, updates); err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Printf("Task stopped. Total time: %v\n", totalTime)

	},
}

func init() {
	rootCmd.AddCommand(stopCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// stopCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// stopCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
