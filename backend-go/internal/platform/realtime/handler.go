package realtime

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// wsConn is the subset of *websocket.Conn used, so the hub can stay transport-agnostic in tests.
type wsConn interface {
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (int, []byte, error)
	Close() error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
	SetPongHandler(h func(string) error)
}

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 50 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Origin is not the auth boundary here — the device token (validated below) is.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Handler upgrades the pickup board's WebSocket, authenticating the device first.
type Handler struct {
	hub       *Hub
	authorize func(token, eventID string) int
}

// NewHandler takes authorize, which answers 0 when the token's device may watch the event's board,
// else the HTTP status to refuse with.
func NewHandler(hub *Hub, authorize func(token, eventID string) int) *Handler {
	return &Handler{hub: hub, authorize: authorize}
}

// ServeWS handles GET /ws/events/:eventId/orders?token=<device token>. The token goes in the query
// (a WebSocket handshake cannot always set headers); "Authorization: Device <token>" also works.
func (h *Handler) ServeWS(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		token = strings.TrimPrefix(c.GetHeader("Authorization"), "Device ")
	}
	eventID := c.Param("eventId")
	if status := h.authorize(token, eventID); status != 0 {
		c.AbortWithStatusJSON(status, gin.H{"status": status, "error": http.StatusText(status), "message": "Not allowed to watch this board"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return // Upgrade already wrote an error response
	}
	client := &Client{hub: h.hub, conn: conn, send: make(chan []byte, 256), room: eventID}
	h.hub.register <- client

	go client.writePump()
	go client.readPump()
}

// readPump drains inbound frames (the dashboard doesn't send any) and detects disconnects.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

// writePump sends queued broadcasts + periodic pings.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
