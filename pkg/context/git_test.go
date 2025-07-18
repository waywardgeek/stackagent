package context

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewGitContextManager(t *testing.T) {
	tempDir := t.TempDir()
	gcm := NewGitContextManager(tempDir)
	
	if gcm == nil {
		t.Fatal("NewGitContextManager returned nil")
	}
	
	if gcm.ContextManager == nil {
		t.Fatal("ContextManager should not be nil")
	}
	
	if gcm.basePath != tempDir {
		t.Errorf("Expected basePath %s, got %s", tempDir, gcm.basePath)
	}
}

func TestGetBranchContextPath(t *testing.T) {
	tempDir := t.TempDir()
	gcm := NewGitContextManager(tempDir)
	
	// Test simple branch name
	path := gcm.getBranchContextPath("main")
	expected := filepath.Join(tempDir, "branches", "main")
	if path != expected {
		t.Errorf("Expected path %s, got %s", expected, path)
	}
	
	// Test branch with slash
	path = gcm.getBranchContextPath("feature/user-auth")
	expected = filepath.Join(tempDir, "branches", "feature_user-auth")
	if path != expected {
		t.Errorf("Expected path %s, got %s", expected, path)
	}
	
	// Test branch with backslash
	path = gcm.getBranchContextPath("hotfix\\bug-fix")
	expected = filepath.Join(tempDir, "branches", "hotfix_bug-fix")
	if path != expected {
		t.Errorf("Expected path %s, got %s", expected, path)
	}
}

func TestSaveContextFor(t *testing.T) {
	tempDir := t.TempDir()
	gcm := NewGitContextManager(tempDir)
	
	// Add some data
	err := gcm.SetProtected("branch-specific-key", "branch-specific-value")
	if err != nil {
		t.Fatalf("SetProtected failed: %v", err)
	}
	
	// Save context for specific branch
	err = gcm.SaveContextFor("feature/test")
	if err != nil {
		t.Fatalf("SaveContextFor failed: %v", err)
	}
	
	// Check that branch directory was created
	branchPath := filepath.Join(tempDir, "branches", "feature_test")
	if _, err := os.Stat(branchPath); os.IsNotExist(err) {
		t.Errorf("Expected branch directory %s to exist", branchPath)
	}
	
	// Check that files were created
	expectedFiles := []string{
		filepath.Join(branchPath, "memory/current.json"),
		filepath.Join(branchPath, "workspace/state.json"),
		filepath.Join(branchPath, "knowledge/patterns.json"),
		filepath.Join(branchPath, "metadata.json"),
	}
	
	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", file)
		}
	}
}

func TestLoadContextFrom(t *testing.T) {
	tempDir := t.TempDir()
	gcm := NewGitContextManager(tempDir)
	
	// Add some data and save for a specific branch
	err := gcm.SetProtected("branch-key", "branch-value")
	if err != nil {
		t.Fatalf("SetProtected failed: %v", err)
	}
	
	err = gcm.SetKnowledge("branch-pattern", "branch-insight")
	if err != nil {
		t.Fatalf("SetKnowledge failed: %v", err)
	}
	
	err = gcm.SaveContextFor("feature/test")
	if err != nil {
		t.Fatalf("SaveContextFor failed: %v", err)
	}
	
	// Clear current context
	gcm.context.Memory = make(map[string]string)
	gcm.context.Knowledge = make(map[string]string)
	
	// Load context from the branch
	err = gcm.LoadContextFrom("feature/test")
	if err != nil {
		t.Fatalf("LoadContextFrom failed: %v", err)
	}
	
	// Verify data was loaded
	value, exists := gcm.GetProtected("branch-key")
	if !exists {
		t.Error("Expected branch-key to exist after load")
	}
	
	if value != "branch-value" {
		t.Errorf("Expected 'branch-value', got '%s'", value)
	}
	
	knowledge, exists := gcm.GetKnowledge("branch-pattern")
	if !exists {
		t.Error("Expected branch-pattern to exist after load")
	}
	
	if knowledge != "branch-insight" {
		t.Errorf("Expected 'branch-insight', got '%s'", knowledge)
	}
}

func TestContextExistsAt(t *testing.T) {
	tempDir := t.TempDir()
	gcm := NewGitContextManager(tempDir)
	
	// Should not exist initially
	if gcm.ContextExistsAt("nonexistent") {
		t.Error("Expected context to not exist for nonexistent branch")
	}
	
	// Save context for a branch
	err := gcm.SetProtected("test-key", "test-value")
	if err != nil {
		t.Fatalf("SetProtected failed: %v", err)
	}
	
	err = gcm.SaveContextFor("test-branch")
	if err != nil {
		t.Fatalf("SaveContextFor failed: %v", err)
	}
	
	// Should exist now
	if !gcm.ContextExistsAt("test-branch") {
		t.Error("Expected context to exist for test-branch")
	}
}

