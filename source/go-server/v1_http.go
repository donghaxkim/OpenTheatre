package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

// handleV1Rooms handles POST /v1/rooms (create).
// Body: {"member": {"id","displayName","avatarColor"}}; returns {"roomId"}.
func (h *slashFix) handleV1Rooms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Member V1Member `json:"member"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.Member.DisplayName == "" {
		http.Error(w, "member.displayName required", http.StatusBadRequest)
		return
	}
	if body.Member.Id == "" {
		http.Error(w, "member.id required", http.StatusBadRequest)
		return
	}
	room := h.v1Srv.CreateRoom(body.Member.DisplayName+"'s room", body.Member.Id)
	h.JSON(w, http.StatusOK, map[string]string{"roomId": room.Id})
}

// handleV1RoomById handles GET /v1/rooms/{id}. 404 if absent.
func (h *slashFix) handleV1RoomById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/v1/rooms/")
	if id == "" || strings.Contains(id, "/") {
		http.Error(w, "invalid room id", http.StatusBadRequest)
		return
	}
	room, ok := h.v1Srv.GetRoom(id)
	if !ok {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}
	h.JSON(w, http.StatusOK, room)
}
