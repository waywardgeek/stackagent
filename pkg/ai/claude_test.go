package ai

import (
	"os"
	"testing"
	"time"

	"stackagent/pkg/shell"
)

func TestNewClaudeClient(t *testing.T) {
	// Test without API key
	os.Unsetenv("ANTHROPIC_API_KEY")
	_, err := NewClaudeClient()
	if err == nil {
		t.Error("Expected error when ANTHROPIC_API_KEY is not set")
	}

	// Test with API key
	os.Setenv("ANTHROPIC_API_KEY", "test-key")
	defer os.Unsetenv("ANTHROPIC_API_KEY")
	
	client, err := NewClaudeClient()
	if err != nil {
		t.Errorf("Expected no error with API key set, got: %v", err)
	}
	
	if client.apiKey != "test-key" {
		t.Errorf("Expected API key 'test-key', got: %s", client.apiKey)
	}
	
	if client.model != "claude-sonnet-4-20250514" {
		t.Errorf("Expected default model 'claude-sonnet-4-20250514', got: %s", client.model)
	}
}

func TestSetModel(t *testing.T) {
	os.Setenv("ANTHROPIC_API_KEY", "test-key")
	defer os.Unsetenv("ANTHROPIC_API_KEY")
	
	client, err := NewClaudeClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	newModel := "claude-3-5-haiku-20241022"
	client.SetModel(newModel)
	
	if client.model != newModel {
		t.Errorf("Expected model %s, got: %s", newModel, client.model)
	}
}

func TestGetAvailableModels(t *testing.T) {
	os.Setenv("ANTHROPIC_API_KEY", "test-key")
	defer os.Unsetenv("ANTHROPIC_API_KEY")
	
	client, err := NewClaudeClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	models := client.GetAvailableModels()
	expectedModels := []string{
		"claude-sonnet-4-20250514",
		"claude-3-5-sonnet-20241022",
		"claude-3-5-haiku-20241022",
		"claude-3-opus-20240229",
	}
	
	if len(models) != len(expectedModels) {
		t.Errorf("Expected %d models, got %d", len(expectedModels), len(models))
	}
	
	for i, expected := range expectedModels {
		if models[i] != expected {
			t.Errorf("Expected model %s at index %d, got %s", expected, i, models[i])
		}
	}
}

func TestEstimateCost(t *testing.T) {
	os.Setenv("ANTHROPIC_API_KEY", "test-key")
	defer os.Unsetenv("ANTHROPIC_API_KEY")
	
	client, err := NewClaudeClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	// Test cost calculation
	// 1000 input tokens = $0.003, 1000 output tokens = $0.015
	cost := client.EstimateCost(1000, 1000)
	expected := 0.003 + 0.015 // $0.018
	
	if cost != expected {
		t.Errorf("Expected cost $%.6f, got $%.6f", expected, cost)
	}
}

func TestEstimateTokens(t *testing.T) {
	os.Setenv("ANTHROPIC_API_KEY", "test-key")
	defer os.Unsetenv("ANTHROPIC_API_KEY")
	
	client, err := NewClaudeClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	// Test token estimation (4 chars per token)
	text := "This is a test message"
	tokens := client.EstimateTokens(text)
	expected := len(text) / 4
	
	if tokens != expected {
		t.Errorf("Expected %d tokens, got %d", expected, tokens)
	}
}

