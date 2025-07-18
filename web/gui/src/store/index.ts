import { create } from 'zustand';
import { subscribeWithSelector } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import type { 
  AppState, 
  Message, 
  FunctionCall, 
  CommandExecution, 
  ContextState,
  ProtectedMemory,
  WorkspaceState,
  KnowledgeBase,
  AIModel,
  Session,
  Notification,
  DebugMessage,
  ShellOperation,
  FileOperation
} from '@/types';

// Enhanced real-time context state
interface RealTimeContext {
  // Live operation tracking
  activeFunctionCalls: Map<string, FunctionCall>;
  activeShellCommands: Map<string, ShellOperation>;
  activeFileOperations: Map<string, FileOperation>;
  
  // Streaming state
  isStreaming: boolean;
  streamingOperationId?: string;
  streamingType?: 'function' | 'shell' | 'file';
  
  // Context derivation
  contextVersion: number;
  lastContextUpdate: Date;
  pendingContextChanges: string[];
  
  // Performance metrics
  operationMetrics: {
    totalOperations: number;
    averageResponseTime: number;
    operationsPerMinute: number;
    lastMinuteOperations: number[];
  };
}

// Enhanced message with real-time tracking
interface EnhancedMessage extends Message {
  // Real-time states
  isLive?: boolean;
  liveOperations?: string[];
  contextSnapshot?: Partial<ContextState>;
  
  // Streaming states
  isStreaming?: boolean;
  streamProgress?: number;
  estimatedCompletion?: Date;
}

interface AppStore extends AppState {
  // Enhanced Real-Time Context
  realTimeContext: RealTimeContext;
  
  // Enhanced Messages
  messages: EnhancedMessage[];
  
  // Debug Messages
  debugMessages: DebugMessage[];
  
  // Terminal Pane State
  isTerminalVisible: boolean;
  currentTerminalOperation?: ShellOperation;
  allShellOperations: ShellOperation[];
  
  // UI Actions
  setLeftPaneWidth: (width: number) => void;
  setRightPaneWidth: (width: number) => void;
  selectFunctionCall: (id?: string) => void;
  toggleTheme: () => void;
  toggleSidebar: () => void;
  
  // Real-Time Context Actions
  startLiveOperation: (id: string, type: 'function' | 'shell' | 'file', operation: FunctionCall | ShellOperation | FileOperation) => void;
  updateLiveOperation: (id: string, updates: Partial<FunctionCall | ShellOperation | FileOperation>) => void;
  completeLiveOperation: (id: string) => void;
  setStreamingState: (streaming: boolean, operationId?: string, type?: 'function' | 'shell' | 'file') => void;
  updateContextVersion: () => void;
  trackOperationMetrics: (operationDuration: number) => void;
  
  // Enhanced Message Actions
  addLiveMessage: (message: Omit<EnhancedMessage, 'id'>) => void;
  updateMessageStreaming: (id: string, streaming: boolean, progress?: number) => void;
  addOperationToMessage: (messageId: string, operationId: string) => void;
  
  // Context Derivation
  deriveContextFromState: () => Partial<ContextState>;
  getActiveOperationsCount: () => number;
  getLiveOperationsByType: (type: 'function' | 'shell' | 'file') => Array<FunctionCall | ShellOperation | FileOperation>;
  
  // Terminal Actions
  showTerminal: (operation?: ShellOperation) => void;
  hideTerminal: () => void;
  addShellOperation: (operation: ShellOperation) => void;
  
  // Session Actions
  startSession: (session: Session) => void;
  endSession: () => void;
  
  // Message Actions
  addMessage: (message: Message) => void;
  updateMessage: (id: string, updates: Partial<Message>) => void;
  clearMessages: () => void;
  
  // Function Call Actions
  addFunctionCall: (functionCall: FunctionCall) => void;
  updateFunctionCall: (id: string, updates: Partial<FunctionCall>) => void;
  removeFunctionCall: (id: string) => void;
  clearFunctionCalls: () => void;
  
