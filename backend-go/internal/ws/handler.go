package ws

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/krnskrll/event-mpos/internal/auth"
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
	// Origin is not the auth boundary here — the JWT (validated below) is. The dashboard connects
	// cross-origin via the dev proxy, so accept the handshake and rely on the token.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Handler upgrades the warehouse dashboard's WebSocket, authenticating the JWT first.
type Handler struct {
	hub *Hub
	jwt *auth.JwtService
}

func NewHandler(hub *Hub, jwt *auth.JwtService) *Handler {
	return &Handler{hub: hub, jwt: jwt}
}

// ServeWS handles GET /ws/events/:eventId/orders?token=<accessToken>. The JWT is required in the
// `token` query param (browsers can't set Authorization on a WebSocket handshake); a Bearer header
// is also accepted for non-browser clients.
func (h *Handler) ServeWS(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		if hdr := c.GetHeader("Authorization"); len(hdr) > 7 && hdr[:7] == "Bearer " {
			token = hdr[7:]
		}
	}
	if _, ok := h.jwt.Parse(token); !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": 401, "error": "Unauthorized", "message": "Authentication required",
		})
		return
	}
	eventID := c.Param("eventId")

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
