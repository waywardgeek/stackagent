# StackAgent MVP Implementation Plan

## Overview
Build a working AI coding agent with revolutionary context management in 8-10 weeks. Each milestone produces working software.

## Development Principles
1. **Working software every week** - No long stretches without runnable code
2. **Test with real tasks** - Use StackAgent to build StackAgent
3. **Simple first, optimize later** - Get it working, then make it fast
4. **Direct implementation** - No abstractions until patterns emerge

## Milestone 1: Basic Shell & Context Foundation (Week 1-2)

### Goal: AI can run commands and manage output efficiently

#### 1.1 Project Setup
```bash
# Initialize Wails project
wails init -n stackagent -t vanilla
cd stackagent

# Add core dependencies
go get github.com/creack/pty
go get github.com/google/uuid
go get github.com/sashabaranov/go-openai  # Or anthropic client
```

#### 1.2 Core Command Execution with Handles
```go
// pkg/shell/capture.go
type OutputHandle struct {
    ID        string
    Command   string
    Buffer    []string  // Start simple with string slice
    Complete  bool
    ExitCode  int
}

type ShellManager struct {
    handles map[string]*OutputHandle
}

func (sm *ShellManager) RunWithCapture(cmd string) (*OutputHandle, error) {
    handle := &OutputHandle{
        ID:      uuid.New().String(),
        Command: cmd,
        Buffer:  []string{},
    }
    
    // Execute command
    cmd := exec.Command("bash", "-c", cmd)
    stdout, _ := cmd.StdoutPipe()
    cmd.Start()
    
    // Capture output
    go func() {
        scanner := bufio.NewScanner(stdout)
        for scanner.Scan() {
            handle.Buffer = append(handle.Buffer, scanner.Text())
        }
        cmd.Wait()
        handle.Complete = true
        handle.ExitCode = cmd.ProcessState.ExitCode()
    }()
    
    sm.handles[handle.ID] = handle
    return handle, nil
}

// Query functions
func (sm *ShellManager) SearchOutput(handleID, pattern string) []string
func (sm *ShellManager) ReadLines(handleID string, start, end int) string
func (sm *ShellManager) GetTail(handleID string, lines int) string
```

#### 1.3 Basic Wails UI
```typescript
// frontend/src/main.ts
import { RunWithCapture, SearchOutput, ReadLines } from '../wailsjs/go/main/App'

// Simple terminal display
function Terminal() {
    const [handles, setHandles] = useState([])
    
    async function runCommand(cmd: string) {
        const handle = await RunWithCapture(cmd)
        setHandles([...handles, handle])
    }
    
    return (
        <div className="terminal">
            <input onKeyPress={e => e.key === 'Enter' && runCommand(e.target.value)} />
            <div className="output">
                {handles.map(h => <OutputDisplay handle={h} />)}
            </div>
        </div>
    )
}
```

#### 1.4 Test Milestone
```bash
# Can run commands and query output
handle := agent.RunWithCapture("ls -la")
files := agent.SearchOutput(handle.ID, ".go")
tail := agent.GetTail(handle.ID, 10)
```

## Milestone 2: Context Persistence with Git (Week 3-4)

### Goal: AI memory persists across sessions in `.stackagent/`

#### 2.1 Context Structure
```go
// pkg/context/manager.go
type ContextManager struct {
    basePath string  // .stackagent/context/
}

type Context struct {
    Memory    map[string]string      // Protected memory
    Workspace WorkspaceState         // Active files, commands
    Knowledge map[string]string      // Learned patterns
}

func (cm *ContextManager) SaveContext(ctx Context) error {
    // Save to .stackagent/context/
    os.MkdirAll(filepath.Join(cm.basePath, "memory"), 0755)
    
    // Write memory/current.json
    memoryData, _ := json.Marshal(ctx.Memory)
    os.WriteFile(filepath.Join(cm.basePath, "memory/current.json"), memoryData, 0644)
    
    // Write workspace/state.json
    workspaceData, _ := json.Marshal(ctx.Workspace)
    os.WriteFile(filepath.Join(cm.basePath, "workspace/state.json"), workspaceData, 0644)
    
    return nil
}

func (cm *ContextManager) LoadContext() (*Context, error) {
    // Load from .stackagent/context/
}

// Protected memory that never falls off
func (cm *ContextManager) SetProtected(key string, value string) error
func (cm *ContextManager) GetProtected(key string) (string, error)
```

