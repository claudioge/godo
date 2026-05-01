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

type Provider interface {
	Generate(prompt string, worktreePath string) (string, error)
}

type Client struct {
	cfg          *config.AIConfig
	repoPath     string
	worktreePath string
	provider     Provider
}

func NewClient(cfg *config.AIConfig) *Client {
	c := &Client{cfg: cfg}
	c.initProvider()
	return c
}

func (c *Client) initProvider() {
	switch c.cfg.Provider {
	case "opencode":
		c.provider = &OpenCodeProvider{}
	case "gemini":
		c.provider = &GeminiCLIProvider{}
	case "claudecode":
		c.provider = &ClaudeCodeProvider{}
	case "codex":
		c.provider = &CodexProvider{}
	case "groq":
		c.provider = &GroqProvider{cfg: c.cfg}
	case "ollama":
		c.provider = &OllamaProvider{cfg: c.cfg}
	case "openai":
		c.provider = &OpenAIProvider{cfg: c.cfg}
	default:
		c.provider = &OpenCodeProvider{}
	}
}

func (c *Client) SetPaths(repoPath, worktreePath string) {
	c.repoPath = repoPath
	c.worktreePath = worktreePath
}

func (c *Client) Generate(prompt string) (string, error) {
	if c.provider == nil {
		return "", fmt.Errorf("AI provider not initialized")
	}
	return c.provider.Generate(prompt, c.worktreePath)
}

// OpenCodeProvider uses the opencode CLI
type OpenCodeProvider struct{}

func (p *OpenCodeProvider) Generate(prompt string, worktreePath string) (string, error) {
	ocPrompt := "Working in: " + worktreePath + ". " + prompt + ". " +
		"Output files in format: FILE: /path/file.ext then CONTENT: <contents>. " +
		"If no changes, say EXACTLY: NO_CHANGES_NEEDED"

	cmd := exec.Command("opencode", "--print-logs", "run", "--", ocPrompt)
	cmd.Dir = worktreePath

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

// GeminiCLIProvider uses the geminicli CLI
type GeminiCLIProvider struct{}

func (p *GeminiCLIProvider) Generate(prompt string, worktreePath string) (string, error) {
	cmd := exec.Command("gemini", "run", prompt)
	cmd.Dir = worktreePath

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		errMsg := string(stderr.Bytes())
		if errMsg != "" {
			return "", fmt.Errorf("gemini error: %s", errMsg)
		}
		return "", fmt.Errorf("gemini failed: %w", err)
	}

	return string(output), nil
}

// ClaudeCodeProvider uses the claudecode CLI
type ClaudeCodeProvider struct{}

func (p *ClaudeCodeProvider) Generate(prompt string, worktreePath string) (string, error) {
	cmd := exec.Command("claudecode", "run", prompt)
	cmd.Dir = worktreePath

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		errMsg := string(stderr.Bytes())
		if errMsg != "" {
			return "", fmt.Errorf("claudecode error: %s", errMsg)
		}
		return "", fmt.Errorf("claudecode failed: %w", err)
	}

	return string(output), nil
}

// CodexProvider uses the codex CLI
type CodexProvider struct{}

func (p *CodexProvider) Generate(prompt string, worktreePath string) (string, error) {
	cmd := exec.Command("codex", "run", prompt)
	cmd.Dir = worktreePath

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		errMsg := string(stderr.Bytes())
		if errMsg != "" {
			return "", fmt.Errorf("codex error: %s", errMsg)
		}
		return "", fmt.Errorf("codex failed: %w", err)
	}

	return string(output), nil
}

// OllamaProvider uses the Ollama API
type OllamaProvider struct {
	cfg *config.AIConfig
}

func (p *OllamaProvider) Generate(prompt string, worktreePath string) (string, error) {
	url := "http://localhost:11434/api/generate"

	reqBody := map[string]interface{}{
		"model":  p.cfg.Model,
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

// GroqProvider uses the Groq API
type GroqProvider struct {
	cfg *config.AIConfig
}

func (p *GroqProvider) Generate(prompt string, worktreePath string) (string, error) {
	url := "https://api.groq.com/openai/v1/chat/completions"

	type Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	model := p.cfg.Model
	if model == "" {
		model = "llama-3.1-70b-versatile"
	}

	reqBody := map[string]interface{}{
		"model": model,
		"messages": []Message{
			{Role: "user", Content: prompt},
		},
		"temperature": p.cfg.Temperature,
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
	req.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)

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

// OpenAIProvider uses the OpenAI API
type OpenAIProvider struct {
	cfg *config.AIConfig
}

func (p *OpenAIProvider) Generate(prompt string, worktreePath string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	type Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	reqBody := map[string]interface{}{
		"model": p.cfg.Model,
		"messages": []Message{
			{Role: "user", Content: prompt},
		},
		"temperature": p.cfg.Temperature,
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
	req.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)

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
	b.WriteString("Generate code based on the following task. ")
	b.WriteString("Respond ONLY with the file content blocks.\n\n")

	b.WriteString("## Task\n")
	b.WriteString("Title: " + taskTitle + "\n")

	if taskDescription != "" {
		b.WriteString("Description: " + taskDescription + "\n")
	}

	b.WriteString("\n## Instructions\n")
	b.WriteString("1. Implement the task by creating or modifying files.\n")
	b.WriteString("2. Use this EXACT format for each file:\n")
	b.WriteString("```\n")
	b.WriteString("FILE: path/to/file.ext\n")
	b.WriteString("CONTENT:\n")
	b.WriteString("<complete file content>\n")
	b.WriteString("```\n")
	b.WriteString("3. If no changes are needed, respond with EXACTLY: NO_CHANGES_NEEDED\n")
	b.WriteString("4. Output ONLY the code blocks, no explanations, no chat.\n")

	return b.String()
}
