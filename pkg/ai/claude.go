package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	debugCallback func(string, interface{}) // Add callback for debug information
}

// Tool definition for function calling
type Tool struct {
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	InputSchema  InputSchema   `json:"input_schema"`
	CacheControl *CacheControl `json:"cache_control,omitempty"`
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
	Type  string                 `json:"type"`
	ToolUseID string                 `json:"tool_use_id"`
	Content string                 `json:"content"`
	IsError bool                   `json:"is_error,omitempty"`
}

// Cache control for prompt caching
type CacheControl struct {
	Type string `json:"type"`
}

// Content block with cache control support
type ContentBlock struct {
	Type         string        `json:"type"`
	Text         string        `json:"text,omitempty"`
	CacheControl *CacheControl `json:"cache_control,omitempty"`
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
	System    interface{}     `json:"system,omitempty"` // Can be string or []ContentBlock
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
		InputTokens              int `json:"input_tokens"`
		OutputTokens             int `json:"output_tokens"`
		CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
		CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	} `json:"usage"`
}

type ClaudeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type ClaudeErrorResponse struct {
	Error ClaudeError `json:"error"`
}

// Cost calculation structures
type TokenCost struct {
	InputTokens              int     `json:"input_tokens"`
	OutputTokens             int     `json:"output_tokens"`
	CacheCreationInputTokens int     `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int     `json:"cache_read_input_tokens"`
	TotalCost                float64 `json:"total_cost"`
	CostBreakdown            struct {
		InputCost        float64 `json:"input_cost"`
		OutputCost       float64 `json:"output_cost"`
		CacheWriteCost   float64 `json:"cache_write_cost"`
		CacheReadCost    float64 `json:"cache_read_cost"`
		CacheSavings     float64 `json:"cache_savings"`
	} `json:"cost_breakdown"`
}

// Model pricing (per million tokens)
func getModelPricing(model string) (inputPrice, outputPrice, cacheWritePrice, cacheReadPrice float64) {
	switch model {
	case "claude-3-5-sonnet-20241022":
		return 3.0, 15.0, 3.75, 0.30  // Sonnet 3.5 pricing
	case "claude-3-opus-20240229":
		return 15.0, 75.0, 18.75, 1.50  // Opus 3 pricing
	case "claude-3-haiku-20240307":
		return 0.25, 1.25, 0.30, 0.03  // Haiku 3 pricing
	default:
		// Default to Sonnet 3.5 pricing
		return 3.0, 15.0, 3.75, 0.30
	}
}

// Calculate cost for a response
func (c *ClaudeClient) CalculateCost(response *ClaudeResponse) TokenCost {
	inputPrice, outputPrice, cacheWritePrice, cacheReadPrice := getModelPricing(response.Model)
	
	// Convert to per-token pricing (divide by 1,000,000)
	inputPrice /= 1000000
	outputPrice /= 1000000
	cacheWritePrice /= 1000000
	cacheReadPrice /= 1000000
	
	cost := TokenCost{
		InputTokens:              response.Usage.InputTokens,
		OutputTokens:             response.Usage.OutputTokens,
		CacheCreationInputTokens: response.Usage.CacheCreationInputTokens,
		CacheReadInputTokens:     response.Usage.CacheReadInputTokens,
	}
	
	// Calculate costs
	cost.CostBreakdown.InputCost = float64(response.Usage.InputTokens) * inputPrice
	cost.CostBreakdown.OutputCost = float64(response.Usage.OutputTokens) * outputPrice
	cost.CostBreakdown.CacheWriteCost = float64(response.Usage.CacheCreationInputTokens) * cacheWritePrice
	cost.CostBreakdown.CacheReadCost = float64(response.Usage.CacheReadInputTokens) * cacheReadPrice
	
	// Calculate potential savings (what cache reads would have cost at full price)
	cost.CostBreakdown.CacheSavings = float64(response.Usage.CacheReadInputTokens) * (inputPrice - cacheReadPrice)
	
	cost.TotalCost = cost.CostBreakdown.InputCost + cost.CostBreakdown.OutputCost + 
		cost.CostBreakdown.CacheWriteCost + cost.CostBreakdown.CacheReadCost
	
	return cost
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

// SetDebugCallback sets a callback function for debug information
func (c *ClaudeClient) SetDebugCallback(callback func(string, interface{})) {
	c.debugCallback = callback
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
		{
			Name:        "read_file",
			Description: "Read the contents of a file. Much more efficient than using cat command for file reading.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"file_path": {
						Type:        "string",
						Description: "The path to the file to read",
					},
					"max_lines": {
						Type:        "integer",
						Description: "Maximum number of lines to read (optional, default: all)",
					},
				},
				Required: []string{"file_path"},
			},
		},
		{
			Name:        "write_file",
			Description: "Write content to a file, creating it if it doesn't exist. Much more efficient than using echo or tee commands.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"file_path": {
						Type:        "string",
						Description: "The path to the file to write",
					},
					"content": {
						Type:        "string",
						Description: "The content to write to the file",
					},
					"append": {
						Type:        "boolean",
						Description: "Whether to append to the file instead of overwriting (optional, default: false)",
					},
				},
				Required: []string{"file_path", "content"},
			},
		},
		{
			Name:        "edit_file",
			Description: "Make specific edits to a file using find and replace operations. Much more efficient than reading, editing, and writing back entire files.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"file_path": {
						Type:        "string",
						Description: "The path to the file to edit",
					},
					"find": {
						Type:        "string",
						Description: "The text to find and replace",
					},
					"replace": {
						Type:        "string",
						Description: "The replacement text",
					},
					"all_occurrences": {
						Type:        "boolean",
						Description: "Whether to replace all occurrences (default: false, replaces only first)",
					},
				},
				Required: []string{"file_path", "find", "replace"},
			},
		},
		{
			Name:        "search_in_file",
			Description: "Search for patterns in a file and return matching lines with context. More efficient than grep for simple searches.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"file_path": {
						Type:        "string",
						Description: "The path to the file to search",
					},
					"pattern": {
						Type:        "string",
						Description: "The text pattern to search for",
					},
					"context_lines": {
						Type:        "integer",
						Description: "Number of context lines to show around matches (optional, default: 2)",
					},
				},
				Required: []string{"file_path", "pattern"},
			},
		},
		{
			Name:        "list_directory",
			Description: "List directory contents with filtering options. More efficient than ls with complex filtering.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"directory_path": {
						Type:        "string",
						Description: "The path to the directory to list",
					},
					"file_extension": {
						Type:        "string",
						Description: "Filter by file extension (optional, e.g., '.go', '.txt')",
					},
					"show_hidden": {
						Type:        "boolean",
						Description: "Whether to show hidden files (optional, default: false)",
					},
					"recursive": {
						Type:        "boolean",
						Description: "Whether to list recursively (optional, default: false)",
					},
				},
				Required: []string{"directory_path"},
			},
			CacheControl: &CacheControl{Type: "ephemeral"}, // Cache all tool definitions
		},
	}
}

// ExecuteFunction executes a function call and returns the result
func (c *ClaudeClient) ExecuteFunction(toolUse ToolUse) (string, error) {
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

	case "read_file":
		filePath, ok := toolUse.Input["file_path"].(string)
		if !ok {
			return "", fmt.Errorf("invalid file_path parameter")
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}

		lines := strings.Split(string(content), "\n")
		
		// Check if max_lines is specified
		if maxLinesVal, exists := toolUse.Input["max_lines"]; exists {
			if maxLinesFloat, ok := maxLinesVal.(float64); ok {
				maxLines := int(maxLinesFloat)
				if maxLines > 0 && maxLines < len(lines) {
					lines = lines[:maxLines]
					return fmt.Sprintf("File: %s (showing first %d lines)\n\n%s", filePath, maxLines, strings.Join(lines, "\n")), nil
				}
			}
		}

		return fmt.Sprintf("File: %s (%d lines)\n\n%s", filePath, len(lines), string(content)), nil

	case "write_file":
		filePath, ok := toolUse.Input["file_path"].(string)
		if !ok {
			return "", fmt.Errorf("invalid file_path parameter")
		}

		content, ok := toolUse.Input["content"].(string)
		if !ok {
			return "", fmt.Errorf("invalid content parameter")
		}

		// Check if append mode is specified
		append := false
		if appendVal, exists := toolUse.Input["append"]; exists {
			if appendBool, ok := appendVal.(bool); ok {
				append = appendBool
			}
		}

		var err error
		if append {
			file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return "", fmt.Errorf("failed to open file for append: %w", err)
			}
			defer file.Close()
			
			_, err = file.WriteString(content)
			if err != nil {
				return "", fmt.Errorf("failed to append to file: %w", err)
			}
			
			return fmt.Sprintf("Successfully appended %d characters to %s", len(content), filePath), nil
		} else {
			err = os.WriteFile(filePath, []byte(content), 0644)
			if err != nil {
				return "", fmt.Errorf("failed to write file: %w", err)
			}
			
			return fmt.Sprintf("Successfully wrote %d characters to %s", len(content), filePath), nil
		}

	case "edit_file":
		filePath, ok := toolUse.Input["file_path"].(string)
		if !ok {
			return "", fmt.Errorf("invalid file_path parameter")
		}

		find, ok := toolUse.Input["find"].(string)
		if !ok {
			return "", fmt.Errorf("invalid find parameter")
		}

		replace, ok := toolUse.Input["replace"].(string)
		if !ok {
			return "", fmt.Errorf("invalid replace parameter")
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}

		contentStr := string(content)
		
		// Check if all_occurrences is specified
		allOccurrences := false
		if allOccVal, exists := toolUse.Input["all_occurrences"]; exists {
			if allOccBool, ok := allOccVal.(bool); ok {
				allOccurrences = allOccBool
			}
		}

		var newContent string
		var count int
		
		if allOccurrences {
			newContent = strings.ReplaceAll(contentStr, find, replace)
			count = strings.Count(contentStr, find)
		} else {
			newContent = strings.Replace(contentStr, find, replace, 1)
			if strings.Contains(contentStr, find) {
				count = 1
			}
		}

		if count == 0 {
			return fmt.Sprintf("No occurrences of '%s' found in %s", find, filePath), nil
		}

		err = os.WriteFile(filePath, []byte(newContent), 0644)
		if err != nil {
			return "", fmt.Errorf("failed to write modified file: %w", err)
		}

		return fmt.Sprintf("Successfully replaced %d occurrence(s) of '%s' with '%s' in %s", count, find, replace, filePath), nil

	case "search_in_file":
		filePath, ok := toolUse.Input["file_path"].(string)
		if !ok {
			return "", fmt.Errorf("invalid file_path parameter")
		}

		pattern, ok := toolUse.Input["pattern"].(string)
		if !ok {
			return "", fmt.Errorf("invalid pattern parameter")
		}

		file, err := os.Open(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		contextLines := 2
		if contextVal, exists := toolUse.Input["context_lines"]; exists {
			if contextFloat, ok := contextVal.(float64); ok {
				contextLines = int(contextFloat)
			}
		}

		var lines []string
		var matches []string
		scanner := bufio.NewScanner(file)
		lineNum := 0
		
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			lines = append(lines, line)
			
			if strings.Contains(line, pattern) {
				start := max(0, lineNum-contextLines-1)
				end := min(len(lines), lineNum+contextLines)
				
				matchResult := fmt.Sprintf("Line %d: %s", lineNum, line)
				if contextLines > 0 {
					matchResult += "\nContext:"
					for i := start; i < end; i++ {
						if i == lineNum-1 {
							matchResult += fmt.Sprintf("  > %d: %s", i+1, lines[i])
						} else {
							matchResult += fmt.Sprintf("    %d: %s", i+1, lines[i])
						}
						if i < end-1 {
							matchResult += "\n"
						}
					}
				}
				matches = append(matches, matchResult)
			}
		}

		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("error reading file: %w", err)
		}

		if len(matches) == 0 {
			return fmt.Sprintf("No matches found for pattern '%s' in %s", pattern, filePath), nil
		}

		return fmt.Sprintf("Found %d match(es) for pattern '%s' in %s:\n\n%s", len(matches), pattern, filePath, strings.Join(matches, "\n\n")), nil

	case "list_directory":
		dirPath, ok := toolUse.Input["directory_path"].(string)
		if !ok {
			return "", fmt.Errorf("invalid directory_path parameter")
		}

		// Get optional parameters
		var fileExt string
		if extVal, exists := toolUse.Input["file_extension"]; exists {
			if extStr, ok := extVal.(string); ok {
				fileExt = extStr
			}
		}

		showHidden := false
		if hiddenVal, exists := toolUse.Input["show_hidden"]; exists {
			if hiddenBool, ok := hiddenVal.(bool); ok {
				showHidden = hiddenBool
			}
		}

		recursive := false
		if recVal, exists := toolUse.Input["recursive"]; exists {
			if recBool, ok := recVal.(bool); ok {
				recursive = recBool
			}
		}

		var files []string
		var err error

		if recursive {
			err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				// Skip hidden files if not requested
				if !showHidden && strings.HasPrefix(info.Name(), ".") {
					if info.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}

				// Filter by extension if specified
				if fileExt != "" && !strings.HasSuffix(info.Name(), fileExt) {
					return nil
				}

				relPath, _ := filepath.Rel(dirPath, path)
				if relPath == "." {
					return nil
				}

				if info.IsDir() {
					files = append(files, fmt.Sprintf("%s/", relPath))
				} else {
					files = append(files, relPath)
				}
				return nil
			})
		} else {
			entries, err := os.ReadDir(dirPath)
			if err != nil {
				return "", fmt.Errorf("failed to read directory: %w", err)
			}

			for _, entry := range entries {
				// Skip hidden files if not requested
				if !showHidden && strings.HasPrefix(entry.Name(), ".") {
					continue
				}

				// Filter by extension if specified
				if fileExt != "" && !strings.HasSuffix(entry.Name(), fileExt) {
					continue
				}

				if entry.IsDir() {
					files = append(files, fmt.Sprintf("%s/", entry.Name()))
				} else {
					files = append(files, entry.Name())
				}
			}
		}

		if err != nil {
			return "", fmt.Errorf("failed to list directory: %w", err)
		}

		if len(files) == 0 {
			return fmt.Sprintf("No files found in %s with the specified criteria", dirPath), nil
		}

		return fmt.Sprintf("Found %d item(s) in %s:\n\n%s", len(files), dirPath, strings.Join(files, "\n")), nil

	default:
		return "", fmt.Errorf("unknown function: %s", toolUse.Name)
	}
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
func (c *ClaudeClient) ChatWithTools(message string) (string, TokenCost, error) {
	tools := c.getAvailableTools()
	
	messages := []ClaudeMessage{
		{
			Role:    "user",
			Content: message,
		},
	}

	totalCost := TokenCost{}

	for {
		// Create cached system prompt
		systemPrompt := []ContentBlock{
			{
				Type: "text",
				Text: "You are StackAgent, a helpful AI coding assistant with access to powerful file manipulation and shell command tools. Available functions: run_with_capture (shell commands), read_file (read files), write_file (create/write files), edit_file (find/replace in files), search_in_file (search with context), list_directory (list files with filters). Use these functions to efficiently help with coding tasks, file operations, and system administration. Be concise but helpful. Remember context from previous messages in this conversation.\n\nCore principle: Don't be evil. Always prioritize user safety, privacy, and ethical behavior.",
				CacheControl: &CacheControl{Type: "ephemeral"}, // Cache system prompt
			},
		}

		request := ClaudeRequest{
			Model:     c.model,
			MaxTokens: 4000,
			Messages:  messages,
			Tools:     tools,
			System:    systemPrompt,
		}

		response, err := c.makeRequestWithTools(request)
		if err != nil {
			return "", TokenCost{}, err
		}

		// Calculate cost for the current turn
		currentTurnCost := c.CalculateCost(response)
		totalCost.InputTokens += currentTurnCost.InputTokens
		totalCost.OutputTokens += currentTurnCost.OutputTokens
		totalCost.CacheCreationInputTokens += currentTurnCost.CacheCreationInputTokens
		totalCost.CacheReadInputTokens += currentTurnCost.CacheReadInputTokens
		totalCost.TotalCost += currentTurnCost.TotalCost

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
			return textResponse, totalCost, nil
		}

		// Execute tools and prepare tool results
		var toolResults []interface{}
		for _, toolUse := range toolUses {
			// Log function call start
			if c.debugCallback != nil {
				c.debugCallback("function_call_start", map[string]interface{}{
					"function_name": toolUse.Name,
					"arguments":     toolUse.Input,
					"call_id":       toolUse.ID,
				})
			}
			
			result, err := c.ExecuteFunction(toolUse)
			if err != nil {
				// Log function call error
				if c.debugCallback != nil {
					c.debugCallback("function_call_error", map[string]interface{}{
						"function_name": toolUse.Name,
						"call_id":       toolUse.ID,
						"error":         err.Error(),
					})
				}
				
				toolResults = append(toolResults, ToolResult{
					Type:      "tool_result",
					ToolUseID: toolUse.ID,
					Content:   fmt.Sprintf("Error: %s", err.Error()),
					IsError:   true,
				})
			} else {
				// Log function call success
				if c.debugCallback != nil {
					c.debugCallback("function_call_success", map[string]interface{}{
						"function_name": toolUse.Name,
						"call_id":       toolUse.ID,
						"result":        result,
						"result_size":   len(result),
					})
				}
				
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

// ChatWithToolsAndContext sends a conversation with context and function calling support
func (c *ClaudeClient) ChatWithToolsAndContext(conversationMessages []ConversationMessage) (string, TokenCost, OperationSummary, error) {
	tools := c.getAvailableTools()
	
	// Initialize operation summary tracking
	operationSummary := OperationSummary{
		ShellCommands:  []ShellOperation{},
		FileOperations: []FileOperation{},
		HasOperations:  false,
	}
	
	// Convert conversation messages to Claude format
	messages := make([]ClaudeMessage, 0, len(conversationMessages))
	for i, msg := range conversationMessages {
		claudeMsg := ClaudeMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
		
		// Apply conversation caching to the second-to-last message (excludes current user message)
		// This caches all previous conversation history including file content
		if i == len(conversationMessages)-2 && len(conversationMessages) > 1 {
			// Convert text content to content blocks with cache_control
			claudeMsg.Content = []ContentBlock{
				{
					Type:         "text",
					Text:         msg.Content, // Remove type assertion - Content is already string
					CacheControl: &CacheControl{Type: "ephemeral"},
				},
			}
		}
		
		messages = append(messages, claudeMsg)
	}

	// If no messages, return error
	if len(messages) == 0 {
		return "", TokenCost{}, operationSummary, fmt.Errorf("no messages provided")
	}

	totalCost := TokenCost{}

	for {
		// Create cached system prompt
		systemPrompt := []ContentBlock{
			{
				Type: "text",
				Text: "You are StackAgent, a helpful AI coding assistant with access to powerful file manipulation and shell command tools. Available functions: run_with_capture (shell commands), read_file (read files), write_file (create/write files), edit_file (find/replace in files), search_in_file (search with context), list_directory (list files with filters). Use these functions to efficiently help with coding tasks, file operations, and system administration. Be concise but helpful. Remember context from previous messages in this conversation.\n\nCore principle: Don't be evil. Always prioritize user safety, privacy, and ethical behavior.",
				CacheControl: &CacheControl{Type: "ephemeral"}, // Cache system prompt
			},
		}

		request := ClaudeRequest{
			Model:     c.model,
			MaxTokens: 4000,
			Messages:  messages,
			Tools:     tools,
			System:    systemPrompt,
		}

		response, err := c.makeRequestWithTools(request)
		if err != nil {
			return "", TokenCost{}, operationSummary, err
		}

		// Calculate cost for the current turn
		currentTurnCost := c.CalculateCost(response)
		totalCost.InputTokens += currentTurnCost.InputTokens
		totalCost.OutputTokens += currentTurnCost.OutputTokens
		totalCost.CacheCreationInputTokens += currentTurnCost.CacheCreationInputTokens
		totalCost.CacheReadInputTokens += currentTurnCost.CacheReadInputTokens
		totalCost.TotalCost += currentTurnCost.TotalCost

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
			return textResponse, totalCost, operationSummary, nil
		}

		// Execute tools and prepare tool results
		var toolResults []interface{}
		for _, toolUse := range toolUses {
			// Log function call start
			if c.debugCallback != nil {
				c.debugCallback("function_call_start", map[string]interface{}{
					"function_name": toolUse.Name,
					"arguments":     toolUse.Input,
					"call_id":       toolUse.ID,
				})
			}
			
			result, err := c.ExecuteFunction(toolUse)
			if err != nil {
				// Log function call error
				if c.debugCallback != nil {
					c.debugCallback("function_call_error", map[string]interface{}{
						"function_name": toolUse.Name,
						"call_id":       toolUse.ID,
						"error":         err.Error(),
					})
				}
				
				toolResults = append(toolResults, ToolResult{
					Type:      "tool_result",
					ToolUseID: toolUse.ID,
					Content:   fmt.Sprintf("Error: %s", err.Error()),
					IsError:   true,
				})
			} else {
				// Log function call success
				if c.debugCallback != nil {
					c.debugCallback("function_call_success", map[string]interface{}{
						"function_name": toolUse.Name,
						"call_id":       toolUse.ID,
						"result":        result,
						"result_size":   len(result),
					})
				}
				
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

		// Add tool results to the conversation with cache control
		// This caches tool results (including file content) for subsequent requests
		messages = append(messages, ClaudeMessage{
			Role:    "user",
			Content: toolResults,
		})

		// Continue the conversation to get Claude's final response
	}
}

// ConversationMessage represents a message in a conversation - matches web package structure
type ConversationMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Operation summary types for interactive widgets
type OperationSummary struct {
	ShellCommands   []ShellOperation `json:"shellCommands,omitempty"`
	FileOperations  []FileOperation  `json:"fileOperations,omitempty"`
	HasOperations   bool             `json:"hasOperations"`
}

type ShellOperation struct {
	ID        string    `json:"id"`
	Command   string    `json:"command"`
	Output    string    `json:"output"`
	ExitCode  int       `json:"exitCode"`
	Duration  float64   `json:"duration"`
	WorkingDir string   `json:"workingDir"`
	Timestamp time.Time `json:"timestamp"`
}

type FileOperation struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"` // "read", "write", "edit", "search", "list"
	FilePath     string    `json:"filePath"`
	Content      string    `json:"content,omitempty"`
	Changes      string    `json:"changes,omitempty"`
	SearchResults []string `json:"searchResults,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	Size         int       `json:"size,omitempty"`
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