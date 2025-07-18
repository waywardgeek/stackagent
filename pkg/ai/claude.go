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
	apiKey      string
	baseURL     string
	httpClient  *http.Client
	model       string
	debugFile   *os.File
	debugEnabled bool
	shellManager *shell.ShellManager
}

// Tool definition for function calling
type Tool struct {
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	InputSchema  InputSchema `json:"input_schema"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// Tool use structures for Claude responses
type ToolUse struct {
	Type  string                 `json:"type"`
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

type ToolResult struct {
	Type       string `json:"type"`
	ToolUseID  string `json:"tool_use_id"`
	Content    string `json:"content"`
	IsError    bool   `json:"is_error,omitempty"`
}

// Claude API request/response structures
type ClaudeMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []ClaudeMessage `json:"messages"`
	Tools     []Tool          `json:"tools,omitempty"`
	System    string          `json:"system,omitempty"`
}

type ClaudeResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type  string                 `json:"type"`
		Text  string                 `json:"text,omitempty"`
		ID    string                 `json:"id,omitempty"`
		Name  string                 `json:"name,omitempty"`
		Input map[string]interface{} `json:"input,omitempty"`
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

// Debug log entry structure
type DebugLogEntry struct {
	Timestamp string      `json:"timestamp"`
	Type      string      `json:"type"` // "request" or "response"
	Method    string      `json:"method"`
	URL       string      `json:"url"`
	Headers   map[string]string `json:"headers,omitempty"`
	Body      interface{} `json:"body,omitempty"`
	StatusCode int        `json:"status_code,omitempty"`
	Error     string      `json:"error,omitempty"`
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
		model:        "claude-sonnet-4-20250514", // Latest Claude Sonnet 4
		shellManager: shell.NewShellManager(),
	}, nil
}

// EnableDebugLogging enables logging of all API calls to a file
func (c *ClaudeClient) EnableDebugLogging(filename string) error {
	if c.debugFile != nil {
		c.debugFile.Close()
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open debug file: %w", err)
	}

	c.debugFile = file
	c.debugEnabled = true

	// Write initial header
	header := DebugLogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Type:      "session_start",
		Method:    "DEBUG",
		URL:       "StackAgent Debug Session Started",
	}
	
	c.writeDebugEntry(header)
	return nil
}

// DisableDebugLogging disables debug logging and closes the file
func (c *ClaudeClient) DisableDebugLogging() error {
	c.debugEnabled = false
	if c.debugFile != nil {
		err := c.debugFile.Close()
		c.debugFile = nil
		return err
	}
	return nil
}

// writeDebugEntry writes a debug entry to the log file
func (c *ClaudeClient) writeDebugEntry(entry DebugLogEntry) {
	if !c.debugEnabled || c.debugFile == nil {
		return
	}

	jsonData, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return
	}

	// Write with separator for readability
	c.debugFile.WriteString("--- DEBUG ENTRY ---\n")
	c.debugFile.Write(jsonData)
	c.debugFile.WriteString("\n\n")
	c.debugFile.Sync() // Ensure data is written immediately
}

// SetModel allows switching between Claude models
func (c *ClaudeClient) SetModel(model string) {
	c.model = model
}

// GetAvailableModels returns list of available Claude models
func (c *ClaudeClient) GetAvailableModels() []string {
	return []string{
		"claude-sonnet-4-20250514",        // Latest Claude Sonnet 4 (recommended)
		"claude-3-5-sonnet-20241022",      // Claude 3.5 Sonnet (previous)
		"claude-3-5-haiku-20241022",       // Fast and cheap
		"claude-3-opus-20240229",          // Most capable Claude 3
	}
}

// EstimateCost estimates the cost of a request
func (c *ClaudeClient) EstimateCost(inputTokens, outputTokens int) float64 {
	// Claude 4 Sonnet pricing: $3/1M input, $15/1M output
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

// Chat sends a simple conversational message to Claude
func (c *ClaudeClient) Chat(message string) (string, error) {
	systemPrompt := `You are StackAgent, a helpful AI coding assistant. You help developers with their code, answer questions, and can execute commands when needed. Be concise but helpful.`
	
	response, err := c.makeAPICall(systemPrompt, message)
	if err != nil {
		return "", err
	}

	return response.Content[0].Text, nil
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

	// Log request if debug enabled
	if c.debugEnabled {
		debugEntry := DebugLogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Type:      "request",
			Method:    "POST",
			URL:       c.baseURL,
			Headers: map[string]string{
				"Content-Type":      "application/json",
				"anthropic-version": "2023-06-01",
				"x-api-key":         "[REDACTED]", // Don't log actual API key
			},
			Body: request,
		}
		c.writeDebugEntry(debugEntry)
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
		// Log error if debug enabled
		if c.debugEnabled {
			debugEntry := DebugLogEntry{
				Timestamp: time.Now().Format(time.RFC3339),
				Type:      "response",
				Method:    "POST",
				URL:       c.baseURL,
				Error:     err.Error(),
			}
			c.writeDebugEntry(debugEntry)
		}
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Log response if debug enabled
	if c.debugEnabled {
		var responseBody interface{}
		json.Unmarshal(body, &responseBody)
		
		debugEntry := DebugLogEntry{
			Timestamp:  time.Now().Format(time.RFC3339),
			Type:       "response",
			Method:     "POST",
			URL:        c.baseURL,
			StatusCode: resp.StatusCode,
			Body:       responseBody,
		}
		c.writeDebugEntry(debugEntry)
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

// debugLog logs debug information to the debug file
func (c *ClaudeClient) debugLog(direction, content string) {
	if c.debugFile != nil {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		c.debugFile.WriteString(fmt.Sprintf("[%s] %s:\n%s\n\n", timestamp, direction, content))
	}
}

// getAvailableTools returns the list of available tools for function calling
func (c *ClaudeClient) getAvailableTools() []Tool {
	return []Tool{
		{
			Name:        "run_with_capture",
			Description: "Execute a shell command and capture its output for analysis. Returns a handle that can be used to query the output.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"command": {
						Type:        "string",
						Description: "The shell command to execute",
					},
				},
				Required: []string{"command"},
			},
		},
	}
}

// executeFunction executes a function call and returns the result
func (c *ClaudeClient) executeFunction(toolUse ToolUse) (string, error) {
	switch toolUse.Name {
	case "run_with_capture":
		command, ok := toolUse.Input["command"].(string)
		if !ok {
			return "", fmt.Errorf("invalid command parameter")
		}

		handle, err := c.shellManager.RunWithCapture(command)
		if err != nil {
			return "", fmt.Errorf("failed to execute command: %w", err)
		}

		// Wait a moment for some output to be captured
		time.Sleep(100 * time.Millisecond)

		// Get the current output
		output, err := c.shellManager.GetTail(handle.ID, 50) // Get last 50 lines
		if err != nil {
			output = "No output captured yet"
		}

		result := fmt.Sprintf("Command executed successfully. Handle ID: %d\n\nOutput:\n%s", handle.ID, output)
		
		if handle.Complete {
			result += fmt.Sprintf("\n\nCommand completed with exit code: %d", handle.ExitCode)
		} else {
			result += "\n\nCommand is still running..."
		}

		return result, nil
	default:
		return "", fmt.Errorf("unknown function: %s", toolUse.Name)
	}
}

// makeRequestWithTools makes an HTTP request with function calling support
func (c *ClaudeClient) makeRequestWithTools(request ClaudeRequest) (*ClaudeResponse, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	if c.debugEnabled {
		c.debugLog("REQUEST", string(requestBody))
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if c.debugEnabled {
		c.debugLog("RESPONSE", string(responseBody))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	var response ClaudeResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// ChatWithTools sends a chat message with function calling support
func (c *ClaudeClient) ChatWithTools(message string) (string, error) {
	tools := c.getAvailableTools()
	
	messages := []ClaudeMessage{
		{
			Role:    "user",
			Content: message,
		},
	}

	for {
		request := ClaudeRequest{
			Model:     c.model,
			MaxTokens: 4000,
			Messages:  messages,
			Tools:     tools,
			System:    "You are StackAgent, a helpful AI coding assistant with access to shell commands through the run_with_capture function. Use this function when users ask you to execute commands, check system status, or perform any tasks that require shell access. Be concise but helpful.",
		}

		response, err := c.makeRequestWithTools(request)
		if err != nil {
			return "", err
		}

		// Check if Claude wants to use a tool
		var toolUses []ToolUse
		var textResponse string

		for _, content := range response.Content {
			if content.Type == "text" {
				textResponse = content.Text
			} else if content.Type == "tool_use" {
				toolUses = append(toolUses, ToolUse{
					Type:  content.Type,
					ID:    content.ID,
					Name:  content.Name,
					Input: content.Input,
				})
			}
		}

		if len(toolUses) == 0 {
			// No tools to execute, return the text response
			return textResponse, nil
		}

		// Execute tools and prepare tool results
		var toolResults []interface{}
		for _, toolUse := range toolUses {
			result, err := c.executeFunction(toolUse)
			if err != nil {
				toolResults = append(toolResults, ToolResult{
					Type:      "tool_result",
					ToolUseID: toolUse.ID,
					Content:   fmt.Sprintf("Error: %s", err.Error()),
					IsError:   true,
				})
			} else {
				toolResults = append(toolResults, ToolResult{
					Type:      "tool_result",
					ToolUseID: toolUse.ID,
					Content:   result,
				})
			}
		}

		// Add Claude's response to the conversation
		messages = append(messages, ClaudeMessage{
			Role:    "assistant",
			Content: response.Content,
		})

		// Add tool results to the conversation
		messages = append(messages, ClaudeMessage{
			Role:    "user",
			Content: toolResults,
		})

		// Continue the conversation to get Claude's final response
	}
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