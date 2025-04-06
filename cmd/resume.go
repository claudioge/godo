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

// resumeCmd represents the resume command
var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume a paused task",
	Long:  `Resume a previously paused task by setting its status back to in-progress and updating the start time.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]

		intId, err := strconv.Atoi(id)
		if err != nil {
			fmt.Println("Invalid task ID")
			return
		}

		tasks, err := taskstore.GetTasks()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		for _, task := range tasks {
			if task.ID == intId {
				if task.Status != taskstore.StatusPaused {
					fmt.Println("Task must be paused to resume")
					return
				}

				now := time.Now()
				updates := map[string]any{
					"status":     taskstore.StatusInProgress,
					"started_at": now,
				}

				err := taskstore.UpdateTask(task.ID, updates)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				fmt.Println("Task resumed")
				return
			}
		}

		fmt.Println("Task not found")
	},
}

func init() {
	rootCmd.AddCommand(resumeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// resumeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// resumeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
