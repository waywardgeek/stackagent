#!/bin/bash

# Debug Logging Demo for StackAgent CLI
# This shows how to use the debug logging feature

echo "🔍 StackAgent CLI Debug Logging Demo"
echo "===================================="
echo

# Check if API key is set
if [ -z "$ANTHROPIC_API_KEY" ]; then
    echo "❌ ANTHROPIC_API_KEY not set. This demo will show debug logging setup only."
    echo "💡 Set your API key to see full request/response logging:"
    echo "   export ANTHROPIC_API_KEY=your_api_key_here"
    echo
fi

echo "1. Running StackAgent CLI with debug logging enabled:"
echo "   ./stackagent-cli -debug=claude_debug.txt"
echo

# Create a test session
echo "2. Demo session (type 'run ls' then 'quit'):"
echo -e "run ls\nquit" | ./stackagent-cli -debug=claude_debug.txt

echo
echo "3. Debug log contents:"
echo "   cat claude_debug.txt"
echo

# Show the debug log
if [ -f claude_debug.txt ]; then
    echo "--- DEBUG LOG START ---"
    cat claude_debug.txt
    echo "--- DEBUG LOG END ---"
    echo
else
    echo "❌ Debug log file not found"
fi

echo "4. Features of debug logging:"
echo "   • Logs all API requests with timestamps"
echo "   • Logs all API responses with status codes"
echo "   • Redacts API keys for security"
echo "   • JSON formatted for easy parsing"
echo "   • Appends to file (doesn't overwrite)"
echo "   • Automatic file sync for real-time viewing"
echo

echo "5. Command line options:"
echo "   -debug=filename.txt    # Enable debug logging to file"
echo "   -d=filename.txt        # Short form"
echo

echo "6. Example usage with API calls:"
echo "   # Enable debug logging and analyze a command"
echo "   ./stackagent-cli -debug=api_debug.txt"
echo "   stackagent> run docker ps"
echo "   stackagent> analyze 2"
echo "   stackagent> quit"
echo

echo "✅ Debug logging demo complete!"
echo "💡 Use debug logs to:"
echo "   • Debug API issues"
echo "   • Monitor token usage"
echo "   • Analyze request/response patterns"
echo "   • Optimize prompts"

# Clean up
rm -f claude_debug.txt claude_io.txt 