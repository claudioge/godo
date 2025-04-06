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

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop working on a task",
	Long:  `Sets a task's status to done and calculates the total time spent`,
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
				if task.Status != taskstore.StatusInProgress {
					fmt.Println("Task is not in progress")
					return
				}

				totalTime := task.TotalTime
				if task.StartedAt != nil {
					totalTime += time.Since(*task.StartedAt)
				}

				updates := map[string]any{
					"status":     taskstore.StatusDone,
					"total_time": totalTime,
					"started_at": nil,
				}

				err := taskstore.UpdateTask(task.ID, updates)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				fmt.Printf("Task stopped. Total time: %v\n", totalTime)
				return
			}
		}

		fmt.Println("Task not found")
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
