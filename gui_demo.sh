#!/bin/bash

echo "🚀 StackAgent GUI Demo - Revolutionary AI Interface"
echo "===================================================="
echo ""

echo "This demo showcases the complete GUI implementation for StackAgent."
echo "The GUI features a dual-pane layout with real-time WebSocket communication,"
echo "context management, and transparent AI function calls."
echo ""

echo "📊 GUI Features Implemented:"
echo "   ✅ Dual-pane layout with resizable divider"
echo "   ✅ Real-time WebSocket communication"
echo "   ✅ Chat interface with message history"
echo "   ✅ Function call widgets with status indicators"
echo "   ✅ Context browser showing memory and knowledge"
echo "   ✅ Command output display with syntax highlighting"
echo "   ✅ Dark/light theme support"
echo "   ✅ Keyboard shortcuts for power users"
echo "   ✅ Status bar with context statistics"
echo "   ✅ Connection status indicators"
echo "   ✅ Cost tracking and session management"
echo ""

echo "🛠️ Technology Stack:"
echo "   • Frontend: React + TypeScript + Tailwind CSS"
echo "   • State Management: Zustand with Immer"
echo "   • Build Tool: Vite"
echo "   • WebSocket: Gorilla WebSocket (Go)"
echo "   • UI Components: Lucide React icons"
echo "   • Styling: CSS custom properties + Tailwind"
echo ""

echo "📁 Project Structure:"
echo "   web/gui/                 - React frontend"
echo "   ├── src/                - Source code"
echo "   │   ├── components/     - React components"
echo "   │   ├── hooks/          - Custom hooks"
echo "   │   ├── store/          - State management"
echo "   │   └── types/          - TypeScript types"
echo "   ├── public/             - Static assets"
echo "   └── dist/               - Build output"
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

echo "1. Installing frontend dependencies..."
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
echo "3. Starting the server..."
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
echo "   🌐 GUI will be available at http://localhost:8080"
echo ""

# Start the server
go run cmd/stackagent-server/main.go &
SERVER_PID=$!

echo "   ✅ Server started (PID: $SERVER_PID)"
echo ""

echo "🎉 GUI Demo Complete!"
echo ""
echo "🔗 URLs:"
echo "   • GUI: http://localhost:8080"
echo "   • WebSocket: ws://localhost:8080/ws"
echo "   • Health Check: http://localhost:8080/api/health"
echo ""

echo "🎮 Try These Features:"
echo "   • Resize the panes by dragging the divider"
echo "   • Use keyboard shortcuts (Ctrl+Shift+T for theme)"
echo "   • Click on function calls to see details"
echo "   • Browse the context in the right pane"
echo "   • Check connection status in the header"
echo ""

echo "⌨️  Keyboard Shortcuts:"
echo "   • Ctrl+Shift+T: Toggle theme"
echo "   • Ctrl+B: Toggle sidebar"
echo "   • Ctrl+K: Command palette"
echo "   • Ctrl+J: Navigate function calls"
echo "   • Ctrl+Shift+C: Clear all"
echo "   • /: Focus chat input"
echo ""

echo "🔧 Development Commands:"
echo "   • npm run dev: Start development server"
echo "   • npm run build: Build for production"
echo "   • npm run preview: Preview build"
echo "   • npm run lint: Run linter"
echo ""

echo "Press Ctrl+C to stop the server..."
echo "Or visit http://localhost:8080 to see the GUI!"

# Wait for user to stop the server
wait $SERVER_PID 