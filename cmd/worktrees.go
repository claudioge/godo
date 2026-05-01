package cmd

import (
	"fmt"
	"godo/internal/git"
	"godo/internal/taskstore"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var worktreesCmd = &cobra.Command{
	Use:   "worktrees",
	Short: "List all active git worktrees",
	Run: func(cmd *cobra.Command, args []string) {
		projects, err := taskstore.GetProjects()
		if err != nil {
			fmt.Printf("Error loading projects: %v\n", err)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "PROJECT\tREPOSITORY\tBRANCH\tPATH")

		found := false
		for _, project := range projects {
			for _, gitLink := range project.GitLinks {
				if gitLink.LocalPath == "" {
					continue
				}

				worktrees, err := git.GetWorktrees(gitLink.LocalPath)
				if err != nil {
					continue
				}

				for _, wt := range worktrees {
					// The main worktree always exists, skip it if it's the same as local path
					if wt.Path == gitLink.LocalPath {
						continue
					}
					
					found = true
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", 
						color.BlueString(project.Name),
						gitLink.Name,
						color.YellowString(wt.Branch),
						wt.Path,
					)
				}
			}
		}

		if !found {
			fmt.Println("No active worktrees found.")
			return
		}

		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(worktreesCmd)
}
