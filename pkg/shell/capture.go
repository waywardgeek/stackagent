package shell

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// OutputHandle represents a handle to command output that can be queried
type OutputHandle struct {
	ID        uint64    // Changed from string to uint64
	Command   string
	Buffer    []string  // Start simple with string slice
	Complete  bool
	ExitCode  int
	StartTime time.Time
	EndTime   *time.Time
	mutex     sync.RWMutex
}

// Stats provides statistics about the output
type Stats struct {
	LineCount int
	Duration  time.Duration
	Complete  bool
	ExitCode  int
}

// Match represents a search match in the output
type Match struct {
	LineNumber int
	Line       string
	Context    []string // Lines before/after for context
}

// ShellManager manages shell sessions and output handles
type ShellManager struct {
	handles   map[uint64]*OutputHandle
	nextID    uint64 // Atomic counter for generating IDs
	mutex     sync.RWMutex
}

// NewShellManager creates a new shell manager
func NewShellManager() *ShellManager {
	return &ShellManager{
		handles: make(map[uint64]*OutputHandle),
		nextID:  1, // Start from 1, 0 can be reserved for invalid/null
	}
}

// RunWithCapture executes a command and returns a handle for querying output
func (sm *ShellManager) RunWithCapture(cmd string) (*OutputHandle, error) {
	// Generate unique uint64 ID
	id := atomic.AddUint64(&sm.nextID, 1)
	
	handle := &OutputHandle{
		ID:        id,
		Command:   cmd,
		Buffer:    []string{},
		StartTime: time.Now(),
	}
	
	// Store handle immediately
	sm.mutex.Lock()
	sm.handles[handle.ID] = handle
	sm.mutex.Unlock()
	
	// Execute command
	execCmd := exec.Command("bash", "-c", cmd)
	stdout, err := execCmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	stderr, err := execCmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	err = execCmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}
	
	// Capture output in goroutines
	var wg sync.WaitGroup
	wg.Add(2)
	
	// Capture stdout
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			handle.mutex.Lock()
			handle.Buffer = append(handle.Buffer, scanner.Text())
			handle.mutex.Unlock()
		}
	}()
	
	// Capture stderr
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			handle.mutex.Lock()
			handle.Buffer = append(handle.Buffer, "STDERR: "+scanner.Text())
			handle.mutex.Unlock()
		}
	}()
	
	// Wait for command completion in background
	go func() {
		wg.Wait()
		err := execCmd.Wait()
		
		handle.mutex.Lock()
		handle.Complete = true
		endTime := time.Now()
		handle.EndTime = &endTime
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				handle.ExitCode = exitError.ExitCode()
			} else {
				handle.ExitCode = -1
			}
		} else {
			handle.ExitCode = 0
		}
		handle.mutex.Unlock()
	}()
	
	return handle, nil
}

// SearchOutput searches for a pattern in the output and returns matches
func (sm *ShellManager) SearchOutput(handleID uint64, pattern string) ([]Match, error) {
	sm.mutex.RLock()
	handle, exists := sm.handles[handleID]
	sm.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("handle %d not found", handleID)
	}
	
	handle.mutex.RLock()
	defer handle.mutex.RUnlock()
	
	var matches []Match
	for i, line := range handle.Buffer {
		if strings.Contains(line, pattern) {
			match := Match{
				LineNumber: i + 1, // 1-indexed
				Line:       line,
				Context:    getContext(handle.Buffer, i, 2), // 2 lines context
			}
			matches = append(matches, match)
		}
	}
	
	return matches, nil
}

// ReadLines returns specific lines from the output
func (sm *ShellManager) ReadLines(handleID uint64, start, end int) (string, error) {
	sm.mutex.RLock()
	handle, exists := sm.handles[handleID]
	sm.mutex.RUnlock()
	
	if !exists {
		return "", fmt.Errorf("handle %d not found", handleID)
	}
	
	handle.mutex.RLock()
	defer handle.mutex.RUnlock()
	
	if start < 1 || start > len(handle.Buffer) {
		return "", fmt.Errorf("start line %d out of range (1-%d)", start, len(handle.Buffer))
	}
	
	if end < start || end > len(handle.Buffer) {
		end = len(handle.Buffer)
	}
	
	// Convert to 0-indexed
	startIdx := start - 1
	endIdx := end
	
	return strings.Join(handle.Buffer[startIdx:endIdx], "\n"), nil
}

// GetTail returns the last N lines of output
func (sm *ShellManager) GetTail(handleID uint64, lines int) (string, error) {
	sm.mutex.RLock()
	handle, exists := sm.handles[handleID]
	sm.mutex.RUnlock()
	
	if !exists {
		return "", fmt.Errorf("handle %d not found", handleID)
	}
	
	handle.mutex.RLock()
	defer handle.mutex.RUnlock()
	
	bufferLen := len(handle.Buffer)
	if lines >= bufferLen {
		return strings.Join(handle.Buffer, "\n"), nil
	}
	
	start := bufferLen - lines
	return strings.Join(handle.Buffer[start:], "\n"), nil
}

// GetStats returns statistics about the command output
func (sm *ShellManager) GetStats(handleID uint64) (*Stats, error) {
	sm.mutex.RLock()
	handle, exists := sm.handles[handleID]
	sm.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("handle %d not found", handleID)
	}
	
	handle.mutex.RLock()
	defer handle.mutex.RUnlock()
	
	stats := &Stats{
		LineCount: len(handle.Buffer),
		Complete:  handle.Complete,
		ExitCode:  handle.ExitCode,
	}
	
	if handle.EndTime != nil {
		stats.Duration = handle.EndTime.Sub(handle.StartTime)
	} else {
		stats.Duration = time.Since(handle.StartTime)
	}
	
	return stats, nil
}

// CleanupHandle removes a handle from memory
func (sm *ShellManager) CleanupHandle(handleID uint64) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	if _, exists := sm.handles[handleID]; !exists {
		return fmt.Errorf("handle %d not found", handleID)
	}
	
	delete(sm.handles, handleID)
	return nil
}

// ListHandles returns all active handle IDs
func (sm *ShellManager) ListHandles() []uint64 {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	var handles []uint64
	for id := range sm.handles {
		handles = append(handles, id)
	}
	
	return handles
}

// Helper function to get context lines around a match
func getContext(lines []string, center, contextSize int) []string {
	start := center - contextSize
	if start < 0 {
		start = 0
	}
	
	end := center + contextSize + 1
	if end > len(lines) {
		end = len(lines)
	}
	
	return lines[start:end]
} 