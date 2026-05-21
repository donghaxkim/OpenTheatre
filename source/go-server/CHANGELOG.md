# OpenTheatre server changelog

## v1 (in progress)

- Added `POST /v1/rooms`, `GET /v1/rooms/{id}` for OpenTheatre rooms.
- Added `WS /v1/ws/{roomId}` for chat, typing, reactions, presence, settings.
- Room model: 6-char codes, democratic/host-only control, in-memory chat
  ring buffer (last 100), multi-tab member presence, 30s reconnect grace.
- Legacy VideoTogether endpoints preserved for vt.js compatibility.
- Removed broken legacy VT Go test files.
- `config.json` is now optional — the server falls back to an empty
  configuration instead of panicking when it is absent.
