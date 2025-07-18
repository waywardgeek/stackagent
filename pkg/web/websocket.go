package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"stackagent/pkg/ai"
)

// WebSocketEvent represents a message sent over WebSocket
type WebSocketEvent struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	SessionID string      `json:"sessionId"`
}

// WebSocketEventType represents the type of WebSocket event
type WebSocketEventType string

const (
	// Existing events
	EventChatMessage WebSocketEventType = "chat_message"
	EventUserMessage WebSocketEventType = "user_message"
	EventAIResponse  WebSocketEventType = "ai_response"
	EventAIError     WebSocketEventType = "ai_error"
	EventPing        WebSocketEventType = "ping"
	EventPong        WebSocketEventType = "pong"
	
	// New Phase 2 streaming events
	EventFunctionCallStarted    WebSocketEventType = "function_call_started"
	EventFunctionCallStreaming  WebSocketEventType = "function_call_streaming"
	EventFunctionCallCompleted  WebSocketEventType = "function_call_completed"
	EventFunctionCallFailed     WebSocketEventType = "function_call_failed"
	EventShellCommandStarted    WebSocketEventType = "shell_command_started"
	EventShellCommandStreaming  WebSocketEventType = "shell_command_streaming"
	EventShellCommandCompleted  WebSocketEventType = "shell_command_completed"
	EventFileOperationStarted   WebSocketEventType = "file_operation_started"
	EventFileOperationStreaming WebSocketEventType = "file_operation_streaming"
	EventFileOperationCompleted WebSocketEventType = "file_operation_completed"
	EventAIStreaming            WebSocketEventType = "ai_streaming"
	EventConfigureStreaming     WebSocketEventType = "configure_streaming"
)

// StreamingCallback represents a callback for streaming events
type StreamingCallback func(eventType WebSocketEventType, data interface{}, sessionID string)

// Use ConversationMessage from ai package
type ConversationMessage = ai.ConversationMessage

// ConversationContext holds the conversation history for a session
type ConversationContext struct {
	SessionID    string                `json:"sessionId"`
	Messages     []ConversationMessage `json:"messages"`
	CreatedAt    time.Time             `json:"createdAt"`
	UpdatedAt    time.Time             `json:"updatedAt"`
	TotalCost    float64               `json:"totalCost"`
	RequestCount int                   `json:"requestCount"`
	CacheStats   struct {
		CacheHits        int     `json:"cacheHits"`
		CacheMisses      int     `json:"cacheMisses"`
		TotalSavings     float64 `json:"totalSavings"`
		CacheEfficiency  float64 `json:"cacheEfficiency"`
	} `json:"cacheStats"`
	
	// New Phase 2 fields for streaming
	ActiveOperations map[string]interface{} `json:"activeOperations"`
	StreamingEnabled bool                   `json:"streamingEnabled"`
	
	mutex        sync.RWMutex
}

// AddMessage adds a message to the conversation context
func (c *ConversationContext) AddMessage(role, content string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.Messages = append(c.Messages, ConversationMessage{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
	c.UpdatedAt = time.Now()
}

// AddCost adds cost information for a request
func (c *ConversationContext) AddCost(cost ai.TokenCost) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.TotalCost += cost.TotalCost
	c.RequestCount++
	c.CacheStats.TotalSavings += cost.CostBreakdown.CacheSavings
	
	// Track cache hits/misses
	if cost.CacheReadInputTokens > 0 {
		c.CacheStats.CacheHits++
	} else {
		c.CacheStats.CacheMisses++
	}
	
	// Calculate cache efficiency (% of requests that hit cache)
	if c.RequestCount > 0 {
		c.CacheStats.CacheEfficiency = (float64(c.CacheStats.CacheHits) / float64(c.RequestCount)) * 100
	}
	
	c.UpdatedAt = time.Now()
}

// GetMessages returns a copy of the conversation messages
func (c *ConversationContext) GetMessages() []ConversationMessage {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	messages := make([]ConversationMessage, len(c.Messages))
	copy(messages, c.Messages)
	return messages
}

// SetStreamingEnabled enables or disables streaming for this context
func (c *ConversationContext) SetStreamingEnabled(enabled bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.StreamingEnabled = enabled
}

// AddActiveOperation adds an active operation to track
func (c *ConversationContext) AddActiveOperation(operationID string, operation interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.ActiveOperations == nil {
		c.ActiveOperations = make(map[string]interface{})
	}
	c.ActiveOperations[operationID] = operation
}

// RemoveActiveOperation removes an active operation
func (c *ConversationContext) RemoveActiveOperation(operationID string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.ActiveOperations != nil {
		delete(c.ActiveOperations, operationID)
	}
}

// ClearMessages clears all conversation messages
func (c *ConversationContext) ClearMessages() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.Messages = []ConversationMessage{}
	c.UpdatedAt = time.Now()
}

