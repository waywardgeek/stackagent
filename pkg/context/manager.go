package context

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ContextManager manages AI context persistence across sessions
type ContextManager struct {
	basePath string
	context  *Context
	mutex    sync.RWMutex
}

// Context represents the AI's persistent state
type Context struct {
	Memory    map[string]string `json:"memory"`    // Protected memory that survives sessions
	Workspace WorkspaceState    `json:"workspace"` // Active files, commands, current state
	Knowledge map[string]string `json:"knowledge"` // Learned patterns and insights
	Metadata  ContextMetadata   `json:"metadata"`  // Session info, timestamps
}

// WorkspaceState tracks the current working state
type WorkspaceState struct {
	ActiveFiles     []string          `json:"active_files"`     // Files being worked on
	CommandHistory  []CommandRecord   `json:"command_history"`  // Recent commands
	WorkingDir      string            `json:"working_dir"`      // Current directory
	ActiveHandles   []uint64          `json:"active_handles"`   // Output handles in use
	CurrentTask     string            `json:"current_task"`     // What we're working on
	ProjectContext  string            `json:"project_context"`  // Project description
	LastActivity    time.Time         `json:"last_activity"`    // When last used
}

// CommandRecord stores information about executed commands
type CommandRecord struct {
	Command   string    `json:"command"`
	HandleID  uint64    `json:"handle_id"`
	Timestamp time.Time `json:"timestamp"`
	ExitCode  int       `json:"exit_code"`
	Duration  string    `json:"duration"`
	Summary   string    `json:"summary,omitempty"` // AI-generated summary
}

// ContextMetadata stores session and versioning information
type ContextMetadata struct {
	Created     time.Time `json:"created"`
	LastUpdated time.Time `json:"last_updated"`
	Version     string    `json:"version"`
	SessionID   string    `json:"session_id"`
	GitBranch   string    `json:"git_branch,omitempty"`
	GitCommit   string    `json:"git_commit,omitempty"`
}

// NewContextManager creates a new context manager
func NewContextManager(basePath string) *ContextManager {
	if basePath == "" {
		basePath = ".stackagent/context"
	}
	
	cm := &ContextManager{
		basePath: basePath,
		context:  &Context{
			Memory:    make(map[string]string),
			Workspace: WorkspaceState{
				ActiveFiles:    []string{},
				CommandHistory: []CommandRecord{},
				ActiveHandles:  []uint64{},
				LastActivity:   time.Now(),
			},
			Knowledge: make(map[string]string),
			Metadata: ContextMetadata{
				Created:   time.Now(),
				Version:   "1.0",
				SessionID: generateSessionID(),
			},
		},
	}
	
	// Ensure directory exists
	os.MkdirAll(basePath, 0755)
	
	return cm
}

// SetProtected sets a protected memory value that survives sessions
func (cm *ContextManager) SetProtected(key, value string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	cm.context.Memory[key] = value
	cm.context.Metadata.LastUpdated = time.Now()
	
	// Immediately save to disk
	return cm.saveContextUnsafe()
}

// GetProtected retrieves a protected memory value
func (cm *ContextManager) GetProtected(key string) (string, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	value, exists := cm.context.Memory[key]
	return value, exists
}

// ListProtected returns all protected memory keys
func (cm *ContextManager) ListProtected() []string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	keys := make([]string, 0, len(cm.context.Memory))
	for key := range cm.context.Memory {
		keys = append(keys, key)
	}
	return keys
}

// DeleteProtected removes a protected memory value
func (cm *ContextManager) DeleteProtected(key string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	delete(cm.context.Memory, key)
	cm.context.Metadata.LastUpdated = time.Now()
	
	return cm.saveContextUnsafe()
}

// UpdateWorkspace updates the workspace state
func (cm *ContextManager) UpdateWorkspace(ws WorkspaceState) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	cm.context.Workspace = ws
	cm.context.Workspace.LastActivity = time.Now()
	cm.context.Metadata.LastUpdated = time.Now()
	
	return cm.saveContextUnsafe()
}

// AddCommandRecord adds a command to the history
func (cm *ContextManager) AddCommandRecord(cmd string, handleID uint64, exitCode int, duration time.Duration) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	record := CommandRecord{
		Command:   cmd,
		HandleID:  handleID,
		Timestamp: time.Now(),
		ExitCode:  exitCode,
		Duration:  duration.String(),
	}
	
	// Keep only last 100 commands
	cm.context.Workspace.CommandHistory = append(cm.context.Workspace.CommandHistory, record)
	if len(cm.context.Workspace.CommandHistory) > 100 {
		cm.context.Workspace.CommandHistory = cm.context.Workspace.CommandHistory[1:]
	}
	
	cm.context.Workspace.LastActivity = time.Now()
	cm.context.Metadata.LastUpdated = time.Now()
	
	return cm.saveContextUnsafe()
}

