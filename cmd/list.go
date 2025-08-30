package cmd

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"godo/internal/taskstore"
	"godo/internal/ui"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all current tasks in a nice and clear way",
	Long:  `Lists all current tasks. Use -i for interactive mode with vim bindings.`,
	Run: func(cmd *cobra.Command, args []string) {
		interactive, _ := cmd.Flags().GetBool("interactive")

		if interactive {
			err := ui.RunInteractiveList()
			if err != nil {
				fmt.Printf("❌ %s Error running interactive list: %v\n", color.RedString("ERROR:"), err)
			}
			return
		}
		tasks, err := taskstore.GetTasks()
		if err != nil {
			fmt.Printf("❌ %s Error listing tasks: %v\n", color.RedString("ERROR:"), err)
			return
		}

		fmt.Printf("%s %s\n", "📋", color.BlueString("Your Tasks"))

		if len(tasks) == 0 {
			fmt.Printf("📭 %s\n", color.YellowString("No tasks found!"))
			return
		}

		// Define status sections in order
		statusSections := []struct {
			status taskstore.TaskStatus
			emoji  string
			color  func(string, ...any) string
		}{
			{taskstore.StatusInProgress, "🔄", color.BlueString},
			{taskstore.StatusTodo, "⭕", color.WhiteString},
			{taskstore.StatusPaused, "⏸️", color.YellowString},
			{taskstore.StatusDone, "✅", color.GreenString},
		}

		// Group tasks by status
		tasksByStatus := make(map[taskstore.TaskStatus][]taskstore.Task)
		for _, task := range tasks {
			tasksByStatus[task.Status] = append(tasksByStatus[task.Status], task)
		}

		// Print each section
		for _, section := range statusSections {
			if tasks := tasksByStatus[section.status]; len(tasks) > 0 {
				fmt.Printf("\n%s %s\n", section.emoji, color.New(color.Bold).SprintFunc()(section.color(string(section.status))))

				for _, task := range tasks {
					timeInfo := ""
					if task.Status == taskstore.StatusInProgress && task.StartedAt != nil {
						timeInfo = color.CyanString(" ⏱ %s", formatDuration(time.Since(*task.StartedAt)))
					} else if task.TotalTime > 0 {
						timeInfo = color.HiBlackString(" ⌛ %s", formatDuration(task.TotalTime))
					}

					fmt.Printf("  #%d %s%s\n", task.ID, color.HiWhiteString(task.Title), timeInfo)
					if task.Description != "" {
						fmt.Printf("     %s\n", color.HiBlackString(task.Description))
					}
				}
			}
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

	// Add interactive flag
	listCmd.Flags().BoolP("interactive", "i", false, "Run in interactive mode with vim bindings")
}
