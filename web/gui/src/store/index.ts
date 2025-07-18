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
  ShellOperation
} from '@/types';

interface AppStore extends AppState {
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
    immer((set) => ({
      // Initial UI State
      leftPaneWidth: 50,
      rightPaneWidth: 50,
      selectedFunctionCall: undefined,
      theme: 'dark',
      sidebarCollapsed: false,
      
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
      }),
      
      endSession: () => set((state) => {
        state.currentSession = undefined;
        state.messages = [];
        state.functionCalls = [];
        state.commandExecutions = [];
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

// Selectors for common state access patterns
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