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
	EventChatMessage WebSocketEventType = "chat_message"
	EventUserMessage WebSocketEventType = "user_message"
	EventAIResponse  WebSocketEventType = "ai_response"
	EventAIError     WebSocketEventType = "ai_error"
	EventPing        WebSocketEventType = "ping"
	EventPong        WebSocketEventType = "pong"
)

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
		var err error
		
		// Get conversation history
		messages := context.GetMessages()
		
		// Send debug information about what we're sending to Claude
		debugInfo := map[string]interface{}{
			"type":        "claude_api_request",
			"message":     "Sending request to Claude API with prompt caching",
			"messageCount": len(messages),
			"messages":    messages,
			"hasSystemPrompt": true,
			"systemPrompt":    "You are StackAgent, a helpful AI coding assistant with access to powerful file manipulation and shell command tools. Available functions: run_with_capture (shell commands), read_file (read files), write_file (create/write files), edit_file (find/replace in files), search_in_file (search with context), list_directory (list files with filters). Use these functions to efficiently help with coding tasks, file operations, and system administration. Be concise but helpful. Remember context from previous messages in this conversation.",
			"cachingEnabled": true,
			"cachedComponents": []string{"system_prompt", "tool_definitions"},
			"costReduction": "Up to 90% for cached content",
		}
		
		debugEvent := WebSocketEvent{
			Type: "debug_message",
			Data: debugInfo,
			Timestamp: time.Now(),
			SessionID: actualSessionID,
		}
		ws.SendToClient(client, debugEvent)
		
		// Convert to Claude format and send with context
		response, cost, err = ws.claude.ChatWithToolsAndContext(messages)
		
		if err != nil {
			log.Printf("Claude API error: %v", err)
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

		// Send AI response with cost information
		aiResponse := WebSocketEvent{
			Type: "ai_response",
			Data: map[string]interface{}{
				"message":   response,
				"timestamp": time.Now(),
				"cost":      cost,
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