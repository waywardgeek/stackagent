package context

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewContextManager(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewContextManager(tempDir)
	
	if cm == nil {
		t.Fatal("NewContextManager returned nil")
	}
	
	if cm.basePath != tempDir {
		t.Errorf("Expected basePath %s, got %s", tempDir, cm.basePath)
	}
	
	if cm.context == nil {
		t.Fatal("Context should not be nil")
	}
	
	if cm.context.Memory == nil {
		t.Error("Memory should be initialized")
	}
	
	if cm.context.Knowledge == nil {
		t.Error("Knowledge should be initialized")
	}
	
	if cm.context.Metadata.SessionID == "" {
		t.Error("SessionID should be generated")
	}
	
	if cm.context.Metadata.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", cm.context.Metadata.Version)
	}
}

func TestProtectedMemory(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewContextManager(tempDir)
	
	// Test SetProtected
	err := cm.SetProtected("test-key", "test-value")
	if err != nil {
		t.Fatalf("SetProtected failed: %v", err)
	}
	
	// Test GetProtected
	value, exists := cm.GetProtected("test-key")
	if !exists {
		t.Error("Expected key to exist")
	}
	
	if value != "test-value" {
		t.Errorf("Expected 'test-value', got '%s'", value)
	}
	
	// Test non-existent key
	_, exists = cm.GetProtected("non-existent")
	if exists {
		t.Error("Expected key to not exist")
	}
	
	// Test ListProtected
	keys := cm.ListProtected()
	if len(keys) != 1 {
		t.Errorf("Expected 1 key, got %d", len(keys))
	}
	
	if keys[0] != "test-key" {
		t.Errorf("Expected 'test-key', got '%s'", keys[0])
	}
	
	// Test DeleteProtected
	err = cm.DeleteProtected("test-key")
	if err != nil {
		t.Fatalf("DeleteProtected failed: %v", err)
	}
	
	_, exists = cm.GetProtected("test-key")
	if exists {
		t.Error("Expected key to be deleted")
	}
}

func TestKnowledgeStorage(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewContextManager(tempDir)
	
	// Test SetKnowledge
	err := cm.SetKnowledge("pattern1", "learned insight")
	if err != nil {
		t.Fatalf("SetKnowledge failed: %v", err)
	}
	
	// Test GetKnowledge
	value, exists := cm.GetKnowledge("pattern1")
	if !exists {
		t.Error("Expected knowledge to exist")
	}
	
	if value != "learned insight" {
		t.Errorf("Expected 'learned insight', got '%s'", value)
	}
	
	// Test non-existent knowledge
	_, exists = cm.GetKnowledge("non-existent")
	if exists {
		t.Error("Expected knowledge to not exist")
	}
}

func TestCommandHistory(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewContextManager(tempDir)
	
	// Test AddCommandRecord
	err := cm.AddCommandRecord("ls -la", 123, 0, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("AddCommandRecord failed: %v", err)
	}
	
	ws := cm.GetWorkspace()
	if len(ws.CommandHistory) != 1 {
		t.Errorf("Expected 1 command in history, got %d", len(ws.CommandHistory))
	}
	
	record := ws.CommandHistory[0]
	if record.Command != "ls -la" {
		t.Errorf("Expected 'ls -la', got '%s'", record.Command)
	}
	
	if record.HandleID != 123 {
		t.Errorf("Expected handle ID 123, got %d", record.HandleID)
	}
	
	if record.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", record.ExitCode)
	}
	
	if record.Duration != "100ms" {
		t.Errorf("Expected duration '100ms', got '%s'", record.Duration)
	}
}

func TestWorkspaceUpdate(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewContextManager(tempDir)
	
	// Create test workspace state
	ws := WorkspaceState{
		ActiveFiles:    []string{"file1.go", "file2.go"},
		WorkingDir:     "/tmp/test",
		ActiveHandles:  []uint64{1, 2, 3},
		CurrentTask:    "implement feature",
		ProjectContext: "test project",
	}
	
	// Test UpdateWorkspace
	err := cm.UpdateWorkspace(ws)
	if err != nil {
		t.Fatalf("UpdateWorkspace failed: %v", err)
	}
	
	// Test GetWorkspace
	retrievedWS := cm.GetWorkspace()
	
	if len(retrievedWS.ActiveFiles) != 2 {
		t.Errorf("Expected 2 active files, got %d", len(retrievedWS.ActiveFiles))
	}
	
	if retrievedWS.WorkingDir != "/tmp/test" {
		t.Errorf("Expected working dir '/tmp/test', got '%s'", retrievedWS.WorkingDir)
	}
	
	if len(retrievedWS.ActiveHandles) != 3 {
		t.Errorf("Expected 3 active handles, got %d", len(retrievedWS.ActiveHandles))
	}
	
	if retrievedWS.CurrentTask != "implement feature" {
		t.Errorf("Expected current task 'implement feature', got '%s'", retrievedWS.CurrentTask)
	}
	
	if retrievedWS.ProjectContext != "test project" {
		t.Errorf("Expected project context 'test project', got '%s'", retrievedWS.ProjectContext)
	}
}

