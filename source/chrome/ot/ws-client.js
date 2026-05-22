// WebSocket client for /v1/ws/{roomId}.
// Pure helpers (encodeEnvelope/parseEnvelope) are exercised by node --test;
// the OTWsClient class is browser-only. Exposed as window.OTWsClient.
(function () {
  function encodeEnvelope(type, data) {
    return JSON.stringify({ type, data: data || {} });
  }

  function parseEnvelope(raw) {
    try {
      const env = JSON.parse(raw);
      if (!env || typeof env.type !== "string") return null;
      return { type: env.type, data: env.data };
    } catch { return null; }
  }

  // Browser-only. Opens the WS, sends join, dispatches typed events to listeners.
  class OTWsClient {
    constructor(wsBase, roomId) {
      this.url = `${wsBase}/v1/ws/${roomId}`;
      this.listeners = {}; // type -> [fn]
      this.ws = null;
    }
    on(type, fn) { (this.listeners[type] || (this.listeners[type] = [])).push(fn); }
    _emit(type, data) { (this.listeners[type] || []).forEach((fn) => fn(data)); }

    connect(member) {
      this.ws = new WebSocket(this.url);
      this.ws.addEventListener("open", () => {
        this.ws.send(encodeEnvelope("join", { member }));
        this._emit("open");
      });
      this.ws.addEventListener("message", (e) => {
        const env = parseEnvelope(e.data);
        if (env) this._emit(env.type, env.data);
      });
      this.ws.addEventListener("close", () => this._emit("close"));
      this.ws.addEventListener("error", () => this._emit("error"));
    }
    send(type, data) {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.ws.send(encodeEnvelope(type, data));
      }
    }
    disconnect() { if (this.ws) this.ws.close(); }
  }

  const api = { encodeEnvelope, parseEnvelope, OTWsClient };
  if (typeof module !== "undefined" && module.exports) module.exports = api;
  if (typeof window !== "undefined") window.OTWsClient = api;
})();
