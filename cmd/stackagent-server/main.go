package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"stackagent/pkg/web"
)

func main() {
	// Create WebSocket server
	wsServer := web.NewWebSocketServer()
	
	// Create HTTP server
	mux := http.NewServeMux()
	
	// Handle WebSocket connections
	mux.HandleFunc("/ws", wsServer.HandleWebSocket)
	
	// Serve static files from the GUI build directory
	guiPath := filepath.Join("web", "gui", "dist")
	if _, err := os.Stat(guiPath); os.IsNotExist(err) {
		log.Printf("GUI build directory not found: %s", guiPath)
		log.Println("To build the GUI, run: cd web/gui && npm install && npm run build")
		
		// Serve a simple HTML page instead
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			html := `
<!DOCTYPE html>
<html>
<head>
    <title>StackAgent</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #0f172a; color: #f1f5f9; }
        .container { max-width: 600px; margin: 0 auto; text-align: center; }
        .logo { font-size: 48px; margin-bottom: 20px; }
        .title { font-size: 24px; margin-bottom: 10px; }
        .subtitle { font-size: 16px; color: #94a3b8; margin-bottom: 30px; }
        .instructions { background: #1e293b; padding: 20px; border-radius: 8px; text-align: left; }
        .code { background: #334155; padding: 10px; border-radius: 4px; font-family: monospace; }
        .status { color: #10b981; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">🚀</div>
        <h1 class="title">StackAgent Server</h1>
        <p class="subtitle">Revolutionary AI coding assistant with persistent memory</p>
        
        <div class="status">✅ Server is running on port 8080</div>
        <div class="status">✅ WebSocket endpoint available at /ws</div>
        
        <div class="instructions">
            <h3>To build and run the GUI:</h3>
            <ol>
                <li>Navigate to the GUI directory:
                    <div class="code">cd web/gui</div>
                </li>
                <li>Install dependencies:
                    <div class="code">npm install</div>
                </li>
                <li>Build the GUI:
                    <div class="code">npm run build</div>
                </li>
                <li>Restart the server:
                    <div class="code">go run cmd/stackagent-server/main.go</div>
                </li>
            </ol>
        </div>
        
        <div class="instructions" style="margin-top: 20px;">
            <h3>Features:</h3>
            <ul>
                <li>✅ Dual-pane GUI with resizable layout</li>
                <li>✅ Real-time WebSocket communication</li>
                <li>✅ Context persistence across sessions</li>
                <li>✅ Git branch-specific AI memory</li>
                <li>✅ Function call transparency</li>
                <li>✅ Command execution tracking</li>
                <li>✅ Dark/light theme support</li>
                <li>✅ Keyboard shortcuts</li>
            </ul>
        </div>
    </div>
</body>
</html>
			`
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(html))
		})
	} else {
		// Serve the built GUI
		fs := http.FileServer(http.Dir(guiPath))
		mux.Handle("/", fs)
		log.Printf("Serving GUI from: %s", guiPath)
	}
	
	// API endpoints
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := fmt.Sprintf(`{
			"status": "ok",
			"timestamp": "%s",
			"clients": %d,
			"version": "1.0.0"
		}`, fmt.Sprintf("%d", os.Getpid()), wsServer.GetClientCount())
		w.Write([]byte(response))
	})
	
	// Start HTTP server with graceful shutdown
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}
	
	// Create HTTP server with timeouts
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	
	log.Printf("🚀 StackAgent Server starting on port %s", port)
	log.Printf("📡 WebSocket endpoint: ws://localhost:%s/ws", port)
	log.Printf("🌐 GUI available at: http://localhost:%s", port)
	log.Printf("🔧 API health check: http://localhost:%s/api/health", port)
	log.Printf("⚖️  Core principle: Don't be evil")
	
	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()
	
	// Set up signal handling for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	// Wait for shutdown signal
	<-quit
	log.Println("🛑 Shutting down server...")
	
	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
	} else {
		log.Println("✅ Server gracefully stopped")
	}
	
	// Clean up WebSocket connections
	log.Println("🧹 Cleaning up WebSocket connections...")
	wsServer.Shutdown()
	
	log.Println("👋 Goodbye!")
} 