func TestOnGitCheckout(t *testing.T) {
	tempDir := t.TempDir()
	gcm := NewGitContextManager(tempDir)
	
	// Use a mock that doesn't rely on git commands
	// We'll test the basic context switching logic without git metadata
	
	// Setup initial context for branch1
	err := gcm.SetProtected("branch1-key", "branch1-value")
	if err != nil {
		t.Fatalf("SetProtected failed: %v", err)
	}
	
	// Test switching from branch1 to branch2 (new branch)
	err = gcm.OnGitCheckout("branch1", "branch2")
	if err != nil {
		t.Fatalf("OnGitCheckout failed: %v", err)
	}
	
	// Verify branch1 context was saved
	if !gcm.ContextExistsAt("branch1") {
		t.Error("Expected branch1 context to be saved")
	}
	
	// Verify branch2 has fresh context
	_, exists := gcm.GetProtected("branch1-key")
	if exists {
		t.Error("Expected branch1-key to not exist in branch2 context")
	}
	
	// Add some data to branch2
	err = gcm.SetProtected("branch2-key", "branch2-value")
	if err != nil {
		t.Fatalf("SetProtected failed: %v", err)
	}
	
	// Switch back to branch1
	err = gcm.OnGitCheckout("branch2", "branch1")
	if err != nil {
		t.Fatalf("OnGitCheckout failed: %v", err)
	}
	
	// Verify branch1 context was restored
	value, exists := gcm.GetProtected("branch1-key")
	if !exists {
		t.Error("Expected branch1-key to exist after switching back")
	}
	
	if value != "branch1-value" {
		t.Errorf("Expected 'branch1-value', got '%s'", value)
	}
	
	// Verify branch2-key doesn't exist in branch1 context
	_, exists = gcm.GetProtected("branch2-key")
	if exists {
		t.Error("Expected branch2-key to not exist in branch1 context")
	}
}

func TestListBranches(t *testing.T) {
	tempDir := t.TempDir()
	gcm := NewGitContextManager(tempDir)
	
	// Initially no branches
	branches := gcm.ListBranches()
	if len(branches) != 0 {
		t.Errorf("Expected 0 branches, got %d", len(branches))
	}
	
	// Save context for multiple branches
	branchNames := []string{"main", "feature/auth", "hotfix/bug-123"}
	
	for _, branch := range branchNames {
		err := gcm.SetProtected(fmt.Sprintf("key-%s", branch), "value")
		if err != nil {
			t.Fatalf("SetProtected failed: %v", err)
		}
		
		err = gcm.SaveContextFor(branch)
		if err != nil {
			t.Fatalf("SaveContextFor failed: %v", err)
		}
	}
	
	// List branches
	branches = gcm.ListBranches()
	if len(branches) != 3 {
		t.Errorf("Expected 3 branches, got %d", len(branches))
	}
	
	// Check that all branches are present
	branchMap := make(map[string]bool)
	for _, branch := range branches {
		branchMap[branch] = true
	}
	
	for _, expected := range branchNames {
		if !branchMap[expected] {
			t.Errorf("Expected branch %s to be in list", expected)
		}
	}
}

func TestGetBranchStats(t *testing.T) {
	tempDir := t.TempDir()
	gcm := NewGitContextManager(tempDir)
	
	// Create context for branch1
	err := gcm.SetProtected("memory-key", "memory-value")
	if err != nil {
		t.Fatalf("SetProtected failed: %v", err)
	}
	
	err = gcm.SetKnowledge("pattern1", "insight1")
	if err != nil {
		t.Fatalf("SetKnowledge failed: %v", err)
	}
	
	err = gcm.AddCommandRecord("test-cmd", 123, 0, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("AddCommandRecord failed: %v", err)
	}
	
	err = gcm.SaveContextFor("branch1")
	if err != nil {
		t.Fatalf("SaveContextFor failed: %v", err)
	}
	
	// Create different context for branch2
	gcm.context.Memory = make(map[string]string)
	gcm.context.Knowledge = make(map[string]string)
	gcm.context.Workspace.CommandHistory = []CommandRecord{}
	
	err = gcm.SetProtected("key1", "value1")
	if err != nil {
		t.Fatalf("SetProtected failed: %v", err)
	}
	
	err = gcm.SetProtected("key2", "value2")
	if err != nil {
		t.Fatalf("SetProtected failed: %v", err)
	}
	
	err = gcm.SaveContextFor("branch2")
	if err != nil {
		t.Fatalf("SaveContextFor failed: %v", err)
	}
	
	// Get stats for all branches
	stats := gcm.GetBranchStats()
	
	if len(stats) != 2 {
		t.Errorf("Expected stats for 2 branches, got %d", len(stats))
	}
	
	// Check branch1 stats
	branch1Stats, exists := stats["branch1"]
	if !exists {
		t.Error("Expected stats for branch1")
	} else {
		if branch1Stats.MemoryEntries != 1 {
			t.Errorf("Expected 1 memory entry for branch1, got %d", branch1Stats.MemoryEntries)
		}
		
		if branch1Stats.KnowledgeEntries != 1 {
			t.Errorf("Expected 1 knowledge entry for branch1, got %d", branch1Stats.KnowledgeEntries)
		}
		
		if branch1Stats.CommandHistory != 1 {
			t.Errorf("Expected 1 command history entry for branch1, got %d", branch1Stats.CommandHistory)
		}
	}
	
	// Check branch2 stats
	branch2Stats, exists := stats["branch2"]
	if !exists {
		t.Error("Expected stats for branch2")
	} else {
		if branch2Stats.MemoryEntries != 2 {
			t.Errorf("Expected 2 memory entries for branch2, got %d", branch2Stats.MemoryEntries)
		}
		
		if branch2Stats.KnowledgeEntries != 0 {
			t.Errorf("Expected 0 knowledge entries for branch2, got %d", branch2Stats.KnowledgeEntries)
		}
		
		if branch2Stats.CommandHistory != 0 {
			t.Errorf("Expected 0 command history entries for branch2, got %d", branch2Stats.CommandHistory)
		}
	}
}

