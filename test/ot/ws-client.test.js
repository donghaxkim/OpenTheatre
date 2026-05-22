const test = require("node:test");
const assert = require("node:assert");
const { encodeEnvelope, parseEnvelope } = require("../../source/chrome/ot/ws-client.js");

test("encodeEnvelope wraps type + data as JSON", () => {
  const out = encodeEnvelope("join", { member: { id: "m_1" } });
  assert.strictEqual(out, '{"type":"join","data":{"member":{"id":"m_1"}}}');
});

test("parseEnvelope returns {type,data} for valid input", () => {
  const env = parseEnvelope('{"type":"room-state","data":{"id":"B3K7M9"}}');
  assert.deepStrictEqual(env, { type: "room-state", data: { id: "B3K7M9" } });
});

test("parseEnvelope returns null for malformed input", () => {
  assert.strictEqual(parseEnvelope("not json"), null);
  assert.strictEqual(parseEnvelope('{"no":"type"}'), null);
});
