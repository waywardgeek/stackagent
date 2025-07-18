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

// WebSocketServer handles WebSocket connections for the GUI
type WebSocketServer struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.RWMutex
	claude     *ai.ClaudeClient
}

// WebSocketEvent represents a WebSocket message
type WebSocketEvent struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	SessionID string      `json:"sessionId"`
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
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		claude:     claude,
	}
}

// Run starts the WebSocket server
func (ws *WebSocketServer) Run() {
	for {
		select {
		case client := <-ws.register:
			ws.mutex.Lock()
			ws.clients[client] = true
			ws.mutex.Unlock()
			log.Printf("Client connected. Total clients: %d", len(ws.clients))
			
			// Send welcome message
			welcome := WebSocketEvent{
				Type:      "session_started",
				Data:      map[string]interface{}{"sessionId": fmt.Sprintf("session_%d", time.Now().Unix())},
				Timestamp: time.Now(),
				SessionID: "current",
			}
			ws.SendToClient(client, welcome)

		case client := <-ws.unregister:
			ws.mutex.Lock()
			if _, ok := ws.clients[client]; ok {
				delete(ws.clients, client)
				client.Close()
			}
			ws.mutex.Unlock()
			log.Printf("Client disconnected. Total clients: %d", len(ws.clients))

		case message := <-ws.broadcast:
			ws.mutex.RLock()
			var failedClients []*websocket.Conn
			for client := range ws.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					client.Close()
					failedClients = append(failedClients, client)
				}
			}
			ws.mutex.RUnlock()
			
			// Remove failed clients
			if len(failedClients) > 0 {
				ws.mutex.Lock()
				for _, client := range failedClients {
					delete(ws.clients, client)
				}
				ws.mutex.Unlock()
			}
		}
	}
}

// SendToClient sends a message to a specific client
func (ws *WebSocketServer) SendToClient(client *websocket.Conn, event WebSocketEvent) error {
	message, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return client.WriteMessage(websocket.TextMessage, message)
}

// BroadcastEvent sends an event to all connected clients
func (ws *WebSocketServer) BroadcastEvent(eventType string, data interface{}) {
	event := WebSocketEvent{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
		SessionID: "current",
	}

	message, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return
	}

	select {
	case ws.broadcast <- message:
	default:
		log.Printf("Broadcast channel full, dropping message")
	}
}

// HandleWebSocket handles WebSocket connections
func (ws *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Allow connections from any origin for development
			// In production, you should restrict this
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Register the client
	ws.register <- conn

	// Handle messages from the client
	go func() {
		defer func() {
			log.Printf("WebSocket handler exiting")
			ws.unregister <- conn
			conn.Close()
		}()

		log.Printf("WebSocket message loop started")
		for {
			_, messageBytes, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				} else {
					log.Printf("WebSocket connection closed: %v", err)
				}
				break
			}

			// Handle incoming message
			var event WebSocketEvent
			if err := json.Unmarshal(messageBytes, &event); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			// Process the event
			ws.handleClientEvent(conn, event)
		}
	}()
}

// handleClientEvent processes events from clients
func (ws *WebSocketServer) handleClientEvent(client *websocket.Conn, event WebSocketEvent) {
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
		// Send context information
		contextData := map[string]interface{}{
			"sessionId":        "current",
			"memoryEntries":    0,
			"knowledgeEntries": 0,
			"commandHistory":   0,
			"activeHandles":    0,
			"activeFiles":      0,
			"lastActivity":     time.Now(),
			"createdAt":        time.Now(),
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

	// Send user message confirmation
	userMessageResponse := WebSocketEvent{
		Type: "user_message",
		Data: map[string]interface{}{
			"id":        messageID,
			"message":   message,
			"timestamp": time.Now(),
		},
		Timestamp: time.Now(),
		SessionID: event.SessionID,
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
			SessionID: event.SessionID,
		}
		ws.SendToClient(client, errorResponse)
		return
	}

	// Send AI response in a goroutine to avoid blocking
	go func() {
		response, err := ws.claude.ChatWithTools(message)
		if err != nil {
			log.Printf("Claude API error: %v", err)
			errorResponse := WebSocketEvent{
				Type: "ai_error",
				Data: map[string]interface{}{
					"error": fmt.Sprintf("Failed to get AI response: %v", err),
				},
				Timestamp: time.Now(),
				SessionID: event.SessionID,
			}
			ws.SendToClient(client, errorResponse)
			return
		}

		// Send AI response
		aiResponse := WebSocketEvent{
			Type: "ai_response",
			Data: map[string]interface{}{
				"message":   response,
				"timestamp": time.Now(),
			},
			Timestamp: time.Now(),
			SessionID: event.SessionID,
		}
		ws.SendToClient(client, aiResponse)
	}()
}

// GetClientCount returns the number of connected clients
func (ws *WebSocketServer) GetClientCount() int {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()
	return len(ws.clients)
} 