#!/bin/bash

echo "🚀 StackAgent Milestone 2 Demo - Context Persistence with Git Integration"
echo "=========================================================================="
echo ""

echo "Building the CLI..."
go build -o stackagent-cli cmd/stackagent-cli/main.go

echo ""
echo "1. Testing basic context management..."
echo "   - Setting protected memory"
echo "   - Checking context statistics"
echo "   - Saving context"
echo ""

echo -e "ctx memory project-name StackAgent\nctx memory description Revolutionary AI coding agent with persistent memory\nctx info\nctx save\nquit" | ./stackagent-cli

echo ""
echo "2. Testing command execution with context tracking..."
echo "   - Running commands"
echo "   - Viewing command history in context"
echo ""

echo -e "run ls -la\nrun whoami\nctx info\nquit" | ./stackagent-cli

echo ""
echo "3. Testing Git integration..."
echo "   - Checking branch context"
echo "   - Viewing Git synchronization"
echo ""

echo -e "ctx branches\nctx sync\nquit" | ./stackagent-cli

echo ""
echo "4. Showing context file structure..."
echo ""

echo "Context directory structure:"
find .stackagent/context -type f -name "*.json" | head -10

echo ""
echo "Sample protected memory:"
cat .stackagent/context/memory/current.json

echo ""
echo "Sample metadata:"
cat .stackagent/context/metadata.json | head -10

echo ""
echo "✅ Milestone 2 Demo Complete!"
echo ""
echo "🎉 Key Features Implemented:"
echo "   • Protected memory that survives sessions"
echo "   • Command history tracking with uint64 handles"
echo "   • Git branch-specific context switching"
echo "   • Persistent workspace state"
echo "   • JSON-based context storage"
echo "   • Thread-safe operations"
echo ""
echo "🔧 Available Commands:"
echo "   • ctx save/load - Manual context persistence"
echo "   • ctx memory - Protected memory management"
echo "   • ctx info - Context statistics and session info"
echo "   • ctx branches - List all branches with context"
echo "   • ctx sync - Synchronize with Git repository"
echo "   • ctx hooks - Setup Git hooks for auto-switching"
echo "   • ctx cleanup - Remove orphaned branch contexts"
echo ""
echo "🌟 Revolutionary Features:"
echo "   • AI memory persists across sessions and branches"
echo "   • Automatic context switching on Git branch changes"
echo "   • uint64 handles for scalable command tracking"
echo "   • Separate contexts for different Git branches"
echo "   • Integration with Claude AI for intelligent analysis"
echo ""
echo "Ready for Milestone 3! 🚀" 