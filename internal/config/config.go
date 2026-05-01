package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	AI       AIConfig       `json:"ai"`
	Git      GitConfig      `json:"git"`
	Worktree WorktreeConfig `json:"worktree"`
}

type AIConfig struct {
	Provider    string  `json:"provider"` // "openai", "anthropic", "ollama"
	APIKey      string  `json:"api_key"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
}

type GitConfig struct {
	DefaultBranch string `json:"default_branch"`
}

type WorktreeConfig struct {
	BasePath string `json:"base_path"` // Where to create worktrees
}

func DefaultConfig() *Config {
	return &Config{
		AI: AIConfig{
			Provider:    "opencode",
			Model:       "",
			Temperature: 0.7,
		},
		Git: GitConfig{
			DefaultBranch: "main",
		},
		Worktree: WorktreeConfig{
			BasePath: "",
		},
	}
}

func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "godo")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, "config.json"), nil
}

func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			if err := Save(cfg); err != nil {
				return nil, err
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Apply defaults for empty fields
	if cfg.AI.Provider == "" {
		cfg.AI.Provider = "ollama"
	}
	if cfg.AI.Model == "" {
		cfg.AI.Model = "llama3.2"
	}
	if cfg.Git.DefaultBranch == "" {
		cfg.Git.DefaultBranch = "main"
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
