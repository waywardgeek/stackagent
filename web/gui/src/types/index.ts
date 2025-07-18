// Core types for StackAgent GUI

export interface User {
  id: string;
  name: string;
  email?: string;
}

export interface Session {
  id: string;
  userId: string;
  createdAt: Date;
  lastActivity: Date;
  totalCost: number;
  messageCount: number;
  gitBranch?: string;
  gitCommit?: string;
}

export interface Message {
  id: string;
  sessionId: string;
  type: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: Date;
  functionCalls?: FunctionCall[];
  tokens?: number;
  cost?: number;
  // New fields for interactive widgets
  operationSummary?: OperationSummary;
}

// New type for operation summaries in chat
export interface OperationSummary {
  shellCommands?: ShellOperation[];
  fileOperations?: FileOperation[];
  hasOperations: boolean;
}

export interface ShellOperation {
  id: string;
  command: string;
  output: string;
  exitCode: number;
  duration: number;
  workingDir: string;
  timestamp: Date;
  // Enhanced for real-time streaming
  isLive?: boolean;
  streamingOutput?: string;
  progress?: number;
}

export interface FileOperation {
  id: string;
  type: 'read' | 'write' | 'edit' | 'search' | 'list';
  filePath: string;
  content?: string;
  changes?: string; // For edits, this would be the diff
  searchResults?: string[];
  timestamp: Date;
  size?: number;
  // Enhanced for real-time streaming
  isLive?: boolean;
  streamingContent?: string;
  progress?: number;
}

export interface FunctionCall {
  id: string;
  messageId: string;
  name: string;
  arguments: Record<string, any>;
  status: 'pending' | 'running' | 'completed' | 'failed';
  result?: any;
  error?: string;
  startTime: Date;
  endTime?: Date;
  duration?: number;
  selected?: boolean;
  // Enhanced for real-time streaming
  isLive?: boolean;
  streamingOutput?: string;
  progress?: number;
}

export interface CommandExecution {
  id: string;
  handleId: number;
  command: string;
  status: 'running' | 'completed' | 'failed';
  exitCode?: number;
  output?: string;
  error?: string;
  startTime: Date;
  endTime?: Date;
  duration?: number;
  lineCount?: number;
}

export interface ContextState {
  sessionId: string;
  memoryEntries: number;
  knowledgeEntries: number;
  commandHistory: number;
  activeHandles: number;
  activeFiles: number;
  lastActivity: Date;
  createdAt: Date;
  gitBranch?: string;
  gitCommit?: string;
  totalCost?: number;
  requestCount?: number;
  cacheStats?: {
    cacheHits: number;
    cacheMisses: number;
    totalSavings: number;
    cacheEfficiency: number;
  };
}

export interface ProtectedMemory {
  [key: string]: string;
}

export interface WorkspaceState {
  activeFiles: string[];
  commandHistory: CommandRecord[];
  workingDir: string;
  activeHandles: number[];
  currentTask?: string;
  projectContext?: string;
  lastActivity: Date;
}

export interface CommandRecord {
  command: string;
  handleId: number;
  timestamp: Date;
  exitCode: number;
  duration: string;
  summary?: string;
}

export interface KnowledgeBase {
  [key: string]: string;
}

export interface AIModel {
  id: string;
  name: string;
  provider: string;
  inputCost: number;
  outputCost: number;
  contextWindow: number;
  description?: string;
}

export interface WebSocketMessage {
  type: 'message' | 'function_call' | 'function_result' | 'context_update' | 'error';
  data: any;
  timestamp: Date;
}

export interface AppState {
  // UI State
  leftPaneWidth: number;
  rightPaneWidth: number;
  selectedFunctionCall?: string;
  theme: 'light' | 'dark';
  sidebarCollapsed: boolean;
  
  // Session State
  currentSession?: Session;
  messages: Message[];
  functionCalls: FunctionCall[];
  
