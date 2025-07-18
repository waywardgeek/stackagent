import { useCallback, useEffect, useRef, useState } from 'react';
import { useAppStore } from '@/store';
import type { WebSocketEvent, WebSocketEventType } from '@/types';

interface UseWebSocketOptions {
  url?: string;
}

export const useWebSocket = (options: UseWebSocketOptions = {}) => {
  const {
    url = 'ws://localhost:8080/ws',
  } = options;
  
  const websocket = useRef<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);
  const [sessionId, setSessionId] = useState<string>('current'); // Store actual session ID
  const isManuallyDisconnected = useRef(false);

  const store = useAppStore();
  const storeRef = useRef(store);
  
  // Update store ref when store changes
  useEffect(() => {
    storeRef.current = store;
  }, [store]);

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
          storeRef.current.addFunctionCall({
            ...data.data,
            status: 'running',
            startTime: new Date(),
          });
          break;
          
        case 'function_call_completed':
          storeRef.current.updateFunctionCall(data.data.id, {
            status: 'completed',
            result: data.data.result,
            endTime: new Date(),
            duration: data.data.duration,
          });
          break;
          
        case 'function_call_failed':
          storeRef.current.updateFunctionCall(data.data.id, {
            status: 'failed',
            error: data.data.error,
            endTime: new Date(),
            duration: data.data.duration,
          });
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
          // Add AI response message to store
          storeRef.current.addMessage({
            id: Date.now().toString(),
            sessionId: data.sessionId || 'current',
            type: 'assistant',
            content: data.data.message,
            timestamp: new Date(data.data.timestamp),
          });
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
  }, []);

  // Handle WebSocket connection open
  const handleOpen = useCallback(() => {
    setIsConnected(true);
    setIsConnecting(false);
    console.log('WebSocket connected successfully');
  }, []);

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

  // Removed attemptReconnect function to simplify dependency chain

  // Handle WebSocket connection close
  const handleClose = useCallback(() => {
    setIsConnected(false);
    setIsConnecting(false);
    setSessionId('current'); // Reset session ID on disconnect
    console.log('WebSocket disconnected');
  }, []);

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
    setSessionId('current'); // Reset session ID on disconnect
  }, [handleOpen, handleMessage, handleClose, handleError]);

  // Send message to WebSocket
  const sendMessage = (type: WebSocketEventType, data: any) => {
    const ws = websocket.current;
    if (ws && ws.readyState === WebSocket.OPEN) {
      const message: WebSocketEvent = {
        type,
        data,
        timestamp: new Date(),
        sessionId: sessionId, // Use the actual session ID
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
  };
}; 