// WebSocketServer manages WebSocket connections and conversations
type WebSocketServer struct {
	upgrader      websocket.Upgrader
	clients       map[*websocket.Conn]string
	conversations map[string]*ConversationContext
	claude        *ai.ClaudeClient
	mutex         sync.RWMutex
	
	// New Phase 2 streaming support
	streamingCallbacks map[string]StreamingCallback // sessionID -> callback
	streamingMutex     sync.RWMutex
}

// NewWebSocketServer creates a new WebSocket server
func NewWebSocketServer() *WebSocketServer {
	// Initialize Claude client
	claude, err := ai.NewClaudeClient()
	if err != nil {
		log.Printf("Warning: Failed to initialize Claude client: %v", err)
		log.Printf("Chat functionality will not be available")
	}

	return &WebSocketServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow connections from any origin for development
				// In production, you should restrict this
				return true
			},
		},
		clients:       make(map[*websocket.Conn]string),
		conversations: make(map[string]*ConversationContext),
		claude:        claude,
		
		// Initialize Phase 2 streaming support
		streamingCallbacks: make(map[string]StreamingCallback),
	}
}

// GetConversationContext returns the conversation context for a session
func (ws *WebSocketServer) GetConversationContext(sessionID string) *ConversationContext {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()
	return ws.conversations[sessionID]
}

// SendToClient sends a message to a specific client
func (ws *WebSocketServer) SendToClient(client *websocket.Conn, event WebSocketEvent) error {
	message, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return client.WriteMessage(websocket.TextMessage, message)
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

// HandleWebSocket handles WebSocket connections
func (ws *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Generate a unique session ID
	sessionID := generateSessionID()
	
	// Register the connection with session ID
	ws.mutex.Lock()
	ws.clients[conn] = sessionID
	// Create new conversation context for this session
	ws.conversations[sessionID] = &ConversationContext{
		SessionID: sessionID,
		Messages:  []ConversationMessage{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		
		// Initialize Phase 2 streaming support
		ActiveOperations: make(map[string]interface{}),
		StreamingEnabled: true,
	}
	ws.mutex.Unlock()

	log.Printf("Client connected with session ID: %s", sessionID)

	// Send session started event
	sessionStartedEvent := WebSocketEvent{
		Type: "session_started",
		Data: map[string]interface{}{
			"sessionId": sessionID,
			"message":   "WebSocket connection established",
		},
		Timestamp: time.Now(),
		SessionID: sessionID,
	}
	ws.SendToClient(conn, sessionStartedEvent)

	// Handle incoming messages
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			} else {
				log.Printf("WebSocket connection closed: %v", err)
			}
			break
		}

		// Parse the message
		var event WebSocketEvent
		if err := json.Unmarshal(message, &event); err != nil {
			log.Printf("JSON unmarshal error: %v", err)
			continue
		}

		// Set session ID if not provided
		if event.SessionID == "" {
			event.SessionID = sessionID
		}

		// Handle the event
		ws.handleEvent(conn, event)
	}

	// Unregister the connection and clean up conversation
	ws.mutex.Lock()
	delete(ws.clients, conn)
	delete(ws.conversations, sessionID)
	ws.mutex.Unlock()

	log.Printf("Client disconnected, session ID: %s", sessionID)
	conn.Close()
}

// handleEvent processes events from clients
func (ws *WebSocketServer) handleEvent(client *websocket.Conn, event WebSocketEvent) {
	log.Printf("Received event from client: %s", event.Type)

	switch event.Type {
	case "ping":
		// Respond with pong
		pong := WebSocketEvent{
			Type:      "pong",
			Data:      nil,
			Timestamp: time.Now(),
			SessionID: event.SessionID,
		}
		ws.SendToClient(client, pong)

	case "get_context":
		// Get the actual conversation context for this session
		context := ws.GetConversationContext(event.SessionID)
		if context == nil {
			log.Printf("No conversation context found for session: %s", event.SessionID)
			return
		}
		
		// Send context information with real data
		contextData := map[string]interface{}{
			"sessionId":        event.SessionID,
			"memoryEntries":    len(context.Messages),
			"knowledgeEntries": 0,
			"commandHistory":   0,
			"activeHandles":    0,
			"activeFiles":      0,
			"lastActivity":     context.UpdatedAt,
			"createdAt":        context.CreatedAt,
		}
		
		response := WebSocketEvent{
			Type:      "context_updated",
			Data:      map[string]interface{}{"contextState": contextData},
			Timestamp: time.Now(),
			SessionID: event.SessionID,
		}
		ws.SendToClient(client, response)

	case "chat_message":
		// Handle chat message
		ws.handleChatMessage(client, event)

	default:
		log.Printf("Unknown event type: %s", event.Type)
	}
}

