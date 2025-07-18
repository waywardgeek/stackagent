#!/bin/bash

echo "🖥️  StackAgent Desktop App Demo"
echo "================================="
echo ""

echo "This demo shows how to run StackAgent as a desktop application using Electron."
echo "The desktop app provides a native experience while maintaining all web features."
echo ""

echo "📊 Desktop App Features:"
echo "   ✅ Native desktop window with system integration"
echo "   ✅ Application menu with keyboard shortcuts"
echo "   ✅ System tray integration (planned)"
echo "   ✅ Native notifications"
echo "   ✅ Cross-platform (Windows, macOS, Linux)"
echo "   ✅ Same GUI as web version"
echo "   ✅ WebSocket communication with backend"
echo ""

echo "🔧 Building and Running:"
echo ""

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js first."
    echo "   Visit: https://nodejs.org/"
    exit 1
fi

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    echo "❌ npm is not installed. Please install npm first."
    exit 1
fi

echo "1. Installing dependencies..."
cd web/gui
if [ ! -d "node_modules" ]; then
    echo "   📦 Installing npm packages..."
    npm install
    if [ $? -eq 0 ]; then
        echo "   ✅ Dependencies installed successfully"
    else
        echo "   ❌ Failed to install dependencies"
        exit 1
    fi
else
    echo "   ✅ Dependencies already installed"
fi

echo ""
echo "2. Building the frontend..."
echo "   🔨 Building React app..."
npm run build
if [ $? -eq 0 ]; then
    echo "   ✅ Frontend built successfully"
else
    echo "   ❌ Failed to build frontend"
    exit 1
fi

echo ""
echo "3. Starting the backend server..."
cd ../..

# Install Go WebSocket dependency
echo "   📦 Installing Go dependencies..."
go mod tidy
if [ $? -eq 0 ]; then
    echo "   ✅ Go dependencies installed"
else
    echo "   ❌ Failed to install Go dependencies"
    exit 1
fi

echo ""
echo "   🚀 Starting StackAgent Server..."
echo "   📡 WebSocket server will run on port 8080"
echo ""

# Start the server in background
go run cmd/stackagent-server/main.go &
SERVER_PID=$!

echo "   ✅ Server started (PID: $SERVER_PID)"
echo ""

# Wait for server to start
sleep 3

echo "4. Launching Desktop App..."
cd web/gui
echo "   🖥️  Opening StackAgent Desktop App..."
echo ""

# Start the Electron app
npm run electron &
ELECTRON_PID=$!

echo "   ✅ Desktop app launched (PID: $ELECTRON_PID)"
echo ""

echo "🎉 Desktop App Demo Complete!"
echo ""
echo "🖥️  Desktop App Features:"
echo "   • Native window with minimize/maximize/close buttons"
echo "   • Application menu with File, Edit, View, StackAgent, Help"
echo "   • Keyboard shortcuts (Ctrl+N for new session, Ctrl+S to save, etc.)"
echo "   • Same dual-pane interface as web version"
echo "   • WebSocket connection to backend"
echo ""

echo "⌨️  Desktop Menu Shortcuts:"
echo "   • Ctrl+N: New Session"
echo "   • Ctrl+O: Open Context"
echo "   • Ctrl+S: Save Context"
echo "   • Ctrl+Shift+C: Clear All"
echo "   • Ctrl+Shift+T: Toggle Theme"
echo "   • Ctrl+B: Toggle Sidebar"
echo "   • Ctrl+Q: Quit App"
echo ""

echo "🔧 Development Commands:"
echo "   • npm run electron:dev: Start dev app with hot reload"
echo "   • npm run electron:pack: Package app for distribution"
echo "   • npm run electron:dist: Build installer"
echo ""

echo "🌐 For debugging, you can still use the web version:"
echo "   • Open http://localhost:8080 in your browser"
echo "   • Use browser dev tools for debugging"
echo ""

echo "Press Ctrl+C to stop the demo..."
echo ""

# Function to cleanup processes
cleanup() {
    echo ""
    echo "🛑 Stopping demo..."
    
    # Kill Electron app
    if kill -0 $ELECTRON_PID 2>/dev/null; then
        kill $ELECTRON_PID
        echo "   ✅ Desktop app stopped"
    fi
    
    # Kill server
    if kill -0 $SERVER_PID 2>/dev/null; then
        kill $SERVER_PID
        echo "   ✅ Server stopped"
    fi
    
    echo "   🎉 Demo cleanup complete"
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

# Wait for user to stop the demo
wait 