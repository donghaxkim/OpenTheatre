// Single source of the OpenTheatre server origin.
// Dev: the Go v1 server runs locally on :5001 (see source/go-server/main.go).
// These values are mirrored in source/extension/background.js (room create)
// and ot/content.js (WebSocket connection) — keep them in sync.
const OT_SERVER_HTTP = "http://localhost:5001";
const OT_SERVER_WS = "ws://localhost:5001";
