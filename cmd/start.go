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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