  // Command Actions
  addCommandExecution: (execution: CommandExecution) => void;
  updateCommandExecution: (id: string, updates: Partial<CommandExecution>) => void;
  removeCommandExecution: (id: string) => void;
  clearCommandExecutions: () => void;
  
  // Context Actions
  setContextState: (contextState: ContextState) => void;
  updateProtectedMemory: (key: string, value: string) => void;
  removeProtectedMemory: (key: string) => void;
  setProtectedMemory: (memory: ProtectedMemory) => void;
  setWorkspaceState: (workspace: WorkspaceState) => void;
  updateKnowledge: (key: string, value: string) => void;
  removeKnowledge: (key: string) => void;
  setKnowledgeBase: (knowledge: KnowledgeBase) => void;
  
  // AI Actions
  setSelectedModel: (modelId: string) => void;
  setAvailableModels: (models: AIModel[]) => void;
  setIsStreaming: (streaming: boolean) => void;
  
  // WebSocket Actions
  setConnected: (connected: boolean) => void;
  incrementReconnectAttempts: () => void;
  resetReconnectAttempts: () => void;
  
  // Debug Actions
  addDebugMessage: (message: Omit<DebugMessage, 'id'>) => void;
  clearDebugMessages: () => void;
  
  // Notification Actions
  notifications: Notification[];
  addNotification: (notification: Omit<Notification, 'id'>) => void;
  removeNotification: (id: string) => void;
  clearNotifications: () => void;
}