// handleChatMessage processes chat messages from clients
func (ws *WebSocketServer) handleChatMessage(client *websocket.Conn, event WebSocketEvent) {
	// Extract message from event data
	data, ok := event.Data.(map[string]interface{})
	if !ok {
		log.Printf("Invalid chat message data format")
		return
	}

	message, ok := data["message"].(string)
	if !ok {
		log.Printf("Invalid chat message format")
		return
	}

	messageID, ok := data["id"].(string)
	if !ok {
		log.Printf("Invalid chat message ID format")
		return
	}

	log.Printf("Received chat message: %s", message)

	// Handle session ID race condition
	actualSessionID := event.SessionID
	if actualSessionID == "current" {
		// Find the actual session ID for this connection
		ws.mutex.RLock()
		if realSessionID, exists := ws.clients[client]; exists {
			actualSessionID = realSessionID
			log.Printf("Correcting session ID from 'current' to '%s'", actualSessionID)
		}
		ws.mutex.RUnlock()
	}

	// Send user message confirmation with corrected session ID
	userMessageResponse := WebSocketEvent{
		Type: "user_message",
		Data: map[string]interface{}{
			"id":        messageID,
			"message":   message,
			"timestamp": time.Now(),
		},
		Timestamp: time.Now(),
		SessionID: actualSessionID,
	}
	ws.SendToClient(client, userMessageResponse)

	// Check if Claude client is available
	if ws.claude == nil {
		errorResponse := WebSocketEvent{
			Type: "ai_error",
			Data: map[string]interface{}{
				"error": "Claude client not initialized. Please set ANTHROPIC_API_KEY environment variable.",
			},
			Timestamp: time.Now(),
			SessionID: actualSessionID,
		}
		ws.SendToClient(client, errorResponse)
		return
	}

	// Get conversation context and add user message
	context := ws.GetConversationContext(actualSessionID)
	if context == nil {
		// Create context if it doesn't exist
		ws.mutex.Lock()
		context = &ConversationContext{
			SessionID: actualSessionID,
			Messages:  []ConversationMessage{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			
			// Initialize Phase 2 streaming support
			ActiveOperations: make(map[string]interface{}),
			StreamingEnabled: true,
		}
		ws.conversations[actualSessionID] = context
		ws.mutex.Unlock()
	}

	// Check if there are messages in the "current" session that need to be transferred
	if actualSessionID != "current" {
		ws.mutex.Lock()
		if currentContext, exists := ws.conversations["current"]; exists && len(currentContext.Messages) > 0 {
			log.Printf("Transferring %d messages from 'current' session to '%s'", len(currentContext.Messages), actualSessionID)
			// Transfer messages from "current" to actual session
			for _, msg := range currentContext.Messages {
				context.Messages = append(context.Messages, msg)
			}
			// Clear the current session
			delete(ws.conversations, "current")
		}
		ws.mutex.Unlock()
	}

	context.AddMessage("user", message)
	
	// Send updated context information after adding user message
	contextData := map[string]interface{}{
		"sessionId":        actualSessionID,
		"memoryEntries":    len(context.Messages),
		"knowledgeEntries": 0,
		"commandHistory":   0,
		"activeHandles":    0,
		"activeFiles":      0,
		"lastActivity":     context.UpdatedAt,
		"createdAt":        context.CreatedAt,
	}
	
	contextResponse := WebSocketEvent{
		Type:      "context_updated",
		Data:      map[string]interface{}{"contextState": contextData},
		Timestamp: time.Now(),
		SessionID: actualSessionID,
	}
	ws.SendToClient(client, contextResponse)

	// Send AI response in a goroutine to avoid blocking
	go func() {
		var response string
		var cost ai.TokenCost
		var operationSummary ai.OperationSummary
		var err error
		
		// EMERGENCY: Disable all streaming to fix infinite loop
		// Set up basic debug callback only
		ws.claude.SetDebugCallback(func(eventType string, data interface{}) {
			debugEvent := WebSocketEvent{
				Type: "debug_message",
				Data: map[string]interface{}{
					"type":    eventType,
					"message": fmt.Sprintf("Function call: %s", eventType),
					"data":    data,
				},
				Timestamp: time.Now(),
				SessionID: actualSessionID,
			}
			ws.SendToClient(client, debugEvent)
		})
		
		// Disable streaming callback entirely
		ws.claude.SetStreamingCallback(nil)
		
		// Get conversation history
		messages := context.GetMessages()
		
		// Send debug information about what we're sending to Claude
		cachedComponents := []string{"system_prompt", "tool_definitions"}
		if len(messages) > 1 {
			cachedComponents = append(cachedComponents, "conversation_history")
		}
		
		debugInfo := map[string]interface{}{
			"type":        "claude_api_request",
			"message":     "Sending request to Claude API with advanced caching",
			"messageCount": len(messages),
			"messages":    messages,
			"hasSystemPrompt": true,
			"systemPrompt":    "You are StackAgent, a helpful AI coding assistant with access to powerful file manipulation and shell command tools. Available functions: run_with_capture (shell commands), read_file (read files), write_file (create/write files), edit_file (find/replace in files), search_in_file (search with context), list_directory (list files with filters). Use these functions to efficiently help with coding tasks, file operations, and system administration. Be concise but helpful. Remember context from previous messages in this conversation.\n\nCore principle: Don't be evil. Always prioritize user safety, privacy, and ethical behavior.",
			"cachingEnabled": true,
			"cachedComponents": cachedComponents,
			"costReduction": "Up to 90% for cached content (including conversation history and file content)",
		}
		
		debugEvent := WebSocketEvent{
			Type: "debug_message",
			Data: debugInfo,
			Timestamp: time.Now(),
			SessionID: actualSessionID,
		}
		ws.SendToClient(client, debugEvent)
		
		// Call Claude API with context
		response, cost, operationSummary, err = ws.claude.ChatWithToolsAndContext(messages)
		if err != nil {
			log.Printf("Claude API error: %v", err)
			
			// Send error response to client
			errorResponse := WebSocketEvent{
				Type: "ai_error", 
				Data: map[string]interface{}{
					"error": fmt.Sprintf("Failed to get AI response: %v", err),
				},
				Timestamp: time.Now(),
				SessionID: actualSessionID,
			}
			ws.SendToClient(client, errorResponse)
			return
		}

		// Add AI response to conversation context
		context.AddMessage("assistant", response)
		
		// Add cost information to context
		context.AddCost(cost)

		// Send AI response with cost and operation summary information
		aiResponse := WebSocketEvent{
			Type: "ai_response",
			Data: map[string]interface{}{
				"message":          response,
				"timestamp":        time.Now(),
				"cost":            cost,
				"operationSummary": operationSummary,
			},
			Timestamp: time.Now(),
			SessionID: actualSessionID,
		}
		ws.SendToClient(client, aiResponse)

		// Send updated context information with cost data
		contextData := map[string]interface{}{
			"sessionId":        actualSessionID,
			"memoryEntries":    len(context.Messages),
			"knowledgeEntries": 0,
			"commandHistory":   0,
			"activeHandles":    0,
			"activeFiles":      0,
			"lastActivity":     context.UpdatedAt,
			"createdAt":        context.CreatedAt,
			"totalCost":        context.TotalCost,
			"requestCount":     context.RequestCount,
			"cacheStats":       context.CacheStats,
		}
		
		contextResponse := WebSocketEvent{
			Type:      "context_updated",
			Data:      map[string]interface{}{"contextState": contextData},
			Timestamp: time.Now(),
			SessionID: actualSessionID,
		}
		ws.SendToClient(client, contextResponse)
	}()
}

// GetClientCount returns the number of connected clients
func (ws *WebSocketServer) GetClientCount() int {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()
	return len(ws.clients)
}

// SetStreamingCallback sets the streaming callback for a session
func (ws *WebSocketServer) SetStreamingCallback(sessionID string, callback StreamingCallback) {
	ws.streamingMutex.Lock()
	defer ws.streamingMutex.Unlock()
	ws.streamingCallbacks[sessionID] = callback
}

// RemoveStreamingCallback removes the streaming callback for a session
func (ws *WebSocketServer) RemoveStreamingCallback(sessionID string) {
	ws.streamingMutex.Lock()
	defer ws.streamingMutex.Unlock()
	delete(ws.streamingCallbacks, sessionID)
}

// SendStreamingEvent sends a streaming event to the client
func (ws *WebSocketServer) SendStreamingEvent(sessionID string, eventType WebSocketEventType, data interface{}) {
	// Find the client connection for this session
	var client *websocket.Conn
	ws.mutex.RLock()
	for conn, sid := range ws.clients {
		if sid == sessionID {
			client = conn
			break
		}
	}
	ws.mutex.RUnlock()
	
	if client == nil {
		log.Printf("No client found for session ID: %s", sessionID)
		return
	}
	
	// Send the streaming event
	event := WebSocketEvent{
		Type:      string(eventType),
		Data:      data,
		Timestamp: time.Now(),
		SessionID: sessionID,
	}
	
	ws.SendToClient(client, event)
	
	// Also call the streaming callback if it exists
	ws.streamingMutex.RLock()
	if callback, exists := ws.streamingCallbacks[sessionID]; exists {
		callback(eventType, data, sessionID)
	}
	ws.streamingMutex.RUnlock()
} 