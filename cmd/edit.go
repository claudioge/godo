/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"godo/internal/taskstore"
	"godo/internal/ui"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a task in your preferred editor",
	Long:  `Opens your configured editor (default: vi) to edit the task title and description`,
	Run: func(cmd *cobra.Command, args []string) {
		taskId, selectedTask, err := ui.GetTaskID(args)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// Create temporary file with task JSON
		type editableTask struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}

		editable := editableTask{
			Title:       selectedTask.Title,
			Description: selectedTask.Description,
		}

		tmpFile, err := os.CreateTemp("", "task-*.json")
		if err != nil {
			fmt.Println("Error creating temp file:", err)
			return
		}
		defer os.Remove(tmpFile.Name())

		encoder := json.NewEncoder(tmpFile)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(editable); err != nil {
			fmt.Println("Error writing task to temp file:", err)
			return
		}
		tmpFile.Close()

		// Get editor from env or default to vi
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}

		// Find the full path of the editor
		editorPath, err := exec.LookPath(editor)
		if err != nil {
			fmt.Println("Error finding editor:", err)
			return
		}

		// Start the editor process
		procAttr := &os.ProcAttr{
			Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		}

		process, err := os.StartProcess(editorPath, []string{editor, tmpFile.Name()}, procAttr)
		if err != nil {
			fmt.Println("Error starting editor:", err)
			return
		}

		// Wait for the editor to close
		_, err = process.Wait()
		if err != nil {
			fmt.Println("Error waiting for editor:", err)
			return
		}

		// Read updated content
		content, err := os.ReadFile(tmpFile.Name())
		if err != nil {
			fmt.Println("Error reading updated file:", err)
			return
		}

		var updated editableTask
		if err := json.Unmarshal(content, &updated); err != nil {
			fmt.Println("Error parsing updated task:", err)
			return
		}

		// Update task
		updates := map[string]any{
			"title":       updated.Title,
			"description": updated.Description,
		}

		if err := taskstore.UpdateTask(taskId, updates); err != nil {
			fmt.Println("Error updating task:", err)
			return
		}

		fmt.Println("Task updated successfully")
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// editCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// editCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
