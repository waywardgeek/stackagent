package shell

import (
	"strings"
	"testing"
	"time"
)

func TestNewShellManager(t *testing.T) {
	sm := NewShellManager()
	
	if sm == nil {
		t.Fatal("NewShellManager returned nil")
	}
	
	if sm.nextID != 1 {
		t.Errorf("Expected nextID to be 1, got %d", sm.nextID)
	}
	
	if len(sm.handles) != 0 {
		t.Errorf("Expected empty handles map, got %d items", len(sm.handles))
	}
}

func TestRunWithCapture(t *testing.T) {
	sm := NewShellManager()
	
	// Test basic command
	handle, err := sm.RunWithCapture("echo Hello World")
	if err != nil {
		t.Fatalf("RunWithCapture failed: %v", err)
	}
	
	// Check handle properties
	if handle.ID != 2 {
		t.Errorf("Expected handle ID 2, got %d", handle.ID)
	}
	
	if handle.Command != "echo Hello World" {
		t.Errorf("Expected command 'echo Hello World', got '%s'", handle.Command)
	}
	
	// Wait for command to complete
	time.Sleep(100 * time.Millisecond)
	
	if !handle.Complete {
		t.Error("Expected command to be complete")
	}
	
	if handle.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", handle.ExitCode)
	}
	
	if len(handle.Buffer) == 0 {
		t.Error("Expected output in buffer")
	}
	
	if handle.Buffer[0] != "Hello World" {
		t.Errorf("Expected 'Hello World' in output, got '%s'", handle.Buffer[0])
	}
}

func TestSequentialIDs(t *testing.T) {
	sm := NewShellManager()
	
	// Create multiple handles
	handle1, err1 := sm.RunWithCapture("echo First")
	handle2, err2 := sm.RunWithCapture("echo Second")
	handle3, err3 := sm.RunWithCapture("echo Third")
	
	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatal("Failed to create handles")
	}
	
	// Check sequential IDs
	if handle1.ID != 2 || handle2.ID != 3 || handle3.ID != 4 {
		t.Errorf("Expected sequential IDs [2,3,4], got [%d,%d,%d]", 
			handle1.ID, handle2.ID, handle3.ID)
	}
	
	// Verify they're sequential
	if handle2.ID != handle1.ID+1 || handle3.ID != handle2.ID+1 {
		t.Error("IDs are not sequential")
	}
}

func TestSearchOutput(t *testing.T) {
	sm := NewShellManager()
	
	// Run command with known output
	handle, err := sm.RunWithCapture("echo Hello World")
	if err != nil {
		t.Fatalf("RunWithCapture failed: %v", err)
	}
	
	// Wait for completion
	time.Sleep(100 * time.Millisecond)
	
	// Test search functionality
	matches, err := sm.SearchOutput(handle.ID, "Hello")
	if err != nil {
		t.Fatalf("SearchOutput failed: %v", err)
	}
	
	if len(matches) != 1 {
		t.Errorf("Expected 1 match, got %d", len(matches))
	}
	
	if matches[0].LineNumber != 1 {
		t.Errorf("Expected line number 1, got %d", matches[0].LineNumber)
	}
	
	if matches[0].Line != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", matches[0].Line)
	}
	
	// Test non-existent pattern
	matches, err = sm.SearchOutput(handle.ID, "NonExistent")
	if err != nil {
		t.Fatalf("SearchOutput failed: %v", err)
	}
	
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches for non-existent pattern, got %d", len(matches))
	}
}

func TestGetTail(t *testing.T) {
	sm := NewShellManager()
	
	// Create multi-line output
	handle, err := sm.RunWithCapture("echo -e 'Line 1\\nLine 2\\nLine 3'")
	if err != nil {
		t.Fatalf("RunWithCapture failed: %v", err)
	}
	
	// Wait for completion
	time.Sleep(100 * time.Millisecond)
	
	// Test tail functionality
	tail, err := sm.GetTail(handle.ID, 2)
	if err != nil {
		t.Fatalf("GetTail failed: %v", err)
	}
	
	lines := strings.Split(tail, "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines in tail, got %d", len(lines))
	}
	
	// Test getting more lines than available
	tail, err = sm.GetTail(handle.ID, 10)
	if err != nil {
		t.Fatalf("GetTail failed: %v", err)
	}
	
	allLines := strings.Split(tail, "\n")
	if len(allLines) != 3 {
		t.Errorf("Expected 3 lines when requesting more than available, got %d", len(allLines))
	}
}

