/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"sort"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"godo/internal/taskstore"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all current tasks in a nice and clear way",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		tasks, err := taskstore.GetTasks()
		if err != nil {
			fmt.Printf("\n❌ %s Error listing tasks: %v\n", color.RedString("ERROR:"), err)
			return
		}

		fmt.Printf("\n%s %s\n", "📋", color.BlueString("Your Tasks"))
		fmt.Printf("%s\n", color.HiBlackString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))

		if len(tasks) == 0 {
			fmt.Printf("\n📭 %s\n\n", color.YellowString("No tasks found!"))
			return
		}

		// Sort tasks by status
		statusOrder := map[taskstore.TaskStatus]int{
			taskstore.StatusInProgress: 1,
			taskstore.StatusTodo:       2,
			taskstore.StatusPaused:     3,
			taskstore.StatusDone:       4,
		}

		sortedTasks := make([]taskstore.Task, len(tasks))
		copy(sortedTasks, tasks)
		sort.Slice(sortedTasks, func(i, j int) bool {
			return statusOrder[sortedTasks[i].Status] < statusOrder[sortedTasks[j].Status]
		})

		for _, task := range sortedTasks {
			var statusEmoji, statusColor string
			var totalTime time.Duration

			switch task.Status {
			case taskstore.StatusTodo:
				statusEmoji = "⭕"
				statusColor = color.WhiteString(string(task.Status))
			case taskstore.StatusInProgress:
				statusEmoji = "🔄"
				statusColor = color.BlueString(string(task.Status))
				if task.StartedAt != nil {
					totalTime = time.Since(*task.StartedAt)
				}
			case taskstore.StatusDone:
				statusEmoji = "✅"
				statusColor = color.GreenString(string(task.Status))
			case taskstore.StatusPaused:
				statusEmoji = "⏸️"
				statusColor = color.YellowString(string(task.Status))
			default:
				statusEmoji = "❓"
				statusColor = color.RedString("unknown")
			}

			// Format ID
			idStr := color.HiBlackString(fmt.Sprintf("#%d", task.ID))

			// Format title with optional description
			titleStr := color.HiWhiteString(task.Title)
			if task.Description != "" {
				titleStr += color.HiBlackString(" - " + task.Description)
			}

			// Format time info
			timeInfo := ""
			if task.Status == taskstore.StatusInProgress && task.StartedAt != nil {
				timeInfo = color.CyanString(" ⏱ %s", formatDuration(totalTime))
			} else if task.TotalTime > 0 {
				timeInfo = color.HiBlackString(" ⌛ %s", formatDuration(task.TotalTime))
			}

			// Print the task line
			fmt.Printf("%s %s [%s]%s\n   %s %s\n",
				statusEmoji,
				idStr,
				statusColor,
				timeInfo,
				"└─",
				titleStr,
			)
		}
		fmt.Println()
	},
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
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
