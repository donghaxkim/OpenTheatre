package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// v1Upgrader accepts WebSocket upgrades. CheckOrigin is permissive because
// the extension injects from arbitrary host pages.
var v1Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// V1WsEnvelope is the JSON wrapper for every WS message in either direction.
type V1WsEnvelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// v1WsClient is one open WebSocket connection.
type v1WsClient struct {
	hub      *v1WsHub
	conn     *websocket.Conn
	roomId   string
	memberId string // set after a "join" message
	send     chan []byte
}

// v1WsHub fans out messages within rooms.
type v1WsHub struct {
	v1Srv *V1Service

	mu            sync.RWMutex
	clientsByRoom map[string]map[*v1WsClient]struct{}
}

func newV1WsHub(v1Srv *V1Service) *v1WsHub {
	return &v1WsHub{
		v1Srv:         v1Srv,
		clientsByRoom: make(map[string]map[*v1WsClient]struct{}),
	}
}

func (h *v1WsHub) register(c *v1WsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	set, ok := h.clientsByRoom[c.roomId]
	if !ok {
		set = make(map[*v1WsClient]struct{})
		h.clientsByRoom[c.roomId] = set
	}
	set[c] = struct{}{}
}

func (h *v1WsHub) unregister(c *v1WsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	set, ok := h.clientsByRoom[c.roomId]
	if !ok {
		return
	}
	delete(set, c)
	if len(set) == 0 {
		delete(h.clientsByRoom, c.roomId)
	}
}

// broadcast sends payload to every client connected to roomId.
func (h *v1WsHub) broadcast(roomId string, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clientsByRoom[roomId] {
		select {
		case c.send <- payload:
		default:
		}
	}
}

// ensureV1WsHub lazily constructs the singleton hub on the slashFix struct.
func (h *slashFix) ensureV1WsHub() *v1WsHub {
	h.v1WsOnce.Do(func() {
		h.v1Ws = newV1WsHub(h.v1Srv)
	})
	return h.v1Ws
}

// handleV1Ws handles GET /v1/ws/{roomId} → WebSocket upgrade.
func (h *slashFix) handleV1Ws(w http.ResponseWriter, r *http.Request) {
	roomId := strings.TrimPrefix(r.URL.Path, "/v1/ws/")
	if roomId == "" || strings.Contains(roomId, "/") {
		http.Error(w, "invalid room id", http.StatusBadRequest)
		return
	}
	if _, ok := h.v1Srv.GetRoom(roomId); !ok {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	conn, err := v1Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	hub := h.ensureV1WsHub()
	client := &v1WsClient{
		hub:    hub,
		conn:   conn,
		roomId: roomId,
		send:   make(chan []byte, 32),
	}
	hub.register(client)

	go client.writePump()
	go client.readPump()
}

// readPump reads + dispatches messages. Per-type handlers added in later tasks.
func (c *v1WsClient) readPump() {
	defer func() {
		c.hub.unregister(c)
		c.conn.Close()
	}()
	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		var env V1WsEnvelope
		if err := json.Unmarshal(raw, &env); err != nil {
			continue
		}
		switch env.Type {
		// Per-type cases added in Tasks 9-13.
		}
	}
}

// writePump pushes outbound messages to the connection.
func (c *v1WsClient) writePump() {
	defer c.conn.Close()
	for payload := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			log.Println("v1 ws write error:", err)
			return
		}
	}
}