  // Context State
  contextState?: ContextState;
  protectedMemory: ProtectedMemory;
  workspaceState?: WorkspaceState;
  knowledgeBase: KnowledgeBase;
  
  // AI State
  selectedModel: string;
  availableModels: AIModel[];
  isStreaming: boolean;
  
  // WebSocket State
  connected: boolean;
  reconnectAttempts: number;
  
  // Commands State
  commandExecutions: CommandExecution[];
  activeCommands: number;
  
  // Terminal State
  isTerminalVisible: boolean;
  currentTerminalOperation?: ShellOperation;
  allShellOperations: ShellOperation[];
}

export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  timestamp: Date;
}

export interface ChatInputState {
  message: string;
  isComposing: boolean;
  mentionedFiles: string[];
  attachments: File[];
}

export interface FilePreview {
  path: string;
  content: string;
  language: string;
  modified: boolean;
  changes?: {
    added: number;
    removed: number;
    modified: number;
  };
}

export interface GitBranch {
  name: string;
  current: boolean;
  hasContext: boolean;
  contextStats?: {
    memoryEntries: number;
    knowledgeEntries: number;
    commandHistory: number;
  };
}

export interface SessionCost {
  totalCost: number;
  inputTokens: number;
  outputTokens: number;
  inputCost: number;
  outputCost: number;
  requestCount: number;
}

export interface Notification {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  message: string;
  timestamp: Date;
  duration?: number;
  actions?: Array<{
    label: string;
    action: () => void;
  }>;
}

// Event types for WebSocket communication
export type WebSocketEventType =
  | 'session_started'
  | 'session_ended'
  | 'message_received'
  | 'function_call_started'
  | 'function_call_completed'
  | 'function_call_failed'
  | 'function_call_streaming'
  | 'shell_command_started'
  | 'shell_command_streaming'
  | 'shell_command_completed'
  | 'file_operation_started'
  | 'file_operation_streaming'
  | 'file_operation_completed'
  | 'context_updated'
  | 'command_started'
  | 'command_completed'
  | 'command_output'
  | 'git_branch_changed'
  | 'cost_updated'
  | 'error_occurred'
  | 'chat_message'
  | 'ai_response'
  | 'ai_streaming'
  | 'ai_error'
  | 'user_message'
  | 'get_context'
  | 'debug_message'
  | 'configure_streaming'
  | 'ping'
  | 'pong';

export interface WebSocketEvent {
  type: WebSocketEventType;
  data: any;
  timestamp: Date;
  sessionId: string;
}

// Real-time streaming data structures
export interface StreamingData {
  operationId: string;
  type: 'function' | 'shell' | 'file';
  content: string;
  progress?: number;
  timestamp: Date;
}

export interface LiveOperationStatus {
  id: string;
  type: 'function' | 'shell' | 'file';
  status: 'starting' | 'running' | 'streaming' | 'completed' | 'failed';
  progress?: number;
  output?: string;
  startTime: Date;
  estimatedCompletion?: Date;
}

// Enhanced chat message types for real-time
export interface LiveMessage extends Message {
  isLive: boolean;
  liveOperations?: string[];
  streamingProgress?: number;
  contextSnapshot?: Partial<ContextState>;
}

// UI View Types
export type ActionView = 'function-call' | 'command-output' | 'context' | 'file-preview' | 'debug-io' | 'live-operations';

// Debug Message Types
export interface DebugMessage {
  id: string;
  timestamp: Date;
  direction: 'sent' | 'received';
  type: 'websocket' | 'api' | 'error' | 'debug_message';
  event?: string;
  data: any;
  rawJson: string;
}

// Real-time performance metrics
export interface PerformanceMetrics {
  totalOperations: number;
  averageResponseTime: number;
  operationsPerMinute: number;
  activeOperations: number;
  streamingOperations: number;
  cacheHitRate: number;
  contextVersion: number;
  lastUpdate: Date;
} 