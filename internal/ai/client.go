package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"

	"godo/internal/config"
)

type Client struct {
	cfg          *config.AIConfig
	repoPath     string
	worktreePath string
}

func NewClient(cfg *config.AIConfig) *Client {
	return &Client{cfg: cfg}
}

func (c *Client) SetPaths(repoPath, worktreePath string) {
	c.repoPath = repoPath
	c.worktreePath = worktreePath
}

func (c *Client) Generate(prompt string) (string, error) {
	switch c.cfg.Provider {
	case "opencode":
		return c.generateOpenCode(prompt)
	case "groq":
		return c.generateGroq(prompt)
	case "ollama":
		return c.generateOllama(prompt)
	case "openai":
		return c.generateOpenAI(prompt)
	default:
		return c.generateOpenCode(prompt)
	}
}

func (c *Client) generateOpenCode(prompt string) (string, error) {
	ocPrompt := "Working in: " + c.worktreePath + ". " + prompt + ". " +
		"Output files in format: FILE: /path/file.ext then CONTENT: <contents>. " +
		"If no changes, say EXACTLY: NO_CHANGES_NEEDED"

	cmd := exec.Command("opencode", "--print-logs", "run", "--", ocPrompt)
	cmd.Dir = c.worktreePath

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		errMsg := string(stderr.Bytes())
		if errMsg != "" {
			return "", fmt.Errorf("opencode error: %s", errMsg)
		}
		return "", fmt.Errorf("opencode failed: %w", err)
	}

	return string(output), nil
}

func (c *Client) generateOllama(prompt string) (string, error) {
	url := "http://localhost:11434/api/generate"

	reqBody := map[string]interface{}{
		"model":  c.cfg.Model,
		"prompt": prompt,
		"stream": false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var genResp struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return genResp.Response, nil
}

func (c *Client) generateGroq(prompt string) (string, error) {
	url := "https://api.groq.com/openai/v1/chat/completions"

	type Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	model := c.cfg.Model
	if model == "" {
		model = "llama-3.1-70b-versatile"
	}

	reqBody := map[string]interface{}{
		"model": model,
		"messages": []Message{
			{Role: "user", Content: prompt},
		},
		"temperature": c.cfg.Temperature,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("groq returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var groqResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return groqResp.Choices[0].Message.Content, nil
}

func (c *Client) generateOpenAI(prompt string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	type Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	reqBody := map[string]interface{}{
		"model": c.cfg.Model,
		"messages": []Message{
			{Role: "user", Content: prompt},
		},
		"temperature": c.cfg.Temperature,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var openaiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openaiResp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return openaiResp.Choices[0].Message.Content, nil
}

func BuildPrompt(taskTitle, taskDescription, repoPath string) string {
	var b strings.Builder

	b.WriteString("You are a helpful coding assistant. ")
	b.WriteString("Generate code based on the following task.\n\n")

	b.WriteString("## Task\n")
	b.WriteString("Title: " + taskTitle + "\n\n")

	if taskDescription != "" {
		b.WriteString("Description:\n" + taskDescription + "\n\n")
	}

	b.WriteString("## Instructions\n")
	b.WriteString("1. Create or modify the necessary files to implement this task\n")
	b.WriteString("2. Output ONLY the file contents, no explanations\n")
	b.WriteString("3. Use this format for each file:\n")
	b.WriteString("```\n")
	b.WriteString("FILE: /path/to/file.ext\n")
	b.WriteString("CONTENT:\n")
	b.WriteString("<file contents here>\n")
	b.WriteString("```\n\n")
	b.WriteString("4. If no files need to be created/modified, respond with:\n")
	b.WriteString("NO_CHANGES_NEEDED\n")

	return b.String()
}