// SetKnowledge stores learned patterns or insights
func (cm *ContextManager) SetKnowledge(key, value string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	cm.context.Knowledge[key] = value
	cm.context.Metadata.LastUpdated = time.Now()
	
	return cm.saveContextUnsafe()
}

// GetKnowledge retrieves stored knowledge
func (cm *ContextManager) GetKnowledge(key string) (string, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	value, exists := cm.context.Knowledge[key]
	return value, exists
}

// GetWorkspace returns the current workspace state
func (cm *ContextManager) GetWorkspace() WorkspaceState {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	return cm.context.Workspace
}

// GetContext returns a copy of the current context
func (cm *ContextManager) GetContext() Context {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	return *cm.context
}

// SaveContext saves the current context to disk
func (cm *ContextManager) SaveContext() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	return cm.saveContextUnsafe()
}

// saveContextUnsafe saves context without locking (internal use)
func (cm *ContextManager) saveContextUnsafe() error {
	// Create directory structure
	dirs := []string{
		filepath.Join(cm.basePath, "memory"),
		filepath.Join(cm.basePath, "workspace"),
		filepath.Join(cm.basePath, "knowledge"),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	
	// Save memory
	if err := cm.saveJSONFile(filepath.Join(cm.basePath, "memory/current.json"), cm.context.Memory); err != nil {
		return fmt.Errorf("failed to save memory: %w", err)
	}
	
	// Save workspace
	if err := cm.saveJSONFile(filepath.Join(cm.basePath, "workspace/state.json"), cm.context.Workspace); err != nil {
		return fmt.Errorf("failed to save workspace: %w", err)
	}
	
	// Save knowledge
	if err := cm.saveJSONFile(filepath.Join(cm.basePath, "knowledge/patterns.json"), cm.context.Knowledge); err != nil {
		return fmt.Errorf("failed to save knowledge: %w", err)
	}
	
	// Save metadata
	if err := cm.saveJSONFile(filepath.Join(cm.basePath, "metadata.json"), cm.context.Metadata); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}
	
	return nil
}

// LoadContext loads context from disk
func (cm *ContextManager) LoadContext() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	// Load memory
	if err := cm.loadJSONFile(filepath.Join(cm.basePath, "memory/current.json"), &cm.context.Memory); err != nil {
		// If file doesn't exist, that's okay - start with empty memory
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to load memory: %w", err)
		}
	}
	
	// Load workspace
	if err := cm.loadJSONFile(filepath.Join(cm.basePath, "workspace/state.json"), &cm.context.Workspace); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to load workspace: %w", err)
		}
	}
	
	// Load knowledge
	if err := cm.loadJSONFile(filepath.Join(cm.basePath, "knowledge/patterns.json"), &cm.context.Knowledge); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to load knowledge: %w", err)
		}
	}
	
	// Load metadata
	if err := cm.loadJSONFile(filepath.Join(cm.basePath, "metadata.json"), &cm.context.Metadata); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to load metadata: %w", err)
		}
	}
	
	return nil
}

// saveJSONFile saves data to a JSON file
func (cm *ContextManager) saveJSONFile(path string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(path, jsonData, 0644)
}

// loadJSONFile loads data from a JSON file
func (cm *ContextManager) loadJSONFile(path string, data interface{}) error {
	fileData, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(fileData, data)
}

// generateSessionID creates a unique session identifier
func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().Unix())
}

// GetStats returns context statistics
func (cm *ContextManager) GetStats() ContextStats {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	return ContextStats{
		MemoryEntries:    len(cm.context.Memory),
		KnowledgeEntries: len(cm.context.Knowledge),
		CommandHistory:   len(cm.context.Workspace.CommandHistory),
		ActiveHandles:    len(cm.context.Workspace.ActiveHandles),
		ActiveFiles:      len(cm.context.Workspace.ActiveFiles),
		LastActivity:     cm.context.Workspace.LastActivity,
		SessionID:        cm.context.Metadata.SessionID,
		CreatedAt:        cm.context.Metadata.Created,
		LastUpdated:      cm.context.Metadata.LastUpdated,
	}
}

// ContextStats provides statistical information about the context
type ContextStats struct {
	MemoryEntries    int       `json:"memory_entries"`
	KnowledgeEntries int       `json:"knowledge_entries"`
	CommandHistory   int       `json:"command_history"`
	ActiveHandles    int       `json:"active_handles"`
	ActiveFiles      int       `json:"active_files"`
	LastActivity     time.Time `json:"last_activity"`
	SessionID        string    `json:"session_id"`
	CreatedAt        time.Time `json:"created_at"`
	LastUpdated      time.Time `json:"last_updated"`
} 