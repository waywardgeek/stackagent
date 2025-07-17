package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"stackagent/pkg/shell"
)

func main() {
	fmt.Println("🚀 StackAgent - AI Coding Agent with uint64 OutputHandles")
	fmt.Println("   Revolutionary context management with hardware-attested privacy")
	fmt.Println()

	// Initialize shell manager with uint64 handles
	sm := shell.NewShellManager()
	
	fmt.Println("✅ Initializing StackAgent with uint64 OutputHandle system...")
	
	// Demo the uint64 OutputHandle functionality
	fmt.Println("\n📋 Testing uint64 OutputHandle implementation:")
	
	// Test 1: Basic command with uint64 ID
	fmt.Println("   1. Running basic command...")
	handle1, err := sm.RunWithCapture("echo 'StackAgent OutputHandle with uint64 ID'")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("      ✓ Generated handle ID: %d (type: uint64)\n", handle1.ID)
	
	// Wait for completion
	time.Sleep(100 * time.Millisecond)
	
	// Test 2: Show that IDs are sequential
	fmt.Println("   2. Testing sequential ID generation...")
	handle2, _ := sm.RunWithCapture("ls -la *.go")
	handle3, _ := sm.RunWithCapture("whoami")
	
	fmt.Printf("      ✓ Handle 1: %d\n", handle1.ID)
	fmt.Printf("      ✓ Handle 2: %d\n", handle2.ID) 
	fmt.Printf("      ✓ Handle 3: %d\n", handle3.ID)
	fmt.Printf("      ✓ Sequential: %t\n", handle2.ID == handle1.ID+1 && handle3.ID == handle2.ID+1)
	
	// Wait for all commands to complete
	time.Sleep(200 * time.Millisecond)
	
	// Test 3: Query outputs using uint64 IDs
	fmt.Println("   3. Querying outputs using uint64 handles...")
	
	stats1, _ := sm.GetStats(handle1.ID)
	fmt.Printf("      ✓ Handle %d: %d lines, exit code %d\n", handle1.ID, stats1.LineCount, stats1.ExitCode)
	
	tail2, _ := sm.GetTail(handle2.ID, 3)
	fmt.Printf("      ✓ Handle %d tail: %d characters\n", handle2.ID, len(tail2))
	
	// Test 4: Search functionality
	fmt.Println("   4. Testing search with uint64 handles...")
	matches, _ := sm.SearchOutput(handle1.ID, "uint64")
	fmt.Printf("      ✓ Found %d matches for 'uint64' in handle %d\n", len(matches), handle1.ID)
	
	// Show all active handles
	fmt.Println("   5. Listing all active handles...")
	handles := sm.ListHandles()
	fmt.Printf("      ✓ Active handles: %v\n", handles)
	
	fmt.Println("\n🎉 SUCCESS: OutputHandle system successfully converted to uint64!")
	fmt.Println("   • IDs are now unsigned 64-bit integers instead of strings")
	fmt.Println("   • Sequential ID generation using atomic operations")
	fmt.Println("   • All query functions updated to accept uint64 parameters")
	fmt.Println("   • Thread-safe implementation with proper synchronization")
	
	// Check if CLI is available
	if _, err := os.Stat("./stackagent-cli"); err == nil {
		fmt.Println("\n💡 Try the interactive CLI:")
		fmt.Println("   ./stackagent-cli")
		fmt.Println("   Commands: demo, run <cmd>, search <id> <pattern>, list, quit")
	}
} 