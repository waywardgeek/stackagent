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
  | 'context_updated'
  | 'command_started'
  | 'command_completed'
  | 'command_output'
  | 'git_branch_changed'
  | 'cost_updated'
  | 'error_occurred'
  | 'chat_message'
  | 'ai_response'
  | 'ai_error'
  | 'user_message'
  | 'get_context'
  | 'debug_message'
  | 'ping'
  | 'pong';

export interface WebSocketEvent {
  type: WebSocketEventType;
  data: any;
  timestamp: Date;
  sessionId: string;
}

// UI View Types
export type ActionView = 'function-call' | 'command-output' | 'context' | 'file-preview' | 'debug-io';

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