package main

import (
	"net/http"
)

// v1WsHub is fully implemented in Task 8.
type v1WsHub struct{}

// handleV1Ws is fully implemented in Task 8. Stub returns 501 for now.
func (h *slashFix) handleV1Ws(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