export const useAppStore = create<AppStore>()(
  subscribeWithSelector(
    immer((set, get) => ({
      // Initial UI State
      leftPaneWidth: 50,
      rightPaneWidth: 50,
      selectedFunctionCall: undefined,
      theme: 'dark',
      sidebarCollapsed: false,
      
      // Initial Real-Time Context
      realTimeContext: {
        activeFunctionCalls: new Map(),
        activeShellCommands: new Map(),
        activeFileOperations: new Map(),
        isStreaming: false,
        contextVersion: 0,
        lastContextUpdate: new Date(),
        pendingContextChanges: [],
        operationMetrics: {
          totalOperations: 0,
          averageResponseTime: 0,
          operationsPerMinute: 0,
          lastMinuteOperations: [],
        },
      },
      
      // Initial Session State
      currentSession: undefined,
      messages: [],
      functionCalls: [],
      
      // Initial Context State
      contextState: undefined,
      protectedMemory: {},
      workspaceState: undefined,
      knowledgeBase: {},
      
      // Initial AI State
      selectedModel: 'claude-sonnet-4-20250514',
      availableModels: [],
      isStreaming: false,
      
      // Initial WebSocket State
      connected: false,
      reconnectAttempts: 0,
      
      // Initial Commands State
      commandExecutions: [],
      activeCommands: 0,
      
      // Initial Notifications
      notifications: [],
      
      // Debug Messages
      debugMessages: [],
      
      // Terminal Pane State
      isTerminalVisible: false,
      currentTerminalOperation: undefined,
      allShellOperations: [],
      
      // Real-Time Context Actions
      startLiveOperation: (id, type, operation) => set((state) => {
        const rtContext = state.realTimeContext;
        
        switch (type) {
          case 'function':
            rtContext.activeFunctionCalls.set(id, operation as FunctionCall);
            break;
          case 'shell':
            rtContext.activeShellCommands.set(id, operation as ShellOperation);
            break;
          case 'file':
            rtContext.activeFileOperations.set(id, operation as FileOperation);
            break;
        }
        
        rtContext.contextVersion++;
        rtContext.lastContextUpdate = new Date();
        rtContext.pendingContextChanges.push(`${type}_started:${id}`);
      }),
      
      updateLiveOperation: (id, updates) => set((state) => {
        const rtContext = state.realTimeContext;
        
        // Update in all possible maps
        if (rtContext.activeFunctionCalls.has(id)) {
          const existing = rtContext.activeFunctionCalls.get(id)!;
          rtContext.activeFunctionCalls.set(id, { ...existing, ...updates });
        }
        if (rtContext.activeShellCommands.has(id)) {
          const existing = rtContext.activeShellCommands.get(id)!;
          rtContext.activeShellCommands.set(id, { ...existing, ...updates });
        }
        if (rtContext.activeFileOperations.has(id)) {
          const existing = rtContext.activeFileOperations.get(id)!;
          rtContext.activeFileOperations.set(id, { ...existing, ...updates });
        }
        
        rtContext.contextVersion++;
        rtContext.lastContextUpdate = new Date();
        rtContext.pendingContextChanges.push(`operation_updated:${id}`);
      }),
      
      completeLiveOperation: (id) => set((state) => {
        const rtContext = state.realTimeContext;
        
        // Get operation details before removing
        const functionOp = rtContext.activeFunctionCalls.get(id);
        const shellOp = rtContext.activeShellCommands.get(id);
        const fileOp = rtContext.activeFileOperations.get(id);
        
        const operation = functionOp || shellOp || fileOp;
        
        if (operation) {
          // Track metrics - only FunctionCall and ShellOperation have duration
          let duration = 0;
          if ('duration' in operation && operation.duration) {
            duration = operation.duration;
          }
          
          state.realTimeContext.operationMetrics.totalOperations++;
          
          // Update average response time
          const currentAvg = state.realTimeContext.operationMetrics.averageResponseTime;
          const totalOps = state.realTimeContext.operationMetrics.totalOperations;
          state.realTimeContext.operationMetrics.averageResponseTime = 
            (currentAvg * (totalOps - 1) + duration) / totalOps;
        }
        
        // Remove from active operations
        rtContext.activeFunctionCalls.delete(id);
        rtContext.activeShellCommands.delete(id);
        rtContext.activeFileOperations.delete(id);
        
        rtContext.contextVersion++;
        rtContext.lastContextUpdate = new Date();
        rtContext.pendingContextChanges.push(`operation_completed:${id}`);
      }),
      
      setStreamingState: (streaming, operationId, type) => set((state) => {
        state.realTimeContext.isStreaming = streaming;
        state.realTimeContext.streamingOperationId = operationId;
        state.realTimeContext.streamingType = type;
        
        if (streaming) {
          state.realTimeContext.pendingContextChanges.push(`streaming_started:${operationId || 'unknown'}`);
        } else {
          state.realTimeContext.pendingContextChanges.push(`streaming_ended:${operationId || 'unknown'}`);
        }
      }),
      
      updateContextVersion: () => set((state) => {
        state.realTimeContext.contextVersion++;
        state.realTimeContext.lastContextUpdate = new Date();
      }),
      
      trackOperationMetrics: (operationDuration) => set((state) => {
        const metrics = state.realTimeContext.operationMetrics;
        const now = Date.now();
        
        // Add to last minute operations
        metrics.lastMinuteOperations.push(now);
        
        // Remove operations older than 1 minute
        const oneMinuteAgo = now - 60000;
        metrics.lastMinuteOperations = metrics.lastMinuteOperations.filter(time => time > oneMinuteAgo);
        
        // Update operations per minute
        metrics.operationsPerMinute = metrics.lastMinuteOperations.length;
        
        // Use the provided duration for average calculation
        if (operationDuration > 0) {
          const currentAvg = metrics.averageResponseTime;
          const totalOps = metrics.totalOperations;
          metrics.averageResponseTime = ((currentAvg * (totalOps - 1)) + operationDuration) / totalOps;
        }
      }),
      
      // Enhanced Message Actions
      addLiveMessage: (message) => set((state) => {
        const id = Date.now().toString();
        const enhancedMessage: EnhancedMessage = {
          ...message,
          id,
          contextSnapshot: get().deriveContextFromState(),
        };
        state.messages.push(enhancedMessage);
      }),
      
      updateMessageStreaming: (id, streaming, progress) => set((state) => {
        const messageIndex = state.messages.findIndex(m => m.id === id);
        if (messageIndex !== -1) {
          state.messages[messageIndex].isStreaming = streaming;
          if (progress !== undefined) {
            state.messages[messageIndex].streamProgress = progress;
          }
        }
      }),
      
      addOperationToMessage: (messageId, operationId) => set((state) => {
        const messageIndex = state.messages.findIndex(m => m.id === messageId);
        if (messageIndex !== -1) {
          const message = state.messages[messageIndex];
          if (!message.liveOperations) {
            message.liveOperations = [];
          }
          if (!message.liveOperations.includes(operationId)) {
            message.liveOperations.push(operationId);
          }
          message.isLive = true;
        }
      }),
      
      // Context Derivation
      deriveContextFromState: () => {
        const state = get();
        const rtContext = state.realTimeContext;
        
        return {
          sessionId: state.currentSession?.id || 'current',
          memoryEntries: Object.keys(state.protectedMemory).length,
          knowledgeEntries: Object.keys(state.knowledgeBase).length,
          commandHistory: state.commandExecutions.length,
          activeHandles: rtContext.activeFunctionCalls.size + rtContext.activeShellCommands.size + rtContext.activeFileOperations.size,
          activeFiles: rtContext.activeFileOperations.size,
          lastActivity: rtContext.lastContextUpdate,
          createdAt: state.currentSession?.createdAt || new Date(),
          totalCost: state.currentSession?.totalCost || 0,
          requestCount: rtContext.operationMetrics.totalOperations,
          cacheStats: {
            cacheHits: 0, // Will be updated by backend
            cacheMisses: 0,
            totalSavings: 0,
            cacheEfficiency: 0,
          },
        };
      },
      
      getActiveOperationsCount: () => {
        const rtContext = get().realTimeContext;
        return rtContext.activeFunctionCalls.size + rtContext.activeShellCommands.size + rtContext.activeFileOperations.size;
      },
      
      getLiveOperationsByType: (type) => {
        const rtContext = get().realTimeContext;
        switch (type) {
          case 'function':
            return Array.from(rtContext.activeFunctionCalls.values());
          case 'shell':
            return Array.from(rtContext.activeShellCommands.values());
          case 'file':
            return Array.from(rtContext.activeFileOperations.values());
          default:
            return [];
        }
      },
      
      // UI Actions
      setLeftPaneWidth: (width) => set((state) => {
        state.leftPaneWidth = width;
        state.rightPaneWidth = 100 - width;
      }),
      
      setRightPaneWidth: (width) => set((state) => {
        state.rightPaneWidth = width;
        state.leftPaneWidth = 100 - width;
      }),
      
      selectFunctionCall: (id) => set((state) => {
        state.selectedFunctionCall = id;
      }),
      
      toggleTheme: () => set((state) => {
        state.theme = state.theme === 'dark' ? 'light' : 'dark';
      }),
      
      toggleSidebar: () => set((state) => {
        state.sidebarCollapsed = !state.sidebarCollapsed;
      }),
      
      // Session Actions
      startSession: (session) => set((state) => {
        state.currentSession = session;
        state.messages = [];
        state.functionCalls = [];
        state.commandExecutions = [];
        // Reset real-time context
        state.realTimeContext = {
          activeFunctionCalls: new Map(),
          activeShellCommands: new Map(),
          activeFileOperations: new Map(),
          isStreaming: false,
          contextVersion: 0,
          lastContextUpdate: new Date(),
          pendingContextChanges: [],
          operationMetrics: {
            totalOperations: 0,
            averageResponseTime: 0,
            operationsPerMinute: 0,
            lastMinuteOperations: [],
          },
        };
      }),
      
      endSession: () => set((state) => {
        state.currentSession = undefined;
        state.messages = [];
        state.functionCalls = [];
        state.commandExecutions = [];
        // Reset real-time context
        state.realTimeContext = {
          activeFunctionCalls: new Map(),
          activeShellCommands: new Map(),
          activeFileOperations: new Map(),
          isStreaming: false,
          contextVersion: 0,
          lastContextUpdate: new Date(),
          pendingContextChanges: [],
          operationMetrics: {
            totalOperations: 0,
            averageResponseTime: 0,
            operationsPerMinute: 0,
            lastMinuteOperations: [],
          },
        };
      }),
      
      // Message Actions
      addMessage: (message) => set((state) => {
        state.messages.push(message);
      }),
      
      updateMessage: (id, updates) => set((state) => {
        const messageIndex = state.messages.findIndex((m: Message) => m.id === id);
        if (messageIndex !== -1) {
          Object.assign(state.messages[messageIndex], updates);
        }
      }),
      
      clearMessages: () => set((state) => {
        state.messages = [];
      }),
      
      // Function Call Actions
      addFunctionCall: (functionCall) => set((state) => {
        state.functionCalls.push(functionCall);
      }),
      
      updateFunctionCall: (id, updates) => set((state) => {
        const functionCallIndex = state.functionCalls.findIndex((fc: FunctionCall) => fc.id === id);
        if (functionCallIndex !== -1) {
          Object.assign(state.functionCalls[functionCallIndex], updates);
        }
      }),
      
      removeFunctionCall: (id) => set((state) => {
        state.functionCalls = state.functionCalls.filter((fc: FunctionCall) => fc.id !== id);
      }),
      
      clearFunctionCalls: () => set((state) => {
        state.functionCalls = [];
      }),
      
      // Command Actions
      addCommandExecution: (execution) => set((state) => {
        state.commandExecutions.push(execution);
        if (execution.status === 'running') {
          state.activeCommands++;
        }
      }),
      
      updateCommandExecution: (id, updates) => set((state) => {
        const executionIndex = state.commandExecutions.findIndex((e: CommandExecution) => e.id === id);
        if (executionIndex !== -1) {
          const execution = state.commandExecutions[executionIndex];
          const wasRunning = execution.status === 'running';
          Object.assign(execution, updates);
          
          if (wasRunning && updates.status && updates.status !== 'running') {
            state.activeCommands = Math.max(0, state.activeCommands - 1);
          }
        }
      }),
      
      removeCommandExecution: (id) => set((state) => {
        const execution = state.commandExecutions.find((e: CommandExecution) => e.id === id);
        if (execution && execution.status === 'running') {
          state.activeCommands = Math.max(0, state.activeCommands - 1);
        }
        state.commandExecutions = state.commandExecutions.filter((e: CommandExecution) => e.id !== id);
      }),
      
      clearCommandExecutions: () => set((state) => {
        state.commandExecutions = [];
        state.activeCommands = 0;
      }),
      
      // Context Actions
      setContextState: (contextState) => set((state) => {
        state.contextState = contextState;
      }),
      
      updateProtectedMemory: (key, value) => set((state) => {
        state.protectedMemory[key] = value;
      }),
      
      removeProtectedMemory: (key) => set((state) => {
        delete state.protectedMemory[key];
      }),
      
      setProtectedMemory: (memory) => set((state) => {
        state.protectedMemory = memory;
      }),
      
      setWorkspaceState: (workspace) => set((state) => {
        state.workspaceState = workspace;
      }),
      
      updateKnowledge: (key, value) => set((state) => {
        state.knowledgeBase[key] = value;
      }),
      
      removeKnowledge: (key) => set((state) => {
        delete state.knowledgeBase[key];
      }),
      
      setKnowledgeBase: (knowledge) => set((state) => {
        state.knowledgeBase = knowledge;
      }),
      
      // AI Actions
      setSelectedModel: (modelId) => set((state) => {
        state.selectedModel = modelId;
      }),
      
      setAvailableModels: (models) => set((state) => {
        state.availableModels = models;
      }),
      
      setIsStreaming: (streaming) => set((state) => {
        state.isStreaming = streaming;
      }),
      
      // WebSocket Actions
      setConnected: (connected) => set((state) => {
        state.connected = connected;
        if (connected) {
          state.reconnectAttempts = 0;
        }
      }),
      
      incrementReconnectAttempts: () => set((state) => {
        state.reconnectAttempts++;
      }),
      
      resetReconnectAttempts: () => set((state) => {
        state.reconnectAttempts = 0;
      }),
      
      // Debug Actions
      addDebugMessage: (message) => set((state) => {
        const id = Date.now().toString();
        state.debugMessages.push({ ...message, id });
      }),
      
      clearDebugMessages: () => set((state) => {
        state.debugMessages = [];
      }),
      
      // Notification Actions
      addNotification: (notification) => set((state) => {
        const id = Date.now().toString();
        state.notifications.push({ ...notification, id });
      }),
      
      removeNotification: (id) => set((state) => {
        state.notifications = state.notifications.filter((n: Notification) => n.id !== id);
      }),
      
      clearNotifications: () => set((state) => {
        state.notifications = [];
      }),
      
      // Terminal Actions
      showTerminal: (operation) => set((state) => {
        state.isTerminalVisible = true;
        if (operation) {
          state.currentTerminalOperation = operation;
        }
      }),
      
      hideTerminal: () => set((state) => {
        state.isTerminalVisible = false;
        state.currentTerminalOperation = undefined;
      }),
      
      addShellOperation: (operation) => set((state) => {
        state.allShellOperations.push(operation);
      }),
    }))
  )
);

