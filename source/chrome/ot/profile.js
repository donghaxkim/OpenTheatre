// OpenTheatre client identity, persisted in localStorage under "ot.profile".
// Works as a CommonJS module (node --test) and as a content script (window.OTProfile).
(function () {
  const PROFILE_KEY = "ot.profile";

  function generateMemberId() {
    return "m_" + Math.random().toString(36).slice(2) + Date.now().toString(36);
  }

  function isValidProfile(p) {
    return !!p && typeof p.memberId === "string" && p.memberId.length > 0
      && typeof p.displayName === "string" && p.displayName.trim().length > 0
      && typeof p.avatarColor === "string" && p.avatarColor.length > 0;
  }

  // Browser-only: read the saved profile, or null if absent/invalid.
  function loadProfile() {
    try {
      const p = JSON.parse(localStorage.getItem(PROFILE_KEY));
      return isValidProfile(p) ? p : null;
    } catch { return null; }
  }

  // Browser-only: persist a profile, generating a memberId if missing.
  function saveProfile({ displayName, avatarColor, memberId }) {
    const profile = { memberId: memberId || generateMemberId(), displayName, avatarColor };
    localStorage.setItem(PROFILE_KEY, JSON.stringify(profile));
    return profile;
  }

  const api = { PROFILE_KEY, generateMemberId, isValidProfile, loadProfile, saveProfile };
  if (typeof module !== "undefined" && module.exports) module.exports = api;
  if (typeof window !== "undefined") window.OTProfile = api;
})();