func TestContextPersistence(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create first context manager
	cm1 := NewContextManager(tempDir)
	
	// Add some data
	err := cm1.SetProtected("persistent-key", "persistent-value")
	if err != nil {
		t.Fatalf("SetProtected failed: %v", err)
	}
	
	err = cm1.SetKnowledge("pattern1", "learned pattern")
	if err != nil {
		t.Fatalf("SetKnowledge failed: %v", err)
	}
	
	err = cm1.AddCommandRecord("test command", 456, 0, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("AddCommandRecord failed: %v", err)
	}
	
	// Save context
	err = cm1.SaveContext()
	if err != nil {
		t.Fatalf("SaveContext failed: %v", err)
	}
	
	// Create second context manager and load
	cm2 := NewContextManager(tempDir)
	err = cm2.LoadContext()
	if err != nil {
		t.Fatalf("LoadContext failed: %v", err)
	}
	
	// Verify data was loaded
	value, exists := cm2.GetProtected("persistent-key")
	if !exists {
		t.Error("Expected persistent key to exist after reload")
	}
	
	if value != "persistent-value" {
		t.Errorf("Expected 'persistent-value', got '%s'", value)
	}
	
	knowledge, exists := cm2.GetKnowledge("pattern1")
	if !exists {
		t.Error("Expected knowledge to exist after reload")
	}
	
	if knowledge != "learned pattern" {
		t.Errorf("Expected 'learned pattern', got '%s'", knowledge)
	}
	
	ws := cm2.GetWorkspace()
	if len(ws.CommandHistory) != 1 {
		t.Errorf("Expected 1 command in history after reload, got %d", len(ws.CommandHistory))
	}
	
	if ws.CommandHistory[0].Command != "test command" {
		t.Errorf("Expected 'test command', got '%s'", ws.CommandHistory[0].Command)
	}
}

func TestDirectoryStructure(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewContextManager(tempDir)
	
	// Add some data and save
	err := cm.SetProtected("test", "value")
	if err != nil {
		t.Fatalf("SetProtected failed: %v", err)
	}
	
	// Check that directories were created
	expectedDirs := []string{
		filepath.Join(tempDir, "memory"),
		filepath.Join(tempDir, "workspace"),
		filepath.Join(tempDir, "knowledge"),
	}
	
	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Expected directory %s to exist", dir)
		}
	}
	
	// Check that files were created
	expectedFiles := []string{
		filepath.Join(tempDir, "memory/current.json"),
		filepath.Join(tempDir, "workspace/state.json"),
		filepath.Join(tempDir, "knowledge/patterns.json"),
		filepath.Join(tempDir, "metadata.json"),
	}
	
	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", file)
		}
	}
}

func TestContextStats(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewContextManager(tempDir)
	
	// Add some data
	cm.SetProtected("key1", "value1")
	cm.SetProtected("key2", "value2")
	cm.SetKnowledge("pattern1", "insight1")
	cm.AddCommandRecord("cmd1", 1, 0, 100*time.Millisecond)
	cm.AddCommandRecord("cmd2", 2, 0, 200*time.Millisecond)
	
	// Update workspace
	ws := cm.GetWorkspace()
	ws.ActiveFiles = []string{"file1.go", "file2.go", "file3.go"}
	ws.ActiveHandles = []uint64{1, 2, 3, 4}
	cm.UpdateWorkspace(ws)
	
	// Get stats
	stats := cm.GetStats()
	
	if stats.MemoryEntries != 2 {
		t.Errorf("Expected 2 memory entries, got %d", stats.MemoryEntries)
	}
	
	if stats.KnowledgeEntries != 1 {
		t.Errorf("Expected 1 knowledge entry, got %d", stats.KnowledgeEntries)
	}
	
	if stats.CommandHistory != 2 {
		t.Errorf("Expected 2 command history entries, got %d", stats.CommandHistory)
	}
	
	if stats.ActiveFiles != 3 {
		t.Errorf("Expected 3 active files, got %d", stats.ActiveFiles)
	}
	
	if stats.ActiveHandles != 4 {
		t.Errorf("Expected 4 active handles, got %d", stats.ActiveHandles)
	}
	
	if stats.SessionID == "" {
		t.Error("SessionID should not be empty")
	}
	
	if stats.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	
	if stats.LastUpdated.IsZero() {
		t.Error("LastUpdated should not be zero")
	}
}

func TestCommandHistoryLimit(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewContextManager(tempDir)
	
	// Add 105 commands (more than the 100 limit)
	for i := 0; i < 105; i++ {
		err := cm.AddCommandRecord(fmt.Sprintf("command-%d", i), uint64(i), 0, 10*time.Millisecond)
		if err != nil {
			t.Fatalf("AddCommandRecord failed: %v", err)
		}
	}
	
	ws := cm.GetWorkspace()
	if len(ws.CommandHistory) != 100 {
		t.Errorf("Expected command history to be limited to 100, got %d", len(ws.CommandHistory))
	}
	
	// Check that the oldest commands were removed
	if ws.CommandHistory[0].Command != "command-5" {
		t.Errorf("Expected first command to be 'command-5', got '%s'", ws.CommandHistory[0].Command)
	}
	
	if ws.CommandHistory[99].Command != "command-104" {
		t.Errorf("Expected last command to be 'command-104', got '%s'", ws.CommandHistory[99].Command)
	}
}

func TestConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewContextManager(tempDir)
	
	// Test concurrent access to protected memory
	done := make(chan bool, 10)
	
	// Multiple goroutines setting different keys
	for i := 0; i < 10; i++ {
		go func(index int) {
			defer func() { done <- true }()
			
			key := fmt.Sprintf("key-%d", index)
			value := fmt.Sprintf("value-%d", index)
			
			err := cm.SetProtected(key, value)
			if err != nil {
				t.Errorf("SetProtected failed: %v", err)
			}
			
			// Verify the value was set
			retrievedValue, exists := cm.GetProtected(key)
			if !exists {
				t.Errorf("Expected key %s to exist", key)
			}
			
			if retrievedValue != value {
				t.Errorf("Expected value %s, got %s", value, retrievedValue)
			}
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify all keys were set
	keys := cm.ListProtected()
	if len(keys) != 10 {
		t.Errorf("Expected 10 keys, got %d", len(keys))
	}
} 