// Package ws is the real-time order push for the warehouse dashboard. It replaces the Kotlin
// STOMP-over-SockJS broker with a native gorilla/websocket hub: clients connect per event room and
// receive WarehouseOrderView JSON messages as orders are created/advanced. Only the web dashboard
// uses this; the Android app does not.
package ws

import (
	"encoding/json"
	"log"
)

// Client is a single dashboard connection, scoped to one event room.
type Client struct {
	hub  *Hub
	conn wsConn
	send chan []byte
	room string
}

type broadcastMsg struct {
	room    string
	payload []byte
}

// Hub owns the room→clients registry. All mutations happen in Run's single goroutine, so no locks
// are needed on the map.
type Hub struct {
	rooms      map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan broadcastMsg
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan broadcastMsg, 64),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			if h.rooms[c.room] == nil {
				h.rooms[c.room] = make(map[*Client]bool)
			}
			h.rooms[c.room][c] = true

		case c := <-h.unregister:
			if clients, ok := h.rooms[c.room]; ok {
				if _, ok := clients[c]; ok {
					delete(clients, c)
					close(c.send)
					if len(clients) == 0 {
						delete(h.rooms, c.room)
					}
				}
			}

		case msg := <-h.broadcast:
			for c := range h.rooms[msg.room] {
				select {
				case c.send <- msg.payload:
				default:
					// slow/dead client — drop it
					close(c.send)
					delete(h.rooms[msg.room], c)
				}
			}
		}
	}
}

// Broadcast pushes a payload (marshaled to JSON) to every client in the room. Called from request
// goroutines after an order is committed; mirrors Kotlin's convertAndSend to /topic/events/{id}/orders.
func (h *Hub) Broadcast(room string, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("ws broadcast marshal failed: %v", err)
		return
	}
	h.broadcast <- broadcastMsg{room: room, payload: data}
}
