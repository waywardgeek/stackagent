import { useCallback, useEffect, useRef, useState } from 'react';
import { useAppStore } from '@/store';
import type { WebSocketEvent, WebSocketEventType } from '@/types';

interface UseWebSocketOptions {
  url?: string;
  enableRealTimeStreaming?: boolean;
  streamingBufferSize?: number;
  reconnectInterval?: number;
}

export const useWebSocket = (options: UseWebSocketOptions = {}) => {
  const {
    url = 'ws://localhost:8080/ws',
    enableRealTimeStreaming = true,
    streamingBufferSize = 1024,
    reconnectInterval = 1000,
  } = options;
  
  const websocket = useRef<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);
  const [sessionId, setSessionId] = useState<string>('current');
  const isManuallyDisconnected = useRef(false);
  
  // Real-time streaming state
  const streamingBuffer = useRef<Map<string, string>>(new Map());
  const streamingTimers = useRef<Map<string, NodeJS.Timeout>>(new Map());
  const [streamingConnections, setStreamingConnections] = useState<Set<string>>(new Set());

  const store = useAppStore();
  const storeRef = useRef(store);
  
  // Update store ref when store changes
  useEffect(() => {
    storeRef.current = store;
  }, [store]);

  // Real-time streaming helpers
  const startStreamingOperation = useCallback((operationId: string, type: 'function' | 'shell' | 'file') => {
    if (!enableRealTimeStreaming) return;
    
    setStreamingConnections(prev => new Set(prev).add(operationId));
    storeRef.current.setStreamingState(true, operationId, type);
    
    // Set up streaming timeout
    const timeout = setTimeout(() => {
      completeStreamingOperation(operationId);
    }, 30000); // 30 second timeout
    
    streamingTimers.current.set(operationId, timeout);
  }, [enableRealTimeStreaming]);

  const updateStreamingOperation = useCallback((operationId: string, data: string) => {
    if (!enableRealTimeStreaming) return;
    
    // Buffer the streaming data
    const currentBuffer = streamingBuffer.current.get(operationId) || '';
    const newBuffer = currentBuffer + data;
    
    // Limit buffer size
    const truncatedBuffer = newBuffer.length > streamingBufferSize ? 
      newBuffer.slice(-streamingBufferSize) : newBuffer;
    
    streamingBuffer.current.set(operationId, truncatedBuffer);
    
    // Update live operation in store
    storeRef.current.updateLiveOperation(operationId, {
      output: truncatedBuffer,
      timestamp: new Date(),
    });
  }, [enableRealTimeStreaming, streamingBufferSize]);

  const completeStreamingOperation = useCallback((operationId: string) => {
    if (!enableRealTimeStreaming) return;
    
    setStreamingConnections(prev => {
      const newSet = new Set(prev);
      newSet.delete(operationId);
      return newSet;
    });
    
    // Clear streaming data
    streamingBuffer.current.delete(operationId);
    
    // Clear timeout
    const timer = streamingTimers.current.get(operationId);
    if (timer) {
      clearTimeout(timer);
      streamingTimers.current.delete(operationId);
    }
    
    // Complete operation in store
    storeRef.current.completeLiveOperation(operationId);
    storeRef.current.setStreamingState(false, operationId);
  }, [enableRealTimeStreaming]);

  // Handle incoming WebSocket messages
  const handleMessage = useCallback((event: MessageEvent) => {
    try {
      const data = JSON.parse(event.data) as WebSocketEvent;
      
      // Log debug message
      storeRef.current.addDebugMessage({
        timestamp: new Date(),
        direction: 'received',
        type: 'websocket',
        event: data.type,
        data: data,
        rawJson: event.data,
      });
      
      switch (data.type) {
        case 'ping':
          // Respond to server ping to keep connection alive
          if (websocket.current?.readyState === WebSocket.OPEN) {
            try {
              websocket.current.send(JSON.stringify({
                type: 'pong',
                data: null,
                timestamp: new Date(),
                sessionId: data.sessionId,
              }));
            } catch (error) {
              console.error('Failed to send pong:', error);
            }
          }
          break;
          
        case 'session_started':
          setIsConnected(true);
          storeRef.current.addNotification({
            type: 'success',
            title: 'Session Started',
            message: `New session: ${data.data.sessionId}`,
            timestamp: new Date(),
          });
          // Store the actual session ID
          setSessionId(data.data.sessionId);
          break;
          
        case 'session_ended':
          storeRef.current.addNotification({
            type: 'info',
            title: 'Session Ended',
            message: 'AI session has ended',
            timestamp: new Date(),
          });
          break;
          
        case 'message_received':
          storeRef.current.addMessage(data.data);
          break;
          
        case 'function_call_started':
          const functionCall = {
            ...data.data,
            status: 'running',
            startTime: new Date(),
          };
          
          storeRef.current.addFunctionCall(functionCall);
          
          // Start real-time streaming for this operation
          startStreamingOperation(data.data.id, 'function');
          break;
          
        case 'function_call_completed':
          storeRef.current.updateFunctionCall(data.data.id, {
            status: 'completed',
            result: data.data.result,
            endTime: new Date(),
            duration: data.data.duration,
          });
          
          // Complete streaming operation
          completeStreamingOperation(data.data.id);
          break;
          
        case 'function_call_failed':
          storeRef.current.updateFunctionCall(data.data.id, {
            status: 'failed',
            error: data.data.error,
            endTime: new Date(),
            duration: data.data.duration,
          });
          
          // Complete streaming operation
          completeStreamingOperation(data.data.id);
          break;
          
        // NEW: Real-time streaming events
        case 'function_call_streaming':
          updateStreamingOperation(data.data.id, data.data.output || '');
          break;
          
        case 'shell_command_started':
          const shellOp = {
            ...data.data,
            timestamp: new Date(),
          };
          
          storeRef.current.startLiveOperation(data.data.id, 'shell', shellOp);
          startStreamingOperation(data.data.id, 'shell');
          break;
          
        case 'shell_command_streaming':
          updateStreamingOperation(data.data.id, data.data.output || '');
          break;
          
        case 'shell_command_completed':
          storeRef.current.updateLiveOperation(data.data.id, {
            output: data.data.output,
            exitCode: data.data.exitCode,
            duration: data.data.duration,
          });
          
          completeStreamingOperation(data.data.id);
          break;
          
        case 'file_operation_started':
          const fileOp = {
            ...data.data,
            timestamp: new Date(),
          };
          
          storeRef.current.startLiveOperation(data.data.id, 'file', fileOp);
          startStreamingOperation(data.data.id, 'file');
          break;
          
        case 'file_operation_streaming':
          updateStreamingOperation(data.data.id, data.data.content || '');
          break;
          
        case 'file_operation_completed':
          storeRef.current.updateLiveOperation(data.data.id, {
            content: data.data.content,
            changes: data.data.changes,
            size: data.data.size,
          });
          
          completeStreamingOperation(data.data.id);
          break;
          
        case 'context_updated':
          if (data.data.contextState) {
            storeRef.current.setContextState(data.data.contextState);
          }
          if (data.data.protectedMemory) {
            storeRef.current.setProtectedMemory(data.data.protectedMemory);
          }
          if (data.data.workspaceState) {
            storeRef.current.setWorkspaceState(data.data.workspaceState);
          }
          if (data.data.knowledgeBase) {
            storeRef.current.setKnowledgeBase(data.data.knowledgeBase);
          }
          
          // Update context version
          storeRef.current.updateContextVersion();
          break;
          
        case 'command_started':
          storeRef.current.addCommandExecution({
            ...data.data,
            status: 'running',
            startTime: new Date(),
          });
          break;
          
        case 'command_completed':
          storeRef.current.updateCommandExecution(data.data.id, {
            status: 'completed',
            output: data.data.output,
            exitCode: data.data.exitCode,
            endTime: new Date(),
            duration: data.data.duration,
          });
          break;
          
        case 'command_output':
          storeRef.current.updateCommandExecution(data.data.id, {
            output: data.data.output,
            lineCount: data.data.lineCount,
          });
          break;
          
        case 'git_branch_changed':
          storeRef.current.addNotification({
            type: 'info',
            title: 'Git Branch Changed',
            message: `Switched to branch: ${data.data.branch}`,
            timestamp: new Date(),
          });
          break;
          
        case 'cost_updated':
          storeRef.current.addNotification({
            type: 'info',
            title: 'Cost Updated',
            message: `Session cost: $${data.data.totalCost.toFixed(4)}`,
            timestamp: new Date(),
          });
          break;
          
        case 'error_occurred':
          storeRef.current.addNotification({
            type: 'error',
            title: 'Error',
            message: data.data.message || 'An error occurred',
            timestamp: new Date(),
          });
          break;
          
        case 'ai_response':
          // Create live message during AI response
          storeRef.current.addLiveMessage({
            sessionId: data.sessionId || 'current',
            type: 'assistant',
            content: data.data.message,
            timestamp: new Date(data.data.timestamp),
            operationSummary: data.data.operationSummary,
            cost: data.data.cost?.totalCost,
            tokens: data.data.cost?.inputTokens + data.data.cost?.outputTokens,
            isLive: false, // Mark as completed
          });
          
          // Add shell operations to the store for terminal pane
          if (data.data.operationSummary?.shellCommands) {
            data.data.operationSummary.shellCommands.forEach((shellOp: any) => {
              storeRef.current.addShellOperation(shellOp);
            });
          }
          
          // Track operation metrics
          if (data.data.operationSummary) {
            const operationCount = (data.data.operationSummary.shellCommands?.length || 0) + 
                                 (data.data.operationSummary.fileOperations?.length || 0);
            
            if (operationCount > 0) {
              storeRef.current.trackOperationMetrics(data.data.duration || 0);
            }
          }
          break;
          
        case 'ai_streaming':
          // Handle streaming AI response
          const streamingMessageId = data.data.messageId || Date.now().toString();
          
          storeRef.current.updateMessageStreaming(streamingMessageId, true, data.data.progress);
          
          // Add partial content to live message
          if (data.data.partialContent) {
            storeRef.current.updateMessage(streamingMessageId, {
              content: data.data.partialContent,
              timestamp: new Date(),
            });
          }
          break;
          
        case 'ai_error':
          // Show AI error notification
          storeRef.current.addNotification({
            type: 'error',
            title: 'AI Error',
            message: data.data.error,
            timestamp: new Date(),
          });
          break;
          
        case 'user_message':
          // User message confirmation - update with correct session ID if needed
          const userMessage = storeRef.current.messages.find(m => m.id === data.data.id);
          if (userMessage && userMessage.sessionId !== data.sessionId) {
            storeRef.current.updateMessage(data.data.id, {
              sessionId: data.sessionId,
            });
          }
          break;
          
        case 'debug_message':
          // Log the debug message
          storeRef.current.addDebugMessage({
            timestamp: new Date(),
            direction: 'received',
            type: 'debug_message',
            event: data.type,
            data: data,
            rawJson: event.data,
          });
          break;
          
        default:
          console.warn('Unknown WebSocket event type:', data.type);
      }
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error);
      storeRef.current.addNotification({
        type: 'error',
        title: 'WebSocket Error',
        message: 'Failed to parse server message',
        timestamp: new Date(),
      });
    }
  }, [startStreamingOperation, updateStreamingOperation, completeStreamingOperation]);

  // Handle WebSocket connection open
  const handleOpen = useCallback(() => {
    setIsConnected(true);
    setIsConnecting(false);
    console.log('WebSocket connected successfully');
    
    // Send initial streaming config
    if (enableRealTimeStreaming) {
      sendMessage('configure_streaming', {
        enabled: true,
        bufferSize: streamingBufferSize,
      });
    }
  }, [enableRealTimeStreaming, streamingBufferSize]);

  // Handle WebSocket errors
  const handleError = useCallback((error: Event) => {
    console.error('WebSocket error:', error);
    setIsConnecting(false);
    
    storeRef.current.addNotification({
      type: 'error',
      title: 'Connection Error',
      message: 'Failed to connect to StackAgent backend',
      timestamp: new Date(),
    });
  }, []);

  // Handle WebSocket connection close
  const handleClose = useCallback(() => {
    setIsConnected(false);
    setIsConnecting(false);
    setSessionId('current');
    
    // Clean up all streaming operations
    streamingConnections.forEach(operationId => {
      completeStreamingOperation(operationId);
    });
    setStreamingConnections(new Set());
    
    // Clean up timers
    streamingTimers.current.forEach(timer => clearTimeout(timer));
    streamingTimers.current.clear();
    
    console.log('WebSocket disconnected');
    
    // Auto-reconnect if not manually disconnected
    if (!isManuallyDisconnected.current && enableRealTimeStreaming) {
      setTimeout(() => {
        if (!isManuallyDisconnected.current) {
          connect();
        }
      }, reconnectInterval);
    }
  }, [completeStreamingOperation, enableRealTimeStreaming, reconnectInterval]);

  // Connect to WebSocket
  const connect = useCallback(() => {
    if (websocket.current?.readyState === WebSocket.OPEN) {
      return;
    }
    
    if (websocket.current?.readyState === WebSocket.CONNECTING) {
      return;
    }
    
    if (isConnecting) {
      return;
    }
    
    isManuallyDisconnected.current = false;
    setIsConnecting(true);
    
    try {
      websocket.current = new WebSocket(url);
      
      websocket.current.addEventListener('open', handleOpen);
      websocket.current.addEventListener('message', handleMessage);
      websocket.current.addEventListener('close', handleClose);
      websocket.current.addEventListener('error', handleError);
      
    } catch (error) {
      console.error('Failed to create WebSocket:', error);
      setIsConnecting(false);
      handleError(error as Event);
    }
  }, [url, isConnecting, handleOpen, handleMessage, handleClose, handleError]);

  // Disconnect from WebSocket
  const disconnect = useCallback(() => {
    isManuallyDisconnected.current = true;
    
    if (websocket.current) {
      websocket.current.removeEventListener('open', handleOpen);
      websocket.current.removeEventListener('message', handleMessage);
      websocket.current.removeEventListener('close', handleClose);
      websocket.current.removeEventListener('error', handleError);
      
      websocket.current.close();
      websocket.current = null;
    }
    
    setIsConnected(false);
    setIsConnecting(false);
    setSessionId('current');
    
    // Clean up streaming operations
    streamingConnections.forEach(operationId => {
      completeStreamingOperation(operationId);
    });
    setStreamingConnections(new Set());
  }, [handleOpen, handleMessage, handleClose, handleError, completeStreamingOperation]);

  // Send message to WebSocket
  const sendMessage = (type: WebSocketEventType, data: any) => {
    const ws = websocket.current;
    if (ws && ws.readyState === WebSocket.OPEN) {
      const message: WebSocketEvent = {
        type,
        data,
        timestamp: new Date(),
        sessionId: sessionId,
      };
      
      try {
        const jsonString = JSON.stringify(message);
        ws.send(jsonString);
        
        // Log debug message
        storeRef.current.addDebugMessage({
          timestamp: new Date(),
          direction: 'sent',
          type: 'websocket',
          event: type,
          data: message,
          rawJson: jsonString,
        });
      } catch (error) {
        console.error('Error sending message:', error);
      }
    } else {
      console.warn('WebSocket not connected - cannot send message');
    }
  };

  // Real-time streaming control
  const enableStreaming = useCallback(() => {
    sendMessage('configure_streaming', { enabled: true });
  }, []);

  const disableStreaming = useCallback(() => {
    sendMessage('configure_streaming', { enabled: false });
  }, []);

  // Auto-connect on mount
  useEffect(() => {
    // connect(); // Removed auto-connect to prevent infinite loop
  }, []);
  
  // Cleanup on unmount
  useEffect(() => {
    return () => {
      disconnect();
    };
  }, [disconnect]);

  return {
    isConnected,
    isConnecting,
    connect,
    disconnect,
    sendMessage,
    
    // Real-time streaming features
    enableStreaming,
    disableStreaming,
    streamingConnections: Array.from(streamingConnections),
    isStreamingEnabled: enableRealTimeStreaming,
  };
}; 