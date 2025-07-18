# StackAgent GUI

Revolutionary AI coding assistant interface with transparent function calls and persistent context management.

## 🚀 Features

### Core Interface
- **Dual-Pane Layout**: Resizable left/right panes with draggable divider
- **Real-Time Communication**: WebSocket-based live updates
- **Responsive Design**: Works on different screen sizes
- **Dark/Light Themes**: User-configurable themes

### Chat Experience
- **Message History**: Scrollable conversation with AI
- **Function Call Widgets**: Interactive elements for each AI action
- **Syntax Highlighting**: Code blocks with proper formatting
- **Status Indicators**: Live status of running commands

### Transparency Features
- **Function Call Details**: Click any function to see arguments, results, errors
- **Command Output**: Real-time terminal output with syntax highlighting
- **Context Browser**: View protected memory, knowledge base, Git state
- **Cost Tracking**: Real-time API usage and cost monitoring

### Context Management
- **Protected Memory**: Persistent key-value storage
- **Knowledge Base**: AI-learned patterns and insights
- **Git Integration**: Branch-specific context switching
- **Session Management**: Track commands and AI decisions

### Power User Features
- **Keyboard Shortcuts**: Extensive shortcut support
- **Command Palette**: Quick access to all functions (Ctrl+K)
- **Search & Filter**: Find across sessions and contexts
- **Export/Import**: Save and share sessions

## 🛠️ Technology Stack

- **Frontend**: React 18 + TypeScript
- **State Management**: Zustand with Immer
- **Styling**: Tailwind CSS + Custom CSS
- **Build Tool**: Vite
- **Icons**: Lucide React
- **WebSocket**: Native WebSocket API
- **Date Handling**: date-fns
- **Notifications**: react-hot-toast

## 📁 Project Structure

```
web/gui/
├── src/
│   ├── components/          # React components
│   │   ├── Layout/         # Layout components
│   │   ├── Chat/           # Chat interface
│   │   └── Action/         # Action pane
│   ├── hooks/              # Custom React hooks
│   ├── store/              # Zustand state management
│   ├── types/              # TypeScript type definitions
│   └── utils/              # Utility functions
├── public/                 # Static assets
├── dist/                   # Build output
└── package.json           # Dependencies
```

## 🔧 Development

### Prerequisites
- Node.js 18+
- npm or yarn

### Setup
```bash
# Navigate to GUI directory
cd web/gui

# Install dependencies
npm install

# Start development server
npm run dev
```

### Available Scripts
- `npm run dev` - Start development server (port 5173)
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint
- `npm run type-check` - Run TypeScript checks

### Development Server
The development server runs on `http://localhost:5173` and proxies API calls to the Go backend on port 8080.

## 🎮 Usage

### Basic Navigation
- **Left Pane**: Chat interface with message history
- **Right Pane**: Function call details and context browser
- **Divider**: Drag to resize panes (20%-80% range)

### Keyboard Shortcuts
- `Ctrl+Shift+T` - Toggle dark/light theme
- `Ctrl+B` - Toggle sidebar
- `Ctrl+K` - Open command palette
- `Ctrl+J` - Navigate to next function call
- `Ctrl+Shift+J` - Navigate to previous function call
- `Ctrl+Shift+C` - Clear all messages
- `/` - Focus chat input
- `Ctrl+C` - Copy selected function result
- `Ctrl+F` - Search in chat
- `Ctrl+E` - Export session

### Function Call Interaction
1. **Click Function Widgets**: Select any function call in the left pane
2. **View Details**: See arguments, results, and timing in the right pane
3. **Copy Results**: Use keyboard shortcut or right-click menu
4. **Retry/Cancel**: Control running functions

### Context Management
- **Protected Memory**: Store data that persists across sessions
- **Knowledge Base**: AI-learned insights and patterns
- **Git Integration**: Context switches with branch changes
- **Session History**: Track all commands and AI interactions

## 🔌 Backend Integration

The GUI communicates with the Go backend via:

### WebSocket Events
- `session_started` - New session initialized
- `message_received` - AI response received
- `function_call_started` - Function execution began
- `function_call_completed` - Function finished successfully
- `function_call_failed` - Function failed with error
- `context_updated` - Context state changed
- `command_started` - Command execution began
- `command_completed` - Command finished
- `git_branch_changed` - Git branch switched
- `cost_updated` - API cost updated

### HTTP API
- `GET /api/health` - Server health check
- `POST /api/chat` - Send chat message
- `GET /api/context` - Get context state
- `POST /api/context` - Update context

## 🎨 Theming

The GUI supports both dark and light themes with:
- CSS custom properties for easy customization
- Tailwind CSS dark mode classes
- Persistent theme preference
- Smooth transitions between themes

### Color Scheme
- **Primary**: Blue (#0ea5e9)
- **Secondary**: Slate/Gray
- **Success**: Green (#22c55e)
- **Warning**: Amber (#f59e0b)
- **Error**: Red (#ef4444)

## 📊 State Management

Uses Zustand for state management with:
- **Immer**: Immutable state updates
- **Persistence**: LocalStorage for preferences
- **Subscriptions**: React to state changes
- **Selectors**: Efficient component updates

### Store Structure
```typescript
interface AppState {
  // UI State
  leftPaneWidth: number
  selectedFunctionCall?: string
  theme: 'light' | 'dark'
  
  // Session State
  messages: Message[]
  functionCalls: FunctionCall[]
  
  // Context State
  protectedMemory: Record<string, string>
  knowledgeBase: Record<string, string>
  
  // Connection State
  connected: boolean
  reconnectAttempts: number
}
```

## 🧪 Testing

```bash
# Run tests
npm test

# Run with coverage
npm run test:coverage

# Run in watch mode
npm run test:watch
```

## 🚀 Deployment

### Production Build
```bash
npm run build
```

### Docker
```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
```

## 🎯 Future Enhancements

- [ ] Multi-tab support for different projects
- [ ] Plugin system for custom function types
- [ ] Advanced search and filtering
- [ ] Collaboration features
- [ ] Mobile responsive design
- [ ] Offline mode support
- [ ] Custom themes and layouts

## 🐛 Troubleshooting

### Common Issues

**WebSocket Connection Failed**
- Ensure Go server is running on port 8080
- Check firewall settings
- Verify CORS configuration

**Build Failures**
- Clear node_modules and reinstall: `rm -rf node_modules && npm install`
- Check Node.js version: `node --version`
- Update dependencies: `npm update`

**Performance Issues**
- Enable React DevTools Profiler
- Check for memory leaks in WebSocket connections
- Optimize large function call lists

## 📄 License

This project is part of StackAgent and follows the same licensing terms. 