func TestContextSeparation(t *testing.T) {
	tempDir := t.TempDir()
	gcm := NewGitContextManager(tempDir)
	
	// Test that contexts are completely separate between branches
	
	// Setup branch1 context
	err := gcm.SetProtected("shared-key", "branch1-value")
	if err != nil {
		t.Fatalf("SetProtected failed: %v", err)
	}
	
	err = gcm.SetKnowledge("pattern", "branch1-pattern")
	if err != nil {
		t.Fatalf("SetKnowledge failed: %v", err)
	}
	
	err = gcm.SaveContextFor("branch1")
	if err != nil {
		t.Fatalf("SaveContextFor failed: %v", err)
	}
	
	// Setup branch2 context with same keys but different values
	err = gcm.SetProtected("shared-key", "branch2-value")
	if err != nil {
		t.Fatalf("SetProtected failed: %v", err)
	}
	
	err = gcm.SetKnowledge("pattern", "branch2-pattern")
	if err != nil {
		t.Fatalf("SetKnowledge failed: %v", err)
	}
	
	err = gcm.SaveContextFor("branch2")
	if err != nil {
		t.Fatalf("SaveContextFor failed: %v", err)
	}
	
	// Load branch1 context
	err = gcm.LoadContextFrom("branch1")
	if err != nil {
		t.Fatalf("LoadContextFrom failed: %v", err)
	}
	
	// Verify branch1 values
	value, exists := gcm.GetProtected("shared-key")
	if !exists {
		t.Error("Expected shared-key to exist in branch1")
	}
	
	if value != "branch1-value" {
		t.Errorf("Expected 'branch1-value', got '%s'", value)
	}
	
	knowledge, exists := gcm.GetKnowledge("pattern")
	if !exists {
		t.Error("Expected pattern to exist in branch1")
	}
	
	if knowledge != "branch1-pattern" {
		t.Errorf("Expected 'branch1-pattern', got '%s'", knowledge)
	}
	
	// Load branch2 context
	err = gcm.LoadContextFrom("branch2")
	if err != nil {
		t.Fatalf("LoadContextFrom failed: %v", err)
	}
	
	// Verify branch2 values
	value, exists = gcm.GetProtected("shared-key")
	if !exists {
		t.Error("Expected shared-key to exist in branch2")
	}
	
	if value != "branch2-value" {
		t.Errorf("Expected 'branch2-value', got '%s'", value)
	}
	
	knowledge, exists = gcm.GetKnowledge("pattern")
	if !exists {
		t.Error("Expected pattern to exist in branch2")
	}
	
	if knowledge != "branch2-pattern" {
		t.Errorf("Expected 'branch2-pattern', got '%s'", knowledge)
	}
}

func TestEmptyBranchSwitch(t *testing.T) {
	tempDir := t.TempDir()
	gcm := NewGitContextManager(tempDir)
	
	// Test switching from empty branch (no previous context)
	err := gcm.OnGitCheckout("", "new-branch")
	if err != nil {
		t.Fatalf("OnGitCheckout with empty 'from' failed: %v", err)
	}
	
	// Should create new context for new-branch
	if gcm.context.Metadata.GitBranch != "new-branch" {
		t.Errorf("Expected GitBranch to be 'new-branch', got '%s'", gcm.context.Metadata.GitBranch)
	}
	
	// Memory should be empty for new branch
	if len(gcm.context.Memory) != 0 {
		t.Errorf("Expected empty memory for new branch, got %d entries", len(gcm.context.Memory))
	}
} 