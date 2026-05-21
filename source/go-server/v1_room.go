package main

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// V1Room is the OpenTheatre v1 room model. Lives in memory only.
type V1Room struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	HostId      string `json:"hostId"`
	ControlMode string `json:"controlMode"` // "democratic" | "host-only"
	UrlSyncMode string `json:"urlSyncMode"` // "ask" | "auto" | "locked"
	CreatedAt   int64  `json:"createdAt"`   // unix milliseconds

	mu      sync.RWMutex
	members map[string]*memberEntry
	chat    []V1ChatMessage
}

// V1SettingsPatch is the body of a settings-update WS message. Empty strings
// mean "no change".
type V1SettingsPatch struct {
	ControlMode string `json:"controlMode,omitempty"`
	UrlSyncMode string `json:"urlSyncMode,omitempty"`
}

// NewV1Room constructs a room with default settings (democratic, ask).
func NewV1Room(id, name, hostId string) *V1Room {
	return &V1Room{
		Id:          id,
		Name:        name,
		HostId:      hostId,
		ControlMode: "democratic",
		UrlSyncMode: "ask",
		CreatedAt:   time.Now().UnixMilli(),
		members:     make(map[string]*memberEntry),
	}
}

// Rename changes the room name. Host-only.
func (r *V1Room) Rename(callerId, newName string) error {
	if callerId != r.HostId {
		return errors.New("not host")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Name = newName
	return nil
}

// UpdateSettings applies a partial settings change. Host-only.
func (r *V1Room) UpdateSettings(callerId string, patch V1SettingsPatch) error {
	if callerId != r.HostId {
		return errors.New("not host")
	}
	if patch.ControlMode != "" && patch.ControlMode != "democratic" && patch.ControlMode != "host-only" {
		return errors.New("invalid ControlMode")
	}
	if patch.UrlSyncMode != "" && patch.UrlSyncMode != "ask" && patch.UrlSyncMode != "auto" && patch.UrlSyncMode != "locked" {
		return errors.New("invalid UrlSyncMode")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if patch.ControlMode != "" {
		r.ControlMode = patch.ControlMode
	}
	if patch.UrlSyncMode != "" {
		r.UrlSyncMode = patch.UrlSyncMode
	}
	return nil
}

// memberEntry tracks one member plus a connection counter. Methods added in Task 4.
type memberEntry struct {
	member V1Member
	conns  int
}

// V1Member is a participant in a v1 room. Fully used from Task 4 onward.
type V1Member struct {
	Id          string `json:"id"`
	DisplayName string `json:"displayName"`
	AvatarColor string `json:"avatarColor"`
	JoinedAt    int64  `json:"joinedAt"`
}

// V1ChatMessage is one entry in the chat history. Used from Task 5 onward.
type V1ChatMessage struct {
	Id        string `json:"id"`
	MemberId  string `json:"memberId"`
	Text      string `json:"text"`
	Timestamp int64  `json:"timestamp"`
}

// AddMember registers a new connection for the given member. New members get
// JoinedAt stamped. Idempotent across reconnects (multi-tab).
func (r *V1Room) AddMember(m V1Member) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if entry, ok := r.members[m.Id]; ok {
		entry.conns++
		return
	}
	if m.JoinedAt == 0 {
		m.JoinedAt = time.Now().UnixMilli()
	}
	r.members[m.Id] = &memberEntry{member: m, conns: 1}
}

// RemoveMember decrements the connection count and removes the member if it
// hits zero. Returns true if the member was removed (last connection closed).
func (r *V1Room) RemoveMember(memberId string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	entry, ok := r.members[memberId]
	if !ok {
		return false
	}
	entry.conns--
	if entry.conns <= 0 {
		delete(r.members, memberId)
		return true
	}
	return false
}

// GetMember returns the member with the given id and whether it exists.
func (r *V1Room) GetMember(memberId string) (V1Member, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entry, ok := r.members[memberId]
	if !ok {
		return V1Member{}, false
	}
	return entry.member, true
}

// MemberCount returns the number of distinct members (not connections).
func (r *V1Room) MemberCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.members)
}

// ConnectionCount returns the number of open WS connections for memberId.
func (r *V1Room) ConnectionCount(memberId string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if entry, ok := r.members[memberId]; ok {
		return entry.conns
	}
	return 0
}

// MemberList returns a snapshot of all members. Order is not guaranteed.
func (r *V1Room) MemberList() []V1Member {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]V1Member, 0, len(r.members))
	for _, entry := range r.members {
		out = append(out, entry.member)
	}
	return out
}

// v1ChatBufferSize is the per-room cap on retained chat messages.
const v1ChatBufferSize = 100

// v1MaxChatTextLen is the per-message server-enforced cap.
const v1MaxChatTextLen = 500

// AppendChat appends a message to the ring buffer and returns the stored
// message with server-stamped Id and Timestamp. Text is truncated to
// v1MaxChatTextLen.
func (r *V1Room) AppendChat(memberId, text string) V1ChatMessage {
	if len(text) > v1MaxChatTextLen {
		text = text[:v1MaxChatTextLen]
	}
	msg := V1ChatMessage{
		Id:        uuid.New().String(),
		MemberId:  memberId,
		Text:      text,
		Timestamp: time.Now().UnixMilli(),
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.chat = append(r.chat, msg)
	if len(r.chat) > v1ChatBufferSize {
		r.chat = r.chat[len(r.chat)-v1ChatBufferSize:]
	}
	return msg
}

// ChatHistory returns an ordered snapshot copy of the chat ring buffer.
func (r *V1Room) ChatHistory() []V1ChatMessage {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]V1ChatMessage, len(r.chat))
	copy(out, r.chat)
	return out
}
