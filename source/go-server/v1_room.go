package main

import (
	"errors"
	"sync"
	"time"
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