// Enhanced selectors for real-time operations
export const selectMessages = (state: AppStore) => state.messages;
export const selectFunctionCalls = (state: AppStore) => state.functionCalls;
export const selectSelectedFunctionCall = (state: AppStore) => 
  state.functionCalls.find(fc => fc.id === state.selectedFunctionCall);
export const selectCommandExecutions = (state: AppStore) => state.commandExecutions;
export const selectActiveCommands = (state: AppStore) => state.activeCommands;
export const selectContextState = (state: AppStore) => state.contextState;
export const selectProtectedMemory = (state: AppStore) => state.protectedMemory;
export const selectWorkspaceState = (state: AppStore) => state.workspaceState;
export const selectKnowledgeBase = (state: AppStore) => state.knowledgeBase;
export const selectCurrentSession = (state: AppStore) => state.currentSession;
export const selectIsConnected = (state: AppStore) => state.connected;
export const selectIsStreaming = (state: AppStore) => state.isStreaming;
export const selectNotifications = (state: AppStore) => state.notifications; 
export const selectDebugMessages = (state: AppStore) => state.debugMessages;

// New real-time selectors
export const selectRealTimeContext = (state: AppStore) => state.realTimeContext;
export const selectActiveFunctionCalls = (state: AppStore) => Array.from(state.realTimeContext.activeFunctionCalls.values());
export const selectActiveShellCommands = (state: AppStore) => Array.from(state.realTimeContext.activeShellCommands.values());
export const selectActiveFileOperations = (state: AppStore) => Array.from(state.realTimeContext.activeFileOperations.values());
export const selectLiveMessages = (state: AppStore) => state.messages.filter(m => m.isLive);
export const selectStreamingMessages = (state: AppStore) => state.messages.filter(m => m.isStreaming);
export const selectOperationMetrics = (state: AppStore) => state.realTimeContext.operationMetrics;
export const selectContextVersion = (state: AppStore) => state.realTimeContext.contextVersion; 