#### 2.2 Git Integration
```go
// pkg/context/git.go
func (cm *ContextManager) OnGitCheckout(from, to string) {
    // Check if .stackagent/context exists at target commit
    if cm.ContextExistsAt(to) {
        cm.LoadContextFrom(to)
        fmt.Printf("Restored AI context from %s\n", to)
    }
}

// Git hooks
// .git/hooks/post-checkout
#!/bin/bash
stackagent context restore --auto
```

#### 2.3 Test Milestone
```bash
# Set persistent memory
agent.SetProtected("current-task", "implement auth")
agent.SetProtected("design-decisions", "using JWT for stateless auth")

# Survives restarts
stackagent restart
task := agent.GetProtected("current-task")  # "implement auth"

# Survives git branches
git checkout -b feature
# Context switches with branch
```

## Milestone 3: Multi-Session Shell with Prompt Detection (Week 5-6)

### Goal: Multiple PTYs with password prompt handling

#### 3.1 PTY Management
```go
// pkg/shell/pty.go
type PTYSession struct {
    Name     string
    pty      *os.File
    cmd      *exec.Cmd
    output   chan string
    state    SessionState
}

type SessionState struct {
    WorkingDir      string
    WaitingForInput bool
    InputPrompt     string
    PromptType      string  // "password", "confirmation"
}

func (sm *ShellManager) NewSession(name string) (*PTYSession, error) {
    cmd := exec.Command("bash")
    ptmx, _ := pty.Start(cmd)
    
    session := &PTYSession{
        Name:   name,
        pty:    ptmx,
        cmd:    cmd,
        output: make(chan string, 1000),
    }
    
    // Monitor for prompts
    go sm.monitorSession(session)
    
    return session, nil
}

func (sm *ShellManager) monitorSession(session *PTYSession) {
    reader := bufio.NewReader(session.pty)
    for {
        line, _ := reader.ReadString('\n')
        session.output <- line
        
        // Detect prompts
        if sm.detectPrompt(line) {
            session.state.WaitingForInput = true
            session.state.InputPrompt = line
            sm.notifyAI(session)
        }
    }
}

func (sm *ShellManager) detectPrompt(line string) bool {
    prompts := []string{
        "[sudo] password",
        "Password:",
        "Enter passphrase",
        "(yes/no)",
    }
    for _, p := range prompts {
        if strings.Contains(line, p) {
            return true
        }
    }
    return false
}
```

#### 3.2 Secrets Oracle
```go
// pkg/secrets/oracle.go
type SecretsOracle struct {
    vault map[string]string  // In-memory for MVP
}

func (so *SecretsOracle) RegisterSecret(name, value string) error {
    so.vault[name] = value
    return nil
}

func (so *SecretsOracle) InjectSecret(sessionName, secretRef string) error {
    session := getSession(sessionName)
    if !session.state.WaitingForInput {
        return fmt.Errorf("session not waiting for input")
    }
    
    secret := so.vault[secretRef]
    session.pty.Write([]byte(secret + "\n"))
    
    // Log for EDR
    logSecretInjection(sessionName, secretRef)
    
    return nil
}
```

#### 3.3 Test Milestone
```bash
# Multiple sessions
agent.NewSession("backend")
agent.NewSession("frontend")
agent.RunInSession("backend", "npm run dev")
agent.RunInSession("frontend", "npm start")

# Password handling
agent.RegisterSecret("sudo-pass", "actual-password")
agent.RunInSession("admin", "sudo apt update")
# AI notified: "Session 'admin' waiting for password"
agent.InjectSecret("admin", "sudo-pass")
# Command proceeds without AI seeing password
```

## Milestone 4: Cost-Aware Context & Caching (Week 7-8)

### Goal: 90% cost reduction through smart context management

#### 4.1 Token Tracking
```go
// pkg/context/tokens.go
type TokenTracker struct {
    sessionTokens int
    sessionCost   float64
    cacheHits     int
}

func (tt *TokenTracker) CountTokens(text string) int {
    // Simple estimation: ~4 chars per token
    return len(text) / 4
}

func (tt *TokenTracker) EstimateCost(tokens int, cached bool) float64 {
    if cached {
        return float64(tokens) * 0.00003  // $0.03 per 1K cached
    }
    return float64(tokens) * 0.0003     // $0.30 per 1K fresh
}
```

#### 4.2 Smart Context Building
```go
// pkg/context/optimizer.go
type ContextOptimizer struct {
    cachedPrefix  string  // System prompt, function defs
    recentWork    string  // Current investigation
    outputHandles map[string]string  // Summaries only
}

func (co *ContextOptimizer) BuildPrompt(query string) string {
    // Cached prefix first (for provider caching)
    prompt := co.cachedPrefix  // 10K tokens, cached
    
    // Add compressed recent work
    compressed := co.compressRecent()  // 50K → 5K tokens
    prompt += "\n\n=== Recent Work ===\n" + compressed
    
    // Add query
    prompt += "\n\nUser Query: " + query
    
    return prompt
}

func (co *ContextOptimizer) compressRecent() string {
    // Summarize instead of including everything
    summary := "Recent commands:\n"
    for id, handle := range co.outputHandles {
        summary += fmt.Sprintf("- %s (ID: %s): %d lines output\n", 
            handle.Command, id, len(handle.Buffer))
    }
    return summary
}
```

#### 4.3 User Controls
```go
// pkg/context/budget.go
type ContextBudget struct {
    MaxTokens     int
    CurrentTokens int
    TargetCost    float64
}

func (cb *ContextBudget) SetBudget(tokens int) {
    cb.MaxTokens = tokens
    // Adjust compression accordingly
}

// Presets
var PRESETS = map[string]ContextBudget{
    "quick-fix":      {MaxTokens: 20_000},   // $0.60/query
    "balanced":       {MaxTokens: 50_000},   // $1.50/query
    "exploration":    {MaxTokens: 150_000},  // $4.50/query
}
```

#### 4.4 Test Milestone
```bash
# Run expensive operation
output := agent.RunWithCapture("find . -type f -name '*.go' | xargs wc -l")

# Check token usage
stats := agent.GetContextStats()
# "Current context: 8,000 tokens ($0.24/query with caching)"

# Reduce if needed
agent.SetContextBudget(30_000)
# "Compressed old investigations. Now at 30,000 tokens ($0.90/query)"
```

## Milestone 5: AI Recommendations & EDR Foundation (Week 9-10)

### Goal: AI helps improve itself + basic security

#### 5.1 Pattern Tracking
```go
// pkg/recommend/tracker.go
type PatternTracker struct {
    patterns map[string]int
}

func (pt *PatternTracker) Track(pattern string) {
    pt.patterns[pattern]++
    if pt.patterns[pattern] >= 3 {
        pt.generateRecommendation(pattern)
    }
}

func (pt *PatternTracker) generateRecommendation(pattern string) {
    rec := Recommendation{
        Title:     "Frequently used pattern",
        Pattern:   pattern,
        Frequency: pt.patterns[pattern],
        Suggestion: "Create helper function",
    }
    sendToUI(rec)
}
```

#### 5.2 Basic EDR
```go
// pkg/edr/collector.go
type EDRCollector struct {
    signals chan Signal
}

type Signal struct {
    Timestamp time.Time
    Action    string
    Details   map[string]interface{}
}

func (edr *EDRCollector) LogAction(action string, details map[string]interface{}) {
    signal := Signal{
        Timestamp: time.Now(),
        Action:    action,
        Details:   details,
    }
    
    // Just collect for MVP, analysis comes later
    edr.signals <- signal
}

// Log all actions
func (sm *ShellManager) RunCommand(cmd string) {
    edr.LogAction("shell_command", map[string]interface{}{
        "command": cmd,
        "session": session.Name,
    })
}
```

#### 5.3 Integration Test
```bash
# Full workflow test
agent.SetProtected("task", "implement user auth")
handle := agent.RunWithCapture("grep -r 'password' --include='*.go'")
results := agent.SearchOutput(handle.ID, "hash")

# AI notices pattern
# "You've searched for passwords 5 times. Recommend: agent.FindSecurityIssues()"

# Check EDR collected signals
signals := agent.GetEDRSignals()
# Shows all commands, file reads, secret injections
```

## Development Schedule

**Week 1-2**: Basic shell with output handles ✓ Working
**Week 3-4**: Git-based context persistence ✓ Working  
**Week 5-6**: Multi-PTY + secrets oracle ✓ Working
**Week 7-8**: Cost optimization + caching ✓ Working
**Week 9-10**: Recommendations + EDR ✓ MVP Complete

## Success Criteria

1. **Can persist context across sessions** - Check with restart
2. **Can handle multiple shells** - Run frontend + backend
3. **Never sees passwords** - Test with sudo commands
4. **90% cost reduction** - Measure tokens before/after
5. **AI can recommend improvements** - Track patterns

## Next Steps After MVP

Once MVP is working and we're using it daily:
1. Add whatever features the AI recommends most
2. Implement CVM security hardening
3. Add AI process forking for past queries
4. Polish UI based on actual usage

## Technical Decisions

- **No fancy abstractions** - Direct implementations only
- **No premature optimization** - Make it work first
- **No complex dependencies** - Stdlib + essential packages
- **No perfect code** - Ship working software weekly

Ready to start building! The first week's goal is just getting `RunWithCapture()` working reliably. 