package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Repo struct {
	Name   string
	Path   string
	URL    string
	Branch string
}

func ScanGitHubFolder() ([]Repo, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	githubPaths := []string{
		filepath.Join(homeDir, "Documents", "GitHub"),
		filepath.Join(homeDir, "github"),
		filepath.Join(homeDir, "GitHub"),
	}

	var repos []Repo

	for _, basePath := range githubPaths {
		entries, err := os.ReadDir(basePath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			repoPath := filepath.Join(basePath, entry.Name())
			if !isGitRepo(repoPath) {
				continue
			}

			remoteURL, err := getRemoteOrigin(repoPath)
			if err != nil {
				remoteURL = ""
			}

			currentBranch, _ := getCurrentBranch(repoPath)

			repos = append(repos, Repo{
				Name:   entry.Name(),
				Path:   repoPath,
				URL:    remoteURL,
				Branch: currentBranch,
			})
		}
	}

	return repos, nil
}

func isGitRepo(path string) bool {
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func getRemoteOrigin(path string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getCurrentBranch(path string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return "main", err
	}
	return strings.TrimSpace(string(output)), nil
}

func CreateBranch(repoPath, branchName string) error {
	cmd := exec.Command("git", "checkout", "-b", branchName)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}
	return nil
}

func BranchExists(repoPath, branchName string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", branchName)
	cmd.Dir = repoPath
	return cmd.Run() == nil
}
