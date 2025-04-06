/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"godo/internal/taskstore"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		tasks, err := taskstore.GetTasks()
		if err != nil {
			fmt.Println("❌ Error listing tasks:", err)
			return
		}

		fmt.Println("\n📋 Your Tasks:")
		fmt.Println("━━━━━━━━━━━━━━━━")
		if len(tasks) == 0 {
			fmt.Println("📭 No tasks found!")
		}
		for _, task := range tasks {
			var statusEmoji string
			var totalTime string
			switch task.Status {
			case taskstore.StatusTodo:
				statusEmoji = "⭕"
			case taskstore.StatusInProgress:
				statusEmoji = "🔄"
				totalTime = time.Now().Sub(*task.StartedAt).String()
			case taskstore.StatusDone:
				statusEmoji = "✅"
			case taskstore.StatusPaused:
				statusEmoji = "⏸️"
			default:
				statusEmoji = "❓"
			}
			timeInfo := ""
			if task.Status == taskstore.StatusInProgress && task.StartedAt != nil {
				timeInfo = fmt.Sprintf(" | Time Spent: %s", totalTime)
			}
			fmt.Printf("%s #%d: %s%s\n", statusEmoji, task.ID, task.Title, timeInfo)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
