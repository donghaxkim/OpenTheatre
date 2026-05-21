package main

import (
	"sync"
)

// V1Service is the in-memory store of OpenTheatre v1 rooms, keyed by 6-char Id.
type V1Service struct {
	rooms sync.Map // map[string]*V1Room
}

// NewV1Service constructs an empty service.
func NewV1Service() *V1Service {
	return &V1Service{}
}

// CreateRoom generates a fresh unique id, constructs and stores a V1Room.
func (s *V1Service) CreateRoom(name, hostId string) *V1Room {
	for {
		id := GenerateV1RoomCode()
		if _, loaded := s.rooms.LoadOrStore(id, (*V1Room)(nil)); loaded {
			continue // collision, retry
		}
		room := NewV1Room(id, name, hostId)
		s.rooms.Store(id, room)
		return room
	}
}

// GetRoom returns the room with the given id and whether it exists.
func (s *V1Service) GetRoom(id string) (*V1Room, bool) {
	v, ok := s.rooms.Load(id)
	if !ok || v == nil {
		return nil, false
	}
	room, ok := v.(*V1Room)
	if !ok || room == nil {
		return nil, false
	}
	return room, true
}

// DeleteRoom removes the room. No-op if absent.
func (s *V1Service) DeleteRoom(id string) {
	s.rooms.Delete(id)
}

// RoomCount returns the number of active rooms.
func (s *V1Service) RoomCount() int {
	n := 0
	s.rooms.Range(func(_, v any) bool {
		if room, ok := v.(*V1Room); ok && room != nil {
			n++
		}
		return true
	})
	return n
}
