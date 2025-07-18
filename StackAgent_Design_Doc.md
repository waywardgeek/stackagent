# StackAgent Technical Design Document

## Executive Summary

StackAgent revolutionizes AI coding agents by enabling full autonomy through a
paradigm shift from prevention-based to detection-based security. StackAgent
allows AI agents to interact directly with web interfaces, shell environments,
and the desktop while ensuring privacy through radical transparency, using
hardware-rooted attestation of detection workloads.

Built on the KISS principle learned from writing 80K lines of production code
in 30 days: reliable Unix tools over complex abstractions, version control for
AI context sharing, and letting the AI tell us what features it needs through
actual usage.

**Key Innovations:**
- AI-managed context windows with 90% cost reduction
- Persistent memory across sessions
- Git-native context storage and sharing
- Talk to past versions of AI or different models
- Hardware-attested privacy protection

## Table of Contents

1. [System Architecture Overview](#1-system-architecture-overview)
2. [Core Innovations](#2-core-innovations)
3. [User Interface Design](#3-user-interface-design)
4. [Security Architecture](#4-security-architecture)
5. [Implementation Roadmap](#5-implementation-roadmap)
6. [Technical Implementation](#6-technical-implementation)
7. [Technology Stack](#7-technology-stack)
8. [Deployment Architecture](#8-deployment-architecture)

## 1. System Architecture Overview

### 1.1 Design Philosophy: KISS (Keep It Simple, Stupid)

Based on successfully building 80K+ lines of production code in 30 days with a
partially malfunctioning Cursor instance, StackAgent embraces simplicity:

**Core Principles:**
1. **Unix tools over complex abstractions** - grep, sed, find always work
2. **Direct access over APIs** - Real PTYs, real files, real processes  
3. **Multiple simple tools over one complex tool** - Each component does one thing well
4. **Fallback to shell** - When fancy features fail, bash saves the day
5. **Proven tech over cutting edge** - Go, Unix utils, not the latest framework

**What We're NOT Building:**
- Complex AST-based code analysis (grep works better)
- "Smart" features that work 80% of the time
- Language server protocol integrations  
- Fancy caching layers that break edge cases
- Abstract file systems or process managers
- Cloud storage for context (git works better)
- Every possible helper function upfront (AI will tell us what's needed)

### 1.2 High-Level Architecture

```
USER'S MACHINE
┌─────────────────────────────────────────────────────────────────┐
│                      StackAgent Client                          │
├─────────────────────────────────────────────────────────────────┤
│  Core Capabilities              │  Context Management           │
│  ┌────────────────────────┐    │  ┌───────────────────────┐   │
│  │ • Shell Multiplexing    │    │  │ • Git-based Storage   │   │
│  │ • Desktop Control       │    │  │ • Cost Optimization   │   │
│  │ • Secrets Oracle        │    │  │ • AI Process Forking  │   │
│  │ • File Operations       │    │  │ • Cache Management    │   │
│  └────────────────────────┘    │  └───────────────────────┘   │
├─────────────────────────────────────────────────────────────────┤
│                    Security Components                          │
│  ┌─────────────────┐  ┌──────────────────────────────────────┐ │
│  │ Secrets Oracle  │  │ EDR Endpoint (Signal Collector)      │ │
│  │ (No AI access)  │  │ - Monitors all agent actions         │ │
│  └─────────────────┘  │ - Encrypts signals locally           │ │
│                       │ - Sends to remote CVMs for analysis  │ │
│                       └──────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                                    │
                        Encrypted Signals Only
                                    ↓
STACKAGENT CLOUD (Privacy-Preserving Analysis)
┌─────────────────────────────────────────────────────────────────┐
│              Confidential Virtual Machines (CVMs)               │
│  Running signal analysis with hardware attestation              │
│  StackAgent operators have zero access to user data             │
└─────────────────────────────────────────────────────────────────┘
                                    │
                         Attestation Verification
                                    ↓
OPENADP DISTRIBUTED TRUST NETWORK
┌─────────────────────────────────────────────────────────────────┐
│  Multiple validators verify CVM integrity before releasing      │
│  decryption keys - no single point of trust                    │
└─────────────────────────────────────────────────────────────────┘
```

### 1.3 Key Components

#### AI Agent Core
- Pluggable LLM backend (supports multiple providers)
- Context-aware command execution with output handles
- Git-native context storage and versioning
- AI process forking for querying past states or different models

#### Shell Environment Manager
- Real PTY support for proper terminal emulation
- Multiple concurrent shell sessions
- Interactive prompt detection (passwords, confirmations)
- Direct command execution (no abstraction layers)

#### Context Management System
- **AI-controlled context windows** - AI manages its own memory
- **Git-based persistence** - Context stored in `.stackagent/` directory
- **Cost optimization** - 90% reduction through smart caching
- **Process forking** - Query past versions without corrupting current state

#### Security Components
- **Secrets Oracle** - AI references secrets by name, never sees values
- **EDR Endpoint** - Local signal collection, encrypted before transmission
- **CVM Analysis** - All analysis in hardware-attested VMs
- **OpenADP Integration** - Distributed trust for attestation verification

## 2. Core Innovations

### 2.1 AI-First Context Management

**Problem**: Current AI agents waste 90% of context on noise, costing users $$$ per query.

**Solution**: AI controls its own context window with purpose-built tools.

```go
// Instead of dumping everything:
// docker logs app | tail -100 | grep error | head -20

// StackAgent approach:
output := agent.RunWithCapture("docker logs app", 
    WithTimeout(30*time.Second),
    WithCallback(5*time.Second))  // Get updates during execution

// AI can now query output like a file:
errors := agent.SearchOutput(output.ID, "error")     // Find ALL errors
context := agent.ReadLines(output.ID, 50, 100)      // Read specific lines
summary := agent.GetSummary(output.ID)              // Get AI-friendly summary

// Cost impact: $0.09 per query instead of $1.05
```

### 2.2 Persistent Working Memory

**Problem**: AI agents forget everything between sessions - like developers with amnesia.

**Solution**: Protected context that persists across sessions, stored in git.

```go
// Monday: Working on auth system
agent.SetProtected("current-task", "auth-implementation")
agent.StoreContext("jwt-research", deepDiveResults)

// Friday: Pick up exactly where you left off
task := agent.GetProtected("current-task")
research := agent.SearchContext("jwt-research", "refresh tokens")
```

### 2.3 Git-Native Context Storage

**Problem**: How to share AI knowledge across team/time?

**Solution**: Store context in version control, just like code.

```
.stackagent/
├── context/
│   ├── memory/
│   │   ├── current.md          # Current working memory
│   │   ├── investigations/     # Past investigation notes
│   │   └── project-plan.md     # Long-term roadmap
│   ├── workspace/
│   │   ├── active-files.json   # Open files & positions
│   │   └── command-history.json # Recent commands
│   └── knowledge/
│       ├── codebase-map.md     # AI's understanding
│       └── patterns.md         # Discovered patterns
```

```bash
# Commit code + AI context together
git add src/auth.ts .stackagent/context/
git commit -m "feat: Add auth - see AI context for decisions"

# New developer joins
git pull
stackagent load  # Inherits all AI knowledge!
```

### 2.4 AI Process Forking

**Problem**: How to query past AI states or get second opinions from different models?

**Solution**: Fork AI processes instead of complex context swapping.

```go
// Ask your past self (simple process fork)
pastMe := agent.ForkPastSelf("v1.0-release")
answer := agent.AskForkedAI(pastMe, "Why did we choose PostgreSQL?")
pastMe.Close()

// Get different model perspectives
opus := agent.ForkDifferentModel("claude-3-opus")
sonnet := agent.ForkDifferentModel("claude-3-sonnet")
gpt4 := agent.ForkDifferentModel("gpt-4")

// Compare their approaches
answers := []string{
    agent.AskForkedAI(opus, "How would you implement this?"),
    agent.AskForkedAI(sonnet, "How would you implement this?"),
    agent.AskForkedAI(gpt4, "How would you implement this?"),
}
```

### 2.5 Interactive Shell with Secrets Oracle

**Problem**: AI either can't handle password prompts or sees passwords in plain text.

**Solution**: AI detects prompts and injects secrets without ever seeing them.

```go
// AI runs sudo command
agent.RunInSession("upgrade", "sudo apt upgrade")
// Gets notified: "Session waiting for input: '[sudo] password for user:'"
agent.InjectSecret("upgrade", "sudo-password")
// Password injected directly to PTY, AI never sees it

// If password accidentally echoed, EDR detects immediately
// "CRITICAL: Secret 'sudo-password' appeared in clear text"
```

### 2.6 Cost-Aware Context Management

**Problem**: Users have no control over AI costs, sessions get expensive.

**Solution**: User-controlled context budgets with intelligent compression.

```go
// Set context budget
agent.SetContextBudget(50_000)  // ~$1.50 per query

// AI compresses intelligently when approaching limit
agent.ReduceToTarget(50_000)
// "Compressed old investigations and file contents.
//  Reduced from 150k to 50k tokens. Cost now ~$1.50/query"

// Presets for different scenarios
agent.UsePreset("quick-fix")     // 20k tokens, $0.60/query
agent.UsePreset("exploration")   // 150k tokens, $4.50/query
agent.UsePreset("cost-conscious") // 50k tokens, $1.50/query
```

## 3. User Interface Design

### 3.1 GUI Philosophy: Transparency and Control

StackAgent's GUI is designed around a core principle: **make AI actions transparent and give users full control**. The interface must clearly show what the AI is doing, why it's doing it, and allow users to understand and intervene at any point.

**Design Principles:**
- **Radical Transparency**: Every AI action is visible and explainable
- **User Control**: Users can inspect, modify, or stop AI actions at any time
- **Context Awareness**: Always show current state, branch, and session info
- **Efficiency**: Minimize cognitive load while maximizing information

### 3.2 Dual-Pane Architecture

The GUI uses a **resizable dual-pane layout** with a draggable vertical divider:

#### Left Pane: Conversation & Control
- **Chat Window**: Scrollable conversation history with AI
- **Text Entry**: Input field for user prompts and commands
- **Session Info**: Current AI cost, model selection, context status
- **Function Widgets**: Interactive elements for each AI function call

#### Right Pane: Action Details
- **Function Viewer**: Shows details of the most recent or selected function call
- **Real-time Updates**: Live output from running commands
- **Context Browser**: View current memory, workspace state, Git status
- **File Preview**: Display files being modified or created

### 3.3 Enhanced Features

**Chat Experience:**
- **Syntax Highlighting**: Code snippets and file contents properly formatted
- **Function Call Indicators**: Status badges (running, completed, failed, pending)
- **Searchable History**: Find previous conversations and commands
- **Export/Import**: Save sessions, share contexts with team

**Context Visualization:**
- **Branch Indicator**: Always show current Git branch and context
- **Memory Dashboard**: View protected memory and knowledge base
- **Session Timeline**: Track commands and AI decisions over time
- **Cost Tracking**: Real-time API usage and cost monitoring

**Interaction Enhancements:**
- **Keyboard Shortcuts**: Power user navigation and commands
- **Dark/Light Themes**: User preference and eye strain reduction
- **Responsive Design**: Works on different screen sizes
- **Real-time Streaming**: Live AI responses and command output

### 3.4 Function Call Integration

**Interactive Function Widgets:**
- **Clickable Status**: Click any function call to view details in right pane
- **Progress Indicators**: Show running commands with progress bars
- **Error Handling**: Clear error messages and retry options
- **Output Preview**: Inline preview of small outputs, full view in right pane

**Right Pane Function Details:**
- **Command Output**: Full terminal output with syntax highlighting
- **File Diffs**: Show code changes with before/after comparison
- **Context Changes**: Display updates to memory, workspace, or Git state
- **Performance Metrics**: Execution time, tokens used, cost incurred

### 3.5 Advanced UI Features

**Workspace Integration:**
- **File Tree**: Navigate project files and see AI modifications
- **Git Integration**: Visual branch switching, commit history, context restoration
- **Multi-tab Support**: Handle multiple conversations or projects
- **Split Views**: Compare different AI responses or function results

**Productivity Features:**
- **Command Palette**: Quick access to all functions and settings
- **Pinned Results**: Keep important function results visible
- **Automation Scripts**: Save and reuse common AI workflows
- **Collaboration Tools**: Share contexts and sessions with team members

### 3.6 Technical Implementation

**Framework Considerations:**
- **Web-based**: Cross-platform compatibility, easy deployment
- **Real-time**: WebSocket connections for live updates
- **Responsive**: CSS Grid/Flexbox for flexible layout
- **Accessible**: ARIA labels, keyboard navigation, screen reader support

**Component Architecture:**
```
GUI/
├── ChatPane/
│   ├── ConversationView
│   ├── InputField
│   ├── SessionInfo
│   └── FunctionWidgets
├── ActionPane/
│   ├── FunctionViewer
│   ├── ContextBrowser
│   ├── FilePreview
│   └── OutputDisplay
├── Common/
│   ├── Layout (resizable panes)
│   ├── Theming
│   └── KeyboardShortcuts
└── Services/
    ├── WebSocket (real-time updates)
    ├── StateManagement
    └── API Integration
```

**Technology Stack:**
- **Frontend**: React/Vue.js + TypeScript
- **Styling**: TailwindCSS for rapid development
- **Real-time**: WebSocket + Server-Sent Events
- **State**: Redux/Zustand for complex state management
- **Build**: Vite for fast development and builds

### 3.7 User Experience Flow

**Typical Session:**
1. **Startup**: GUI loads, shows context status, Git branch
2. **Chat**: User types request, AI responds with function calls
3. **Transparency**: Each function call shows as widget in chat
4. **Details**: User clicks function widget → details appear in right pane
5. **Control**: User can stop, modify, or retry any action
6. **Context**: All actions update persistent memory and workspace

**Power User Features:**
- **Keyboard Navigation**: Tab through function calls, use shortcuts
- **Batch Operations**: Select multiple function calls, apply actions
- **Session Management**: Save, load, and switch between contexts
- **Advanced Filtering**: Search across all sessions and contexts

This GUI design ensures that StackAgent is not just powerful, but also **transparent, controllable, and user-friendly** - making AI assistance accessible to both novice and expert developers.

## 4. Security Architecture

### 4.1 Privacy-First Design

**Key Innovation**: Hardware attestation protects user's trade secrets FROM StackAgent operators.

1. **Separation of Collection and Analysis**
   - EDR endpoints (on user machines) ONLY collect signals
   - All analysis happens in CVMs (cloud or on-prem), NEVER locally

2. **Privacy Through Encryption**
   - Signals encrypted on user's machine before transmission
   - Only attested CVMs can decrypt (via OpenADP key distribution)
   - StackAgent operators never have access to raw signals

3. **User Transparency**
   - Users can view full clear text of concerning signals
   - Complete audit trail of what was detected and why

### 3.2 Trust Flow with OpenADP

```
1. EDR endpoint encrypts signals using CVM public keys
2. Signal analysis CVMs request attestation from hardware
3. CVMs submit attestation reports to OpenADP network
4. Multiple OpenADP validators verify attestation independently
5. Only after verification do validators release key shares
6. CVM reconstructs decryption key and processes signals
7. If code is tampered, attestation fails = no keys = no access
```

### 3.3 Enterprise Deployment Options

**Option 1: StackAgent Cloud CVMs** (Default)
- Signals encrypted before leaving premises
- Analysis in StackAgent's hardware-attested CVMs
- Zero access for StackAgent operators

**Option 2: On-Premise CVMs** (Enterprise)
- Run identical CVM infrastructure on-premise
- Signals never leave enterprise network
- Same security guarantees as cloud

## 5. Implementation Roadmap

### Phase 1: Core MVP with Context Innovation (2-3 months)

**High-Value/High-Risk Features First:**

#### Context Management (Revolutionary - De-risk early)
- [ ] **AI-controlled context windows** - RunWithCapture pattern
- [ ] **Output handles** - Query command output like files
- [ ] **Git-based context storage** - Version control for AI memory
- [ ] **Protected persistent memory** - Context that survives sessions
- [ ] **Cost tracking & budgets** - User controls context size/cost
- [ ] **Cache-aware optimization** - 90% cost reduction

#### Core Capabilities (Must work reliably)
- [ ] **Shell multiplexing** - Multiple PTYs with interactive detection
- [ ] **Secrets oracle** - Password injection without AI seeing values
- [ ] **File operations** - Simple read/write/search (no fancy AST)
- [ ] **Process management** - Start/stop/monitor services
- [ ] **AI recommendation system** - AI tells us what to build next

#### Basic Security (Foundation)
- [ ] **Local EDR endpoint** - Signal collection only
- [ ] **Signal encryption** - Before leaving user machine
- [ ] **Basic secrets vault** - Local encrypted storage

#### Simple UI
- [ ] **Wails-based desktop app** - Single binary, 40-80MB
- [ ] **Terminal display** - Show active sessions
- [ ] **Cost dashboard** - Real-time token usage

**Explicitly NOT in MVP:**
- Complex code analysis
- Language servers
- Cloud storage (using git)
- Semantic search
- Advanced UI features

### Phase 2: Security Hardening (1-2 months)
- [ ] Deploy signal analysis CVMs (AMD SEV-SNP/Intel TDX)
- [ ] OpenADP attestation integration
- [ ] Enterprise on-prem CVM option
- [ ] Hardware-backed secrets storage
- [ ] Secure update mechanism

### Phase 3: Advanced Context Features (2 months)
- [ ] **AI process forking** - Query past versions or different models
- [ ] **Context branching** - Try experiments without losing work
- [ ] **Team knowledge sharing** - via git pull/push
- [ ] **Pattern learning** - Detect repeated sequences
- [ ] **Advanced compression** - Smarter context reduction

### Phase 4: Scale & Polish (2 months)
- [ ] Multi-LLM support with easy switching
- [ ] Advanced EDR analytics
- [ ] Enterprise SSO/SAML
- [ ] Performance optimizations
- [ ] Plugin system for extensions

## 6. Technical Implementation

### 6.1 Context-Aware Command Execution

```go
type CommandCapture interface {
    // Run ANY command, get a handle (not dumped output)
    RunWithCapture(cmd string, opts CaptureOptions) OutputHandle
    
    // Query the output like a file
    SearchOutput(handle OutputHandle, pattern string) []Match
    ReadLines(handle OutputHandle, start, end int) string
    GetTail(handle OutputHandle, lines int) string
    GetStats(handle OutputHandle) Stats
    
    // For long-running commands
    StreamUpdates(handle OutputHandle) <-chan Update
}

// Implementation that saves tokens/money
func RunWithCapture(cmd string, opts CaptureOptions) (*OutputHandle, error) {
    handle := &OutputHandle{
        ID:        uuid.New().String(),
        Command:   cmd,
        StartTime: time.Now(),
    }
    
    // Execute with streaming
    cmd := exec.Command("bash", "-c", cmd)
    stdout, _ := cmd.StdoutPipe()
    
    go func() {
        scanner := bufio.NewScanner(stdout)
        for scanner.Scan() {
            handle.buffer.Append(scanner.Text())
            
            // Callback for long-running commands
            if opts.CallbackInterval > 0 {
                handle.NotifyAI()
            }
        }
    }()
    
    // Store handle for queries
    outputHandles[handle.ID] = handle
    return handle, nil
}
```

### 5.2 Shell Session Management

```go
type ShellSession struct {
    name       string
    pty        *os.File
    cmd        *exec.Cmd
    state      ShellState
}

type ShellState struct {
    WaitingForInput bool
    InputPrompt     string  // "[sudo] password for user:"
    PromptType      string  // "password", "confirmation"
}

func (sm *ShellManager) DetectPrompt(output string) {
    prompts := []string{
        "[sudo] password",
        "Password:",
        "Enter passphrase",
        "(yes/no)",
    }
    
    for _, prompt := range prompts {
        if strings.Contains(output, prompt) {
            session.state.WaitingForInput = true
            session.state.InputPrompt = output
            sm.NotifyAI(session)
        }
    }
}
```

### 5.3 Git-Based Context Persistence

```go
type GitContextManager struct {
    repoPath string
}

func (gcm *GitContextManager) SaveContext(ctx Context) error {
    // Save to .stackagent/context/
    files := map[string]interface{}{
        "memory/current.md":         ctx.CurrentMemory,
        "workspace/active-files.json": ctx.ActiveFiles,
        "knowledge/patterns.md":     ctx.LearnedPatterns,
    }
    
    for path, content := range files {
        fullPath := filepath.Join(".stackagent/context", path)
        if err := writeFile(fullPath, content); err != nil {
            return err
        }
    }
    
    return nil
}

func (gcm *GitContextManager) OnGitCheckout(from, to string) {
    if gcm.ContextExistsFor(to) {
        gcm.RestoreContext(to)
        gcm.NotifyUser("Restored AI context from " + to)
    }
}
```

### 5.4 Cache-Aware Token Optimization

```go
type CacheOptimizer struct {
    cachedPrefix []Token  // Rarely changes, cheap
    recentWork   []Token  // Changes often, expensive
}

func (co *CacheOptimizer) BuildPrompt(query string) Prompt {
    // Keep cached content at start for provider caching
    return Prompt{
        CachedSection: co.cachedPrefix,    // $0.003/1k tokens
        FreshSection:  co.CompressRecent(), // $0.030/1k tokens
        Query:         query,
    }
}

func (co *CacheOptimizer) CompressRecent() []Token {
    // Aggressively compress expensive recent tokens
    // 50k tokens → 5k tokens = 90% cost reduction
    return Summarize(co.recentWork, ratio=0.9)
}
```

## 7. Technology Stack

### 7.1 Recommended Architecture: Wails (Go + Web Frontend)

**Decision**: Lightweight desktop app with Go backend and modern web UI.

```yaml
application:
  framework: Wails v2
  language: Go 1.21+
  
frontend:
  ui_library: React or Svelte
  terminal: xterm.js
  build_tool: Vite
  
core_services:
  shell_management: github.com/creack/pty
  browser_automation: chromedp
  desktop_control: robotgo
  secrets: age encryption + keyring
  
deployment:
  format: Single native binary per platform
  size: ~40-80MB
  auto_update: Built-in Wails updater
```

### 6.2 Why This Stack

1. **Single Binary** - Easy distribution, no complex installers
2. **Native Performance** - Direct system access for shells/files
3. **Modern UI** - Web tech for UI without Electron bloat
4. **Go Throughout** - One language for all backend logic
5. **Proven Tools** - PTY, grep, sed over complex abstractions

## 8. Deployment Architecture

### 8.1 StackAgent Cloud (Default)

```
Users Worldwide → Encrypted Signals → StackAgent CVMs
                                      ├─ Hardware Attested
                                      ├─ OpenADP Verified
                                      └─ Zero Operator Access
```

### 8.2 Enterprise On-Premise

```
Enterprise Users → Encrypted Signals → On-Prem CVMs
(Stays on-prem)                        ├─ Same Security
                                      ├─ Full Control
                                      └─ SIEM Integration
```
