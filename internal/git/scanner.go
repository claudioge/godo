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

type Worktree struct {
	Path   string
	Branch string
}

func CreateWorktree(repoPath, branchName, worktreePath string) error {
	if worktreePath == "" {
		worktreePath = filepath.Join(filepath.Dir(repoPath), branchName)
	}

	if BranchExists(repoPath, branchName) {
		if _, err := os.Stat(worktreePath); err == nil {
			return nil
		}
		return fmt.Errorf("branch %q already exists but worktree does not", branchName)
	}

	if _, err := os.Stat(worktreePath); err == nil {
		return fmt.Errorf("worktree path %q already exists", worktreePath)
	}

	cmd := exec.Command("git", "worktree", "add", worktreePath, "-b", branchName)
	cmd.Dir = repoPath
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}
	return nil
}

func GetWorktrees(repoPath string) ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var worktrees []Worktree
	lines := strings.Split(string(output), "\n")
	var current Worktree

	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			if current.Path != "" {
				worktrees = append(worktrees, current)
			}
			current = Worktree{Path: strings.TrimPrefix(line, "worktree ")}
		}
		if strings.HasPrefix(line, "branch refs/heads/") {
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		}
	}
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees, nil
}

func WriteFile(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

type FileChange struct {
	Path    string
	Content string
	Action  string // "create", "modify", "delete"
}

func ParseAIResponse(response string) ([]FileChange, error) {
	var changes []FileChange

	sections := strings.Split(response, "```")
	for i := 1; i < len(sections); i += 2 {
		if i+1 >= len(sections) {
			break
		}

		header := strings.TrimSpace(sections[i])
		content := strings.TrimSpace(sections[i+1])

		lines := strings.SplitN(header, "\n", 2)
		if len(lines) < 2 {
			continue
		}

		pathLine := strings.TrimPrefix(lines[0], "FILE:")
		path := strings.TrimSpace(pathLine)

		if strings.HasPrefix(content, "CONTENT:") {
			content = strings.TrimPrefix(content, "CONTENT:")
			content = strings.TrimSpace(content)
		}

		if path != "" {
			changes = append(changes, FileChange{
				Path:    path,
				Content: content,
				Action:  "create",
			})
		}
	}

	return changes, nil
}

func ApplyChanges(repoPath string, changes []FileChange) error {
	for _, change := range changes {
		fullPath := filepath.Join(repoPath, change.Path)
		if err := WriteFile(fullPath, change.Content); err != nil {
			return fmt.Errorf("failed to write %s: %w", change.Path, err)
		}
	}
	return nil
}

func GetDefaultBranch(repoPath string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD@{upstream}")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "main"
	}
	branch := strings.TrimSpace(string(output))
	if strings.Contains(branch, "/") {
		parts := strings.Split(branch, "/")
		return parts[len(parts)-1]
	}
	return "main"
}

func CreateBackup(repoPath, backupName string, files []string) error {
	backupDir := filepath.Join(os.TempDir(), "godo-backups", backupName)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup dir: %w", err)
	}

	for _, file := range files {
		src := filepath.Join(repoPath, file)
		dst := filepath.Join(backupDir, file)

		if _, err := os.Stat(src); err == nil {
			data, err := os.ReadFile(src)
			if err != nil {
				continue
			}

			if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
				continue
			}

			if err := os.WriteFile(dst, data, 0644); err != nil {
				continue
			}
		}
	}

	return nil
}

func RestoreBackup(repoPath, backupName string) error {
	backupDir := filepath.Join(os.TempDir(), "godo-backups", backupName)

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("no backup found: %w", err)
	}

	for _, entry := range entries {
		src := filepath.Join(backupDir, entry.Name())
		dst := filepath.Join(repoPath, entry.Name())

		data, err := os.ReadFile(src)
		if err != nil {
			continue
		}

		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			continue
		}

		if err := os.WriteFile(dst, data, 0644); err != nil {
			continue
		}
	}

	return nil
}

func DeleteWorktree(repoPath, worktreePath string) error {
	cmd := exec.Command("git", "worktree", "remove", "--force", worktreePath)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}
	return nil
}

func DeleteBranch(repoPath, branchName string) error {
	cmd := exec.Command("git", "branch", "-D", branchName)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}
	return nil
}
