package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"stackagent/pkg/shell"
)

// Claude API client
type ClaudeClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	model      string
}

// Claude API request/response structures
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []ClaudeMessage `json:"messages"`
	System    string          `json:"system,omitempty"`
}

type ClaudeResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type ClaudeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type ClaudeErrorResponse struct {
	Error ClaudeError `json:"error"`
}

// AI analysis results
type CommandAnalysis struct {
	Summary     string   `json:"summary"`
	KeyFindings []string `json:"key_findings"`
	Suggestions []string `json:"suggestions"`
	Risk        string   `json:"risk"` // "low", "medium", "high"
	TokensUsed  int      `json:"tokens_used"`
	Cost        float64  `json:"cost"`
}

// NewClaudeClient creates a new Claude API client
func NewClaudeClient() (*ClaudeClient, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable is required")
	}

	return &ClaudeClient{
		apiKey:  apiKey,
		baseURL: "https://api.anthropic.com/v1/messages",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		model: "claude-3-5-sonnet-20241022", // Latest Claude 3.5 Sonnet
	}, nil
}

// SetModel allows switching between Claude models
func (c *ClaudeClient) SetModel(model string) {
	c.model = model
}

// GetAvailableModels returns list of available Claude models
func (c *ClaudeClient) GetAvailableModels() []string {
	return []string{
		"claude-3-5-sonnet-20241022", // Latest Claude 3.5 Sonnet (recommended)
		"claude-3-5-haiku-20241022",  // Fast and cheap
		"claude-3-opus-20240229",     // Most capable
	}
}

// EstimateCost estimates the cost of a request
func (c *ClaudeClient) EstimateCost(inputTokens, outputTokens int) float64 {
	// Claude 3.5 Sonnet pricing: $3/1M input, $15/1M output
	inputCost := float64(inputTokens) * 3.00 / 1000000
	outputCost := float64(outputTokens) * 15.00 / 1000000
	return inputCost + outputCost
}

// EstimateTokens roughly estimates token count from text
func (c *ClaudeClient) EstimateTokens(text string) int {
	// Rough estimation: ~4 characters per token
	return len(text) / 4
}

// AnalyzeCommand analyzes a command and its output using Claude
func (c *ClaudeClient) AnalyzeCommand(sm *shell.ShellManager, handleID uint64, question string) (*CommandAnalysis, error) {
	// Get command output
	handle, exists := sm.GetHandle(handleID)
	if !exists {
		return nil, fmt.Errorf("handle %d not found", handleID)
	}

	// Get command info
	stats, err := sm.GetStats(handleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	// Get output content
	output, err := sm.GetTail(handleID, 100) // Last 100 lines
	if err != nil {
		return nil, fmt.Errorf("failed to get output: %w", err)
	}

	// Construct prompt
	systemPrompt := `You are an expert system administrator and DevOps engineer. 
Analyze command outputs and provide insights, security observations, and suggestions.
Always respond in a structured format with clear sections.`

	userPrompt := fmt.Sprintf(`Please analyze this command execution:

Command: %s
Exit Code: %d
Duration: %v
Lines of Output: %d

Output:
%s

User Question: %s

Please provide:
1. A brief summary of what the command did
2. Key findings from the output
3. Any suggestions for improvement or follow-up actions
4. Risk assessment (low/medium/high) for any security or operational concerns

Keep the response concise but informative.`, 
		handle.Command, stats.ExitCode, stats.Duration, stats.LineCount, output, question)

	// Make API call
	response, err := c.makeAPICall(systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}

	// Parse response into structured format
	analysis := &CommandAnalysis{
		Summary:     extractSection(response.Content[0].Text, "summary"),
		KeyFindings: extractList(response.Content[0].Text, "key findings"),
		Suggestions: extractList(response.Content[0].Text, "suggestions"),
		Risk:        extractRisk(response.Content[0].Text),
		TokensUsed:  response.Usage.InputTokens + response.Usage.OutputTokens,
		Cost:        c.EstimateCost(response.Usage.InputTokens, response.Usage.OutputTokens),
	}

	return analysis, nil
}

// AskAboutOutput asks Claude a specific question about command output
func (c *ClaudeClient) AskAboutOutput(sm *shell.ShellManager, handleID uint64, question string) (string, error) {
	// Get command output
	handle, exists := sm.GetHandle(handleID)
	if !exists {
		return "", fmt.Errorf("handle %d not found", handleID)
	}

	output, err := sm.GetTail(handleID, 50)
	if err != nil {
		return "", fmt.Errorf("failed to get output: %w", err)
	}

	systemPrompt := "You are a helpful assistant analyzing command line output. Be concise and specific."
	userPrompt := fmt.Sprintf(`Command: %s
Output: %s

Question: %s`, handle.Command, output, question)

	response, err := c.makeAPICall(systemPrompt, userPrompt)
	if err != nil {
		return "", err
	}

	return response.Content[0].Text, nil
}

// GenerateCommand asks Claude to suggest a command based on a description
func (c *ClaudeClient) GenerateCommand(description string) (string, error) {
	systemPrompt := `You are an expert system administrator. Generate safe, well-explained commands.
Always provide the command and a brief explanation of what it does.
If the request is potentially dangerous, suggest safer alternatives.`

	userPrompt := fmt.Sprintf("Generate a command for: %s", description)

	response, err := c.makeAPICall(systemPrompt, userPrompt)
	if err != nil {
		return "", err
	}

	return response.Content[0].Text, nil
}

// makeAPICall makes the actual HTTP request to Claude API
func (c *ClaudeClient) makeAPICall(system, user string) (*ClaudeResponse, error) {
	request := ClaudeRequest{
		Model:     c.model,
		MaxTokens: 1024,
		System:    system,
		Messages: []ClaudeMessage{
			{Role: "user", Content: user},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp ClaudeErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("Claude API error: %s", errorResp.Error.Message)
	}

	var response ClaudeResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// Helper functions for parsing Claude's response
func extractSection(text, section string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(section)) && strings.Contains(line, ":") {
			// Extract text after the colon
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
			// If no colon, check next line
			if i+1 < len(lines) {
				return strings.TrimSpace(lines[i+1])
			}
		}
	}
	return "No summary available"
}

func extractList(text, section string) []string {
	lines := strings.Split(text, "\n")
	var items []string
	inSection := false
	
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(section)) {
			inSection = true
			continue
		}
		if inSection {
			line = strings.TrimSpace(line)
			if line == "" {
				break
			}
			// Handle different bullet point types
			if strings.HasPrefix(line, "- ") {
				items = append(items, strings.TrimSpace(line[2:]))
			} else if strings.HasPrefix(line, "• ") {
				items = append(items, strings.TrimSpace(line[3:]))
			} else if strings.HasPrefix(line, "* ") {
				items = append(items, strings.TrimSpace(line[2:]))
			} else if len(line) > 0 && !strings.Contains(line, ":") {
				items = append(items, line)
			}
		}
	}
	return items
}

func extractRisk(text string) string {
	text = strings.ToLower(text)
	if strings.Contains(text, "high risk") || strings.Contains(text, "dangerous") {
		return "high"
	}
	if strings.Contains(text, "medium risk") || strings.Contains(text, "caution") {
		return "medium"
	}
	return "low"
} 