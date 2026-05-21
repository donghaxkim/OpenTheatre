package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

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
	typingTimers  map[typingTimerKey]*time.Timer
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

// readPump reads + dispatches messages.
func (c *v1WsClient) readPump() {
	defer func() {
		c.hub.unregister(c)
		c.conn.Close()
		c.hub.onClientGone(c)
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
		case "join":
			c.handleJoin(env.Data)
		case "chat":
			c.handleChat(env.Data)
		case "typing-start":
			c.handleTypingStart()
		case "typing-stop":
			c.handleTypingStop()
		case "reaction-burst":
			c.handleReactionBurst(env.Data)
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

// handleJoin registers the member, sends room-state to this client, and
// broadcasts member-joined to others.
func (c *v1WsClient) handleJoin(data json.RawMessage) {
	var payload struct {
		Member V1Member `json:"member"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return
	}
	if payload.Member.Id == "" {
		return
	}
	room, ok := c.hub.v1Srv.GetRoom(c.roomId)
	if !ok {
		return
	}
	c.memberId = payload.Member.Id
	wasNew := room.ConnectionCount(payload.Member.Id) == 0
	room.AddMember(payload.Member)

	c.sendEnvelope("room-state", buildRoomStateSnapshot(room))

	if wasNew {
		c.hub.broadcastExcept(c.roomId, c, "member-joined", map[string]any{
			"member": payload.Member,
		})
	}
}

// v1RoomStateSnapshot is the payload of a room-state message.
type v1RoomStateSnapshot struct {
	Id          string          `json:"id"`
	Name        string          `json:"name"`
	HostId      string          `json:"hostId"`
	ControlMode string          `json:"controlMode"`
	UrlSyncMode string          `json:"urlSyncMode"`
	Members     []V1Member      `json:"members"`
	ChatHistory []V1ChatMessage `json:"chatHistory"`
}

func buildRoomStateSnapshot(r *V1Room) v1RoomStateSnapshot {
	return v1RoomStateSnapshot{
		Id:          r.Id,
		Name:        r.Name,
		HostId:      r.HostId,
		ControlMode: r.ControlMode,
		UrlSyncMode: r.UrlSyncMode,
		Members:     r.MemberList(),
		ChatHistory: r.ChatHistory(),
	}
}

// sendEnvelope marshals and queues an envelope for this client only.
func (c *v1WsClient) sendEnvelope(msgType string, data any) {
	rawData, err := json.Marshal(data)
	if err != nil {
		return
	}
	payload, err := json.Marshal(V1WsEnvelope{Type: msgType, Data: rawData})
	if err != nil {
		return
	}
	select {
	case c.send <- payload:
	default:
	}
}

// broadcastExcept fans data out to every client in roomId except skip.
func (h *v1WsHub) broadcastExcept(roomId string, skip *v1WsClient, msgType string, data any) {
	rawData, err := json.Marshal(data)
	if err != nil {
		return
	}
	payload, err := json.Marshal(V1WsEnvelope{Type: msgType, Data: rawData})
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clientsByRoom[roomId] {
		if client == skip {
			continue
		}
		select {
		case client.send <- payload:
		default:
		}
	}
}

// broadcastAll fans data out to every client in roomId, including the sender.
func (h *v1WsHub) broadcastAll(roomId string, msgType string, data any) {
	rawData, err := json.Marshal(data)
	if err != nil {
		return
	}
	payload, err := json.Marshal(V1WsEnvelope{Type: msgType, Data: rawData})
	if err != nil {
		return
	}
	h.broadcast(roomId, payload)
}

// onClientGone is invoked from readPump's defer after the connection closes.
// Removes the member; broadcasts member-left when their last connection went.
// Task 14 extends this with the reconnect-grace teardown.
func (h *v1WsHub) onClientGone(c *v1WsClient) {
	if c.memberId == "" {
		return
	}
	room, ok := h.v1Srv.GetRoom(c.roomId)
	if !ok {
		return
	}
	if last := room.RemoveMember(c.memberId); last {
		h.broadcastExcept(c.roomId, c, "member-left", map[string]any{
			"memberId": c.memberId,
		})
	}
}

// handleChat appends a chat message to the room buffer and broadcasts it to
// all clients in the room (including sender, as confirmation).
func (c *v1WsClient) handleChat(data json.RawMessage) {
	if c.memberId == "" {
		return
	}
	var payload struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return
	}
	if payload.Text == "" {
		return
	}
	room, ok := c.hub.v1Srv.GetRoom(c.roomId)
	if !ok {
		return
	}
	msg := room.AppendChat(c.memberId, payload.Text)
	c.hub.broadcastAll(c.roomId, "chat", map[string]any{"message": msg})
}

// v1TypingExpireAfter is how long after typing-start the server auto-emits
// typing-stop if no explicit stop arrives.
const v1TypingExpireAfter = 5 * time.Second

// typingTimerKey identifies a per-room-per-member typing-stop timer.
type typingTimerKey struct {
	roomId   string
	memberId string
}

func (c *v1WsClient) handleTypingStart() {
	if c.memberId == "" {
		return
	}
	c.hub.broadcastExcept(c.roomId, c, "typing-start", map[string]any{"memberId": c.memberId})
	c.hub.armTypingExpire(c.roomId, c.memberId)
}

func (c *v1WsClient) handleTypingStop() {
	if c.memberId == "" {
		return
	}
	c.hub.cancelTypingExpire(c.roomId, c.memberId)
	c.hub.broadcastExcept(c.roomId, c, "typing-stop", map[string]any{"memberId": c.memberId})
}

// armTypingExpire (re)starts the 5s auto-stop timer for (roomId, memberId).
func (h *v1WsHub) armTypingExpire(roomId, memberId string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.typingTimers == nil {
		h.typingTimers = make(map[typingTimerKey]*time.Timer)
	}
	key := typingTimerKey{roomId, memberId}
	if t, ok := h.typingTimers[key]; ok {
		t.Stop()
	}
	h.typingTimers[key] = time.AfterFunc(v1TypingExpireAfter, func() {
		h.mu.Lock()
		delete(h.typingTimers, key)
		h.mu.Unlock()
		h.broadcastAll(roomId, "typing-stop", map[string]any{"memberId": memberId})
	})
}

// cancelTypingExpire stops any pending typing-stop timer for (roomId, memberId).
func (h *v1WsHub) cancelTypingExpire(roomId, memberId string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	key := typingTimerKey{roomId, memberId}
	if t, ok := h.typingTimers[key]; ok {
		t.Stop()
		delete(h.typingTimers, key)
	}
}

// handleReactionBurst broadcasts an ephemeral emoji burst to the whole room.
func (c *v1WsClient) handleReactionBurst(data json.RawMessage) {
	if c.memberId == "" {
		return
	}
	var payload struct {
		Emoji string `json:"emoji"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return
	}
	if payload.Emoji == "" {
		return
	}
	c.hub.broadcastAll(c.roomId, "reaction-burst", map[string]any{
		"memberId": c.memberId,
		"emoji":    payload.Emoji,
	})
}