func TestExtractSection(t *testing.T) {
	text := `Summary: This is a test summary
	
	Other content here`
	
	result := extractSection(text, "summary")
	expected := "This is a test summary"
	
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestExtractList(t *testing.T) {
	text := `Key findings:
	- First finding
	- Second finding
	• Third finding
	* Fourth finding
	
	Other section`
	
	result := extractList(text, "key findings")
	expected := []string{
		"First finding",
		"Second finding",
		"Third finding",
		"Fourth finding",
	}
	
	if len(result) != len(expected) {
		t.Errorf("Expected %d items, got %d", len(expected), len(result))
	}
	
	for i, item := range expected {
		if i < len(result) && result[i] != item {
			t.Errorf("Expected item '%s' at index %d, got '%s'", item, i, result[i])
		}
	}
}

func TestExtractRisk(t *testing.T) {
	tests := []struct {
		text     string
		expected string
	}{
		{"This is high risk and dangerous", "high"},
		{"Medium risk situation with caution needed", "medium"},
		{"This is a low risk operation", "low"},
		{"No risk indicators", "low"},
	}
	
	for _, test := range tests {
		result := extractRisk(test.text)
		if result != test.expected {
			t.Errorf("For text '%s', expected risk '%s', got '%s'", 
				test.text, test.expected, result)
		}
	}
}

func TestAnalyzeCommandWithoutAPIKey(t *testing.T) {
	// Test that methods fail gracefully without API key
	os.Unsetenv("ANTHROPIC_API_KEY")
	
	_, err := NewClaudeClient()
	if err == nil {
		t.Error("Expected error when creating client without API key")
	}
}

func TestIntegrationWithShellManager(t *testing.T) {
	// This is a unit test that doesn't make real API calls
	// It tests the integration structure
	
	sm := shell.NewShellManager()
	
	// Create a test handle
	handle, err := sm.RunWithCapture("echo test")
	if err != nil {
		t.Fatalf("Failed to create handle: %v", err)
	}
	
	// Wait for completion
	time.Sleep(100 * time.Millisecond)
	
	// Test that we can retrieve the handle
	retrievedHandle, exists := sm.GetHandle(handle.ID)
	if !exists {
		t.Error("Handle should exist")
	}
	
	if retrievedHandle.ID != handle.ID {
		t.Errorf("Expected handle ID %d, got %d", handle.ID, retrievedHandle.ID)
	}
	
	// Test that we can get output for analysis
	output, err := sm.GetTail(handle.ID, 10)
	if err != nil {
		t.Errorf("Failed to get output: %v", err)
	}
	
	if output == "" {
		t.Error("Expected non-empty output")
	}
}

// Mock test for API structure validation
func TestClaudeRequestStructure(t *testing.T) {
	request := ClaudeRequest{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 1024,
		System:    "You are a helpful assistant",
		Messages: []ClaudeMessage{
			{Role: "user", Content: "Hello"},
		},
	}
	
	if request.Model != "claude-sonnet-4-20250514" {
		t.Error("Model not set correctly")
	}
	
	if request.MaxTokens != 1024 {
		t.Error("MaxTokens not set correctly")
	}
	
	if len(request.Messages) != 1 {
		t.Error("Messages not set correctly")
	}
	
	if request.Messages[0].Role != "user" {
		t.Error("Message role not set correctly")
	}
}

func TestDebugLogging(t *testing.T) {
	os.Setenv("ANTHROPIC_API_KEY", "test-key")
	defer os.Unsetenv("ANTHROPIC_API_KEY")
	
	client, err := NewClaudeClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	// Test enabling debug logging
	debugFile := "test_debug.log"
	defer os.Remove(debugFile) // Clean up after test
	
	err = client.EnableDebugLogging(debugFile)
	if err != nil {
		t.Fatalf("Failed to enable debug logging: %v", err)
	}
	
	// Verify debug is enabled
	if !client.debugEnabled {
		t.Error("Debug logging should be enabled")
	}
	
	if client.debugFile == nil {
		t.Error("Debug file should be set")
	}
	
	// Test disabling debug logging
	err = client.DisableDebugLogging()
	if err != nil {
		t.Fatalf("Failed to disable debug logging: %v", err)
	}
	
	// Verify debug is disabled
	if client.debugEnabled {
		t.Error("Debug logging should be disabled")
	}
	
	if client.debugFile != nil {
		t.Error("Debug file should be nil after disabling")
	}
	
	// Verify debug file was created and has content
	if _, err := os.Stat(debugFile); os.IsNotExist(err) {
		t.Error("Debug file should exist after logging")
	}
}

func TestDebugLogEntry(t *testing.T) {
	entry := DebugLogEntry{
		Timestamp:  "2024-01-01T00:00:00Z",
		Type:       "request",
		Method:     "POST",
		URL:        "https://api.anthropic.com/v1/messages",
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]interface{}{
			"model": "claude-sonnet-4-20250514",
		},
	}
	
	if entry.Timestamp == "" {
		t.Error("Timestamp should be set")
	}
	
	if entry.Type != "request" {
		t.Error("Type should be 'request'")
	}
	
	if entry.Method != "POST" {
		t.Error("Method should be 'POST'")
	}
	
	if entry.StatusCode != 200 {
		t.Error("StatusCode should be 200")
	}
	
	if entry.Headers["Content-Type"] != "application/json" {
		t.Error("Content-Type header should be application/json")
	}
	
	if entry.Body.(map[string]interface{})["model"] != "claude-sonnet-4-20250514" {
		t.Error("Model in body should be claude-sonnet-4-20250514")
	}
}

func TestCommandAnalysisStructure(t *testing.T) {
	analysis := CommandAnalysis{
		Summary:     "Test summary",
		KeyFindings: []string{"Finding 1", "Finding 2"},
		Suggestions: []string{"Suggestion 1"},
		Risk:        "low",
		TokensUsed:  100,
		Cost:        0.001,
	}
	
	if analysis.Summary != "Test summary" {
		t.Error("Summary not set correctly")
	}
	
	if len(analysis.KeyFindings) != 2 {
		t.Error("KeyFindings not set correctly")
	}
	
	if len(analysis.Suggestions) != 1 {
		t.Error("Suggestions not set correctly")
	}
	
	if analysis.Risk != "low" {
		t.Error("Risk not set correctly")
	}
	
	if analysis.TokensUsed != 100 {
		t.Error("TokensUsed not set correctly")
	}
	
	if analysis.Cost != 0.001 {
		t.Error("Cost not set correctly")
	}
} 