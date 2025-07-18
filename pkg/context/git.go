package context

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GitContextManager handles Git-specific context operations
type GitContextManager struct {
	*ContextManager
}

// NewGitContextManager creates a context manager with Git integration
func NewGitContextManager(basePath string) *GitContextManager {
	if basePath == "" {
		basePath = ".stackagent/context"
	}
	
	return &GitContextManager{
		ContextManager: NewContextManager(basePath),
	}
}

// OnGitCheckout handles context restoration when switching Git branches
func (gcm *GitContextManager) OnGitCheckout(from, to string) error {
	// Save current context with branch info
	if from != "" {
		gcm.context.Metadata.GitBranch = from
		gcm.context.Metadata.GitCommit = gcm.getCurrentCommit()
		if err := gcm.SaveContextFor(from); err != nil {
			return fmt.Errorf("failed to save context for branch %s: %w", from, err)
		}
		fmt.Printf("💾 Saved context for branch: %s\n", from)
	}
	
	// Try to restore context for target branch
	if gcm.ContextExistsAt(to) {
		if err := gcm.LoadContextFrom(to); err != nil {
			return fmt.Errorf("failed to load context for branch %s: %w", to, err)
		}
		fmt.Printf("🔄 Restored AI context from branch: %s\n", to)
	} else {
		// Create new context for this branch
		gcm.context.Metadata.GitBranch = to
		gcm.context.Metadata.GitCommit = gcm.getCurrentCommit()
		gcm.context.Memory = make(map[string]string) // Fresh memory for new branch
		
		// But keep some workspace state
		gcm.context.Workspace.ActiveFiles = []string{}
		gcm.context.Workspace.CommandHistory = []CommandRecord{}
		gcm.context.Workspace.ActiveHandles = []uint64{}
		
		fmt.Printf("🆕 Created new context for branch: %s\n", to)
	}
	
	return nil
}

// ContextExistsAt checks if context exists for a specific branch
func (gcm *GitContextManager) ContextExistsAt(branch string) bool {
	branchPath := gcm.getBranchContextPath(branch)
	contextFile := filepath.Join(branchPath, "metadata.json")
	
	_, err := os.Stat(contextFile)
	return err == nil
}

// LoadContextFrom loads context from a specific branch
func (gcm *GitContextManager) LoadContextFrom(branch string) error {
	branchPath := gcm.getBranchContextPath(branch)
	
	// Clear existing context before loading new branch context
	gcm.context.Memory = make(map[string]string)
	gcm.context.Knowledge = make(map[string]string)
	gcm.context.Workspace = WorkspaceState{
		ActiveFiles:    []string{},
		CommandHistory: []CommandRecord{},
		ActiveHandles:  []uint64{},
		LastActivity:   time.Now(),
	}
	
	// Temporarily change basePath to load from branch-specific directory
	originalBasePath := gcm.basePath
	gcm.basePath = branchPath
	
	err := gcm.LoadContext()
	
	// Restore original basePath
	gcm.basePath = originalBasePath
	
	if err != nil {
		return err
	}
	
	// Update metadata
	gcm.context.Metadata.GitBranch = branch
	gcm.context.Metadata.GitCommit = gcm.getCurrentCommit()
	
	return nil
}

// SaveContextFor saves context for a specific branch
func (gcm *GitContextManager) SaveContextFor(branch string) error {
	branchPath := gcm.getBranchContextPath(branch)
	
	// Ensure branch directory exists
	if err := os.MkdirAll(branchPath, 0755); err != nil {
		return fmt.Errorf("failed to create branch context directory: %w", err)
	}
	
	// Temporarily change basePath to save to branch-specific directory
	originalBasePath := gcm.basePath
	gcm.basePath = branchPath
	
	// Update Git metadata
	gcm.context.Metadata.GitBranch = branch
	gcm.context.Metadata.GitCommit = gcm.getCurrentCommit()
	
	err := gcm.SaveContext()
	
	// Restore original basePath
	gcm.basePath = originalBasePath
	
	return err
}

// getBranchContextPath returns the path for branch-specific context
func (gcm *GitContextManager) getBranchContextPath(branch string) string {
	// Clean branch name for filesystem
	cleanBranch := strings.ReplaceAll(branch, "/", "_")
	cleanBranch = strings.ReplaceAll(cleanBranch, "\\", "_")
	
	return filepath.Join(gcm.basePath, "branches", cleanBranch)
}

// getCurrentBranch returns the current Git branch
func (gcm *GitContextManager) getCurrentBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	
	return strings.TrimSpace(string(output))
}

// getCurrentCommit returns the current Git commit hash
func (gcm *GitContextManager) getCurrentCommit() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	
	return strings.TrimSpace(string(output))
}