func TestGetStats(t *testing.T) {
	sm := NewShellManager()
	
	handle, err := sm.RunWithCapture("echo Test")
	if err != nil {
		t.Fatalf("RunWithCapture failed: %v", err)
	}
	
	// Wait for completion
	time.Sleep(100 * time.Millisecond)
	
	stats, err := sm.GetStats(handle.ID)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}
	
	if stats.LineCount != 1 {
		t.Errorf("Expected 1 line, got %d", stats.LineCount)
	}
	
	if !stats.Complete {
		t.Error("Expected command to be complete")
	}
	
	if stats.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", stats.ExitCode)
	}
	
	if stats.Duration <= 0 {
		t.Error("Expected positive duration")
	}
}

func TestListHandles(t *testing.T) {
	sm := NewShellManager()
	
	// Initially should be empty
	handles := sm.ListHandles()
	if len(handles) != 0 {
		t.Errorf("Expected 0 handles initially, got %d", len(handles))
	}
	
	// Create some handles
	handle1, _ := sm.RunWithCapture("echo First")
	handle2, _ := sm.RunWithCapture("echo Second")
	
	handles = sm.ListHandles()
	if len(handles) != 2 {
		t.Errorf("Expected 2 handles, got %d", len(handles))
	}
	
	// Check the handles are in the list
	found1, found2 := false, false
	for _, id := range handles {
		if id == handle1.ID {
			found1 = true
		}
		if id == handle2.ID {
			found2 = true
		}
	}
	
	if !found1 || !found2 {
		t.Error("Not all handles found in list")
	}
}

func TestCleanupHandle(t *testing.T) {
	sm := NewShellManager()
	
	handle, err := sm.RunWithCapture("echo Test")
	if err != nil {
		t.Fatalf("RunWithCapture failed: %v", err)
	}
	
	// Verify handle exists
	handles := sm.ListHandles()
	if len(handles) != 1 {
		t.Errorf("Expected 1 handle, got %d", len(handles))
	}
	
	// Cleanup handle
	err = sm.CleanupHandle(handle.ID)
	if err != nil {
		t.Fatalf("CleanupHandle failed: %v", err)
	}
	
	// Verify handle is gone
	handles = sm.ListHandles()
	if len(handles) != 0 {
		t.Errorf("Expected 0 handles after cleanup, got %d", len(handles))
	}
	
	// Try to cleanup non-existent handle
	err = sm.CleanupHandle(999)
	if err == nil {
		t.Error("Expected error when cleaning up non-existent handle")
	}
}

func TestErrorHandling(t *testing.T) {
	sm := NewShellManager()
	
	// Test operations on non-existent handle
	nonExistentID := uint64(999)
	
	_, err := sm.SearchOutput(nonExistentID, "test")
	if err == nil {
		t.Error("Expected error for non-existent handle in SearchOutput")
	}
	
	_, err = sm.GetTail(nonExistentID, 5)
	if err == nil {
		t.Error("Expected error for non-existent handle in GetTail")
	}
	
	_, err = sm.GetStats(nonExistentID)
	if err == nil {
		t.Error("Expected error for non-existent handle in GetStats")
	}
	
	_, err = sm.ReadLines(nonExistentID, 1, 5)
	if err == nil {
		t.Error("Expected error for non-existent handle in ReadLines")
	}
}

func TestConcurrency(t *testing.T) {
	sm := NewShellManager()
	
	// Test concurrent handle creation
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			handle, err := sm.RunWithCapture("echo Concurrent")
			if err != nil {
				t.Errorf("Concurrent RunWithCapture failed: %v", err)
			}
			
			// Verify handle is valid
			if handle.ID == 0 {
				t.Error("Invalid handle ID in concurrent test")
			}
			
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify we have 10 handles
	handles := sm.ListHandles()
	if len(handles) != 10 {
		t.Errorf("Expected 10 handles from concurrent test, got %d", len(handles))
	}
} 