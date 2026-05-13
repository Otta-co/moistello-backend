package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	ws "github.com/moistello/backend/internal/websocket"
)

// WebSocketHandler handles HTTP-to-WebSocket upgrades and manages client
// connections via the Hub.
type WebSocketHandler struct {
	hub      *ws.Hub
	upgrader websocket.Upgrader
}

// NewWebSocketHandler creates a new WebSocketHandler backed by the given Hub.
// The upgrader is configured with generous buffer sizes and a permissive
// origin check for development; restrict CheckOrigin in production.
func NewWebSocketHandler(hub *ws.Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins in development; restrict in production
				return true
			},
		},
	}
}

// HandleWebSocket upgrades the HTTP connection to a WebSocket and registers
// the new client with the Hub. The userID is extracted from the Gin context
// (set by the auth middleware) or defaults to "anonymous".
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	userID := c.GetString("userID")
	if userID == "" {
		userID = "anonymous"
	}

	client := ws.NewClient(h.hub, conn, userID)
	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}

// Hub returns the underlying Hub instance for metrics and integration.
func (h *WebSocketHandler) Hub() *ws.Hub {
	return h.hub
}
