package cmd

import (
	"fmt"

	"godo/internal/taskstore"

	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new task",
	Long:  `Add a new task to the taks list by passing it as a string argument.`,
	Args:  cobra.MinimumNArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		title := args[0]
		description := ""

		if len(args) > 1 {
			description = args[1]
		}

		if err := taskstore.AddTask(title, description); err != nil {
			fmt.Println("Error adding task:", err)
			return
		}

		fmt.Println("add called")
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
