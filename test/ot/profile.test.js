const test = require("node:test");
const assert = require("node:assert");
const { generateMemberId, isValidProfile } = require("../../source/chrome/ot/profile.js");

test("generateMemberId produces a unique m_-prefixed id", () => {
  const a = generateMemberId();
  const b = generateMemberId();
  assert.match(a, /^m_[a-z0-9]+$/);
  assert.notStrictEqual(a, b);
});

test("isValidProfile requires memberId, displayName, avatarColor", () => {
  assert.strictEqual(isValidProfile({ memberId: "m_1", displayName: "Al", avatarColor: "#cf9a52" }), true);
  assert.strictEqual(isValidProfile({ memberId: "m_1", displayName: "", avatarColor: "#cf9a52" }), false);
  assert.strictEqual(isValidProfile({ displayName: "Al", avatarColor: "#cf9a52" }), false);
  assert.strictEqual(isValidProfile(null), false);
});
