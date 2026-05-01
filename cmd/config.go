package cmd

import (
	"fmt"
	"godo/internal/config"
	"strings"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View or update configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		if len(args) == 0 {
			fmt.Println("Current Configuration:")
			fmt.Printf("  AI Provider:    %s\n", cfg.AI.Provider)
			fmt.Printf("  AI Model:       %s\n", cfg.AI.Model)
			fmt.Printf("  AI API Key:     %s\n", maskKey(cfg.AI.APIKey))
			fmt.Printf("  Worktree Base:  %s\n", cfg.Worktree.BasePath)
			return
		}

		if len(args) < 2 {
			fmt.Println("Usage: godo config <key> <value>")
			return
		}

		key := strings.ToLower(args[0])
		value := args[1]

		switch key {
		case "ai.provider":
			cfg.AI.Provider = value
		case "ai.model":
			cfg.AI.Model = value
		case "ai.api_key":
			cfg.AI.APIKey = value
		case "worktree.base_path":
			cfg.Worktree.BasePath = value
		default:
			fmt.Printf("Unknown config key: %s\n", key)
			return
		}

		if err := config.Save(cfg); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
		} else {
			fmt.Printf("Updated %s to %s\n", key, value)
		}
	},
}

func maskKey(key string) string {
	if len(key) < 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func init() {
	rootCmd.AddCommand(configCmd)
}
