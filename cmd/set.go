/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"godo/internal/taskstore"
	"strconv"

	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set satus of task",
	Long: `Update the status of a task
  Valid options are:
    - "todo"
	- "in-progress"
	- "done"
	- "paused"
  `,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		status := args[1]

		fmt.Printf("In set, args: %s, %s", id, status)

		intId, err := strconv.Atoi(id)
		if err != nil {
			fmt.Println("Invalid ID")
			return
		}

		// check if status is a valid TaskStatus value
		validStatus, err := taskstore.GetTaskStatus(status)
		if err != nil {
			fmt.Println(err)
			return
		}

		updates := map[string]any{
			"status": validStatus,
		}

		taskstore.UpdateTask(intId, updates)
	},
}

func init() {
	rootCmd.AddCommand(setCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