// isGitRepository checks if current directory is a Git repository
func (gcm *GitContextManager) isGitRepository() bool {
	_, err := os.Stat(".git")
	return err == nil
}

// InitializeGitHooks sets up Git hooks for automatic context switching
func (gcm *GitContextManager) InitializeGitHooks() error {
	if !gcm.isGitRepository() {
		return fmt.Errorf("not a Git repository")
	}
	
	// Create post-checkout hook
	hookPath := ".git/hooks/post-checkout"
	hookContent := `#!/bin/bash
# StackAgent Context Management Hook
# This hook automatically switches AI context when changing branches

PREVIOUS_HEAD=$1
NEW_HEAD=$2
BRANCH_SWITCH=$3

if [ "$BRANCH_SWITCH" = "1" ]; then
    # Get branch names
    PREVIOUS_BRANCH=$(git name-rev --name-only $PREVIOUS_HEAD 2>/dev/null || echo "")
    NEW_BRANCH=$(git name-rev --name-only $NEW_HEAD 2>/dev/null || echo "")
    
    # Only try to restore context if stackagent-cli is available
    if command -v stackagent-cli &> /dev/null; then
        echo "🔄 StackAgent: Switching context from $PREVIOUS_BRANCH to $NEW_BRANCH"
        # Note: This would call stackagent-cli with context restore command
        # For now, just notify the user
        echo "💡 Run 'stackagent-cli context restore' to restore AI context"
    fi
fi
`
	
	if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
		return fmt.Errorf("failed to create post-checkout hook: %w", err)
	}
	
	fmt.Printf("✅ Git hooks initialized for context switching\n")
	return nil
}

// ListBranches returns all branches that have context
func (gcm *GitContextManager) ListBranches() []string {
	branchesPath := filepath.Join(gcm.basePath, "branches")
	entries, err := os.ReadDir(branchesPath)
	if err != nil {
		return []string{}
	}
	
	branches := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			// Convert filesystem-safe name back to branch name
			branch := strings.ReplaceAll(entry.Name(), "_", "/")
			branches = append(branches, branch)
		}
	}
	
	return branches
}

// GetBranchStats returns statistics for all branches with context
func (gcm *GitContextManager) GetBranchStats() map[string]ContextStats {
	branches := gcm.ListBranches()
	stats := make(map[string]ContextStats)
	
	for _, branch := range branches {
		branchPath := gcm.getBranchContextPath(branch)
		tempGCM := &GitContextManager{
			ContextManager: NewContextManager(branchPath),
		}
		
		if err := tempGCM.LoadContext(); err == nil {
			stats[branch] = tempGCM.GetStats()
		}
	}
	
	return stats
}

// CleanupBranches removes context for branches that no longer exist
func (gcm *GitContextManager) CleanupBranches() error {
	// Get all Git branches
	cmd := exec.Command("git", "branch", "-a")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get Git branches: %w", err)
	}
	
	// Parse branch names
	gitBranches := make(map[string]bool)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "*") {
			// Clean up branch name
			branch := strings.TrimPrefix(line, "remotes/origin/")
			branch = strings.TrimPrefix(branch, "* ")
			gitBranches[branch] = true
		}
	}
	
	// Check context branches
	contextBranches := gcm.ListBranches()
	cleaned := 0
	
	for _, branch := range contextBranches {
		if !gitBranches[branch] {
			// Branch no longer exists, remove its context
			branchPath := gcm.getBranchContextPath(branch)
			if err := os.RemoveAll(branchPath); err != nil {
				fmt.Printf("⚠️  Failed to remove context for branch %s: %v\n", branch, err)
			} else {
				fmt.Printf("🗑️  Removed context for deleted branch: %s\n", branch)
				cleaned++
			}
		}
	}
	
	if cleaned > 0 {
		fmt.Printf("✅ Cleaned up %d orphaned branch contexts\n", cleaned)
	} else {
		fmt.Printf("✅ No orphaned branch contexts found\n")
	}
	
	return nil
}

// SyncWithGit synchronizes context with current Git state
func (gcm *GitContextManager) SyncWithGit() error {
	if !gcm.isGitRepository() {
		return fmt.Errorf("not a Git repository")
	}
	
	currentBranch := gcm.getCurrentBranch()
	currentCommit := gcm.getCurrentCommit()
	
	// Update metadata
	gcm.context.Metadata.GitBranch = currentBranch
	gcm.context.Metadata.GitCommit = currentCommit
	
	// Save context for current branch
	if err := gcm.SaveContextFor(currentBranch); err != nil {
		return fmt.Errorf("failed to save context for current branch: %w", err)
	}
	
	fmt.Printf("🔄 Context synchronized with Git (branch: %s, commit: %s)\n", 
		currentBranch, currentCommit[:8])
	
	return nil
} 