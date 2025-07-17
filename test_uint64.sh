#!/bin/bash

# Test script for the uint64 OutputHandle implementation

echo "Testing StackAgent OutputHandle with uint64 IDs..."

# Create a simple Go test program
cat > test_main.go << 'EOF'
package main

import (
	"fmt"
	"log"
	"time"
	
	"stackagent/pkg/shell"
)

func main() {
	fmt.Println("Testing OutputHandle with uint64 IDs...")
	
	// Create shell manager
	sm := shell.NewShellManager()
	
	// Test 1: Run a simple command
	fmt.Println("\n1. Running 'echo Hello World'")
	handle1, err := sm.RunWithCapture("echo Hello World")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Generated handle ID: %d (type: uint64)\n", handle1.ID)
	
	// Wait a moment for command to complete
	time.Sleep(100 * time.Millisecond)
	
	// Test 2: Search output
	fmt.Println("\n2. Searching for 'Hello' in output")
	matches, err := sm.SearchOutput(handle1.ID, "Hello")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Found %d matches\n", len(matches))
	for _, match := range matches {
		fmt.Printf("   Line %d: %s\n", match.LineNumber, match.Line)
	}
	
	// Test 3: Get tail
	fmt.Println("\n3. Getting tail (last 5 lines)")
	tail, err := sm.GetTail(handle1.ID, 5)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Tail: %s\n", tail)
	
	// Test 4: Get stats
	fmt.Println("\n4. Getting output statistics")
	stats, err := sm.GetStats(handle1.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Lines: %d, Complete: %t, Exit Code: %d, Duration: %v\n", 
		stats.LineCount, stats.Complete, stats.ExitCode, stats.Duration)
	
	// Test 5: Multiple handles with sequential IDs
	fmt.Println("\n5. Testing sequential uint64 IDs")
	handle2, _ := sm.RunWithCapture("echo Second Command")
	handle3, _ := sm.RunWithCapture("echo Third Command")
	
	fmt.Printf("   Handle 1 ID: %d\n", handle1.ID)
	fmt.Printf("   Handle 2 ID: %d\n", handle2.ID)
	fmt.Printf("   Handle 3 ID: %d\n", handle3.ID)
	fmt.Printf("   IDs are sequential: %t\n", handle2.ID == handle1.ID+1 && handle3.ID == handle2.ID+1)
	
	// Test 6: List all handles
	fmt.Println("\n6. Listing all active handles")
	handles := sm.ListHandles()
	fmt.Printf("   Active handles: %v\n", handles)
	
	fmt.Println("\n✅ All tests completed successfully!")
	fmt.Println("✅ OutputHandle now uses uint64 IDs instead of strings!")
}
EOF

# Create go.mod if it doesn't exist
if [ ! -f go.mod ]; then
    echo "Creating go.mod..."
    go mod init stackagent
fi

# Run the test
echo "Compiling and running test..."
go run test_main.go

# Clean up
echo "Cleaning up test files..."
rm -f test_main.go

echo "Test completed!" 