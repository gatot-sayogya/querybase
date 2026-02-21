package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/yourorg/querybase/internal/service"
)

// WebSocketMessage represents a message sent/received via WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WebSocketHub maintains active client connections
type WebSocketHub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		clients:    make(map[*websocket.Conn]bool),
	}
}

// Run starts the WebSocket hub
func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("WebSocket client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
				log.Printf("WebSocket client disconnected. Total clients: %d", len(h.clients))
			}
		}
	}
}

// Broadcast sends a message to all connected clients
func (h *WebSocketHub) Broadcast(message []byte) {
	for client := range h.clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Error sending message to client: %v", err)
			h.unregister <- client
		}
	}
}

// WebSocketUpgradeConfig configures WebSocket upgrade
var WebSocketUpgradeConfig = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// WebSocketHandler handles WebSocket connections for schema updates
type WebSocketHandler struct {
	hub           *WebSocketHub
	schemaService *service.SchemaService
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *WebSocketHub, schemaService *service.SchemaService) *WebSocketHandler {
	return &WebSocketHandler{
		hub:           hub,
		schemaService: schemaService,
	}
}

// HandleWebSocket handles WebSocket connection upgrades
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := WebSocketUpgradeConfig.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Register client
	h.hub.register <- conn

	// Ensure client is unregistered when connection closes
	defer func() {
		h.hub.unregister <- conn
	}()

	// Send welcome message
	welcomeMsg := WebSocketMessage{
		Type: "connected",
		Payload: map[string]string{
			"message": "WebSocket connection established",
		},
	}
	welcomeBytes, _ := json.Marshal(welcomeMsg)
	conn.WriteMessage(websocket.TextMessage, welcomeBytes)

	// Handle incoming messages
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		if messageType == websocket.TextMessage {
			var wsMsg WebSocketMessage
			if err := json.Unmarshal(message, &wsMsg); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			h.handleMessage(c.Request.Context(), conn, &wsMsg)
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (h *WebSocketHandler) handleMessage(ctx context.Context, conn *websocket.Conn, msg *WebSocketMessage) {
	switch msg.Type {
	case "get_schema":
		// Client requests schema for a data source
		payload, ok := msg.Payload.(map[string]interface{})
		if !ok {
			h.sendError(conn, "Invalid payload format")
			return
		}

		dataSourceID, ok := payload["data_source_id"].(string)
		if !ok {
			h.sendError(conn, "data_source_id is required")
			return
		}

		schema, err := h.schemaService.GetSchema(ctx, dataSourceID)
		if err != nil {
			h.sendError(conn, err.Error())
			return
		}

		// Send schema back to client
		response := WebSocketMessage{
			Type:    "schema",
			Payload: schema,
		}
		responseBytes, _ := json.Marshal(response)
		conn.WriteMessage(websocket.TextMessage, responseBytes)

	case "subscribe_schema":
		// Subscribe to schema updates for a data source
		// For now, just acknowledge subscription
		ackMsg := WebSocketMessage{
			Type: "subscribed",
			Payload: map[string]string{
				"message": "Subscribed to schema updates",
			},
		}
		ackBytes, _ := json.Marshal(ackMsg)
		conn.WriteMessage(websocket.TextMessage, ackBytes)

	case "subscribe_stats":
		// Subscribe to global dashboard stat updates
		ackMsg := WebSocketMessage{
			Type: "subscribed_stats",
			Payload: map[string]string{
				"message": "Subscribed to dashboard stats updates",
			},
		}
		ackBytes, _ := json.Marshal(ackMsg)
		conn.WriteMessage(websocket.TextMessage, ackBytes)

	default:
		h.sendError(conn, "Unknown message type: "+msg.Type)
	}
}

// sendError sends an error message to the client
func (h *WebSocketHandler) sendError(conn *websocket.Conn, errMsg string) {
	errorMsg := WebSocketMessage{
		Type: "error",
		Payload: map[string]string{
			"error": errMsg,
		},
	}
	errorBytes, _ := json.Marshal(errorMsg)
	conn.WriteMessage(websocket.TextMessage, errorBytes)
}

// BroadcastSchemaUpdate broadcasts a schema update to all connected clients
func (h *WebSocketHandler) BroadcastSchemaUpdate(dataSourceID string, schema interface{}) {
	message := WebSocketMessage{
		Type: "schema_update",
		Payload: map[string]interface{}{
			"data_source_id": dataSourceID,
			"schema":         schema,
		},
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling schema update: %v", err)
		return
	}

	h.hub.Broadcast(messageBytes)
}

// BroadcastStatsChanged broadcasts a notification that stats have changed
func (h *WebSocketHandler) BroadcastStatsChanged() {
	message := WebSocketMessage{
		Type: "stats_changed",
		Payload: map[string]string{
			"message": "Dashboard statistics have been updated",
		},
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling stats_changed: %v", err)
		return
	}

	h.hub.Broadcast(messageBytes)
}
