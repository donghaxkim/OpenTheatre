// OpenTheatre v2 content script. Owns the WebSocket, the in-memory room
// mirror, and the self profile. Translates WS events into sidebar renders and
// sidebar events into WS sends. Runs in the ISOLATED content-script world.
(function () {
  const SERVER_WS = "ws://localhost:5001"; // mirror of ot/config.js OT_SERVER_WS

  // --- hide the legacy VideoTogether panel UI ---------------------------
  // vt.js stays loaded (the sync engine is preserved); only its old panel,
  // #openTheatreLoading and #openTheatreFlyPannel, is hidden. A stylesheet
  // applies even to elements vt.js injects after us.
  function hideLegacyPanel() {
    if (document.getElementById("ot-hide-legacy")) return;
    const s = document.createElement("style");
    s.id = "ot-hide-legacy";
    s.textContent = "#openTheatreLoading,#openTheatreFlyPannel{display:none !important;}";
    (document.head || document.documentElement).appendChild(s);
  }
  hideLegacyPanel();
  document.addEventListener("DOMContentLoaded", hideLegacyPanel);

  // --- room state -------------------------------------------------------
  let joined = false;
  let ws = null;
  let reconnectTimer = null;
  let self = null; // {id, displayName, avatarColor}
  const state = { roomId: "", name: "", hostId: "", members: {} }; // members: id -> member
  const typing = {}; // memberId -> true

  function memberView(m) {
    return {
      id: m.id, displayName: m.displayName, avatarColor: m.avatarColor,
      isHost: m.id === state.hostId, isSelf: !!self && m.id === self.id
    };
  }
  function presenceList() {
    return Object.keys(state.members).map((id) => memberView(state.members[id]));
  }
  function authorOf(memberId) {
    const m = state.members[memberId];
    if (m) return { displayName: m.displayName, avatarColor: m.avatarColor };
    if (self && memberId === self.id) {
      return { displayName: self.displayName, avatarColor: self.avatarColor };
    }
    return { displayName: "Someone", avatarColor: "#6b6862" };
  }
  function isSelf(memberId) { return !!self && memberId === self.id; }

  function refreshTyping() {
    const names = Object.keys(typing)
      .filter((id) => !isSelf(id))
      .map((id) => authorOf(id).displayName);
    let text = "";
    if (names.length === 1) text = names[0] + " is typing";
    else if (names.length === 2) text = names[0] + " and " + names[1] + " are typing";
    else if (names.length > 2) text = "Several people are typing";
    window.OTSidebar.setTyping(text);
  }

  // --- WebSocket --------------------------------------------------------
  function wireWs(client) {
    client.on("open", () => window.OTSidebar.setConnection(""));
    client.on("close", scheduleReconnect);
    client.on("error", () => {});
    client.on("room-state", onRoomState);
    client.on("member-joined", onMemberJoined);
    client.on("member-left", onMemberLeft);
    client.on("chat", onChat);
    client.on("typing-start", (d) => { if (d && d.memberId) { typing[d.memberId] = true; refreshTyping(); } });
    client.on("typing-stop", (d) => { if (d && d.memberId) { delete typing[d.memberId]; refreshTyping(); } });
    client.on("reaction-burst", (d) => { if (d && d.emoji) window.OTOverlay.burst(d.emoji); });
    client.on("room-renamed", (d) => {
      if (d && d.name) { state.name = d.name; window.OTSidebar.setRoomName(d.name); }
    });
  }
  function connectWs() {
    ws = new window.OTWsClient.OTWsClient(SERVER_WS, state.roomId);
    wireWs(ws);
    ws.connect(self);
  }
  function scheduleReconnect() {
    window.OTSidebar.setConnection("Connection lost — reconnecting…");
    clearTimeout(reconnectTimer);
    reconnectTimer = setTimeout(connectWs, 2500);
  }
  function send(type, data) { if (ws) ws.send(type, data); }

  // --- WS event handlers ------------------------------------------------
  function onRoomState(room) {
    if (!room) return;
    state.name = room.name;
    state.hostId = room.hostId;
    state.members = {};
    (room.members || []).forEach((m) => { state.members[m.id] = m; });
    window.OTSidebar.setRoomName(room.name);
    window.OTSidebar.renderPresence(presenceList());
    window.OTSidebar.clearChat();
    (room.chatHistory || []).forEach((msg) => {
      window.OTSidebar.appendChat({
        text: msg.text, timestamp: msg.timestamp,
        author: authorOf(msg.memberId), isSelf: isSelf(msg.memberId)
      });
    });
  }
  function onMemberJoined(d) {
    if (!d || !d.member || !d.member.id) return;
    state.members[d.member.id] = d.member;
    window.OTSidebar.renderPresence(presenceList());
    window.OTSidebar.appendSystem({ name: d.member.displayName, rest: "joined" });
  }
  function onMemberLeft(d) {
    if (!d || !d.memberId) return;
    const m = state.members[d.memberId];
    delete state.members[d.memberId];
    delete typing[d.memberId];
    window.OTSidebar.renderPresence(presenceList());
    refreshTyping();
    if (m) window.OTSidebar.appendSystem({ name: m.displayName, rest: "left" });
  }
  function onChat(d) {
    if (!d || !d.message) return;
    const msg = d.message;
    delete typing[msg.memberId];
    refreshTyping();
    window.OTSidebar.appendChat({
      text: msg.text, timestamp: msg.timestamp,
      author: authorOf(msg.memberId), isSelf: isSelf(msg.memberId)
    });
  }

  // --- sidebar -> us ----------------------------------------------------
  function saveProfile(p) {
    if (!self) return;
    self.displayName = p.displayName;
    self.avatarColor = p.avatarColor;
    try {
      localStorage.setItem("ot.profile", JSON.stringify({
        memberId: self.id, displayName: self.displayName, avatarColor: self.avatarColor
      }));
    } catch (e) { /* host page may block localStorage; non-fatal */ }
    if (state.members[self.id]) {
      state.members[self.id].displayName = self.displayName;
      state.members[self.id].avatarColor = self.avatarColor;
    }
    window.OTSidebar.renderPresence(presenceList());
  }

  // --- entry ------------------------------------------------------------
  async function joinRoom(roomId, member) {
    if (joined || !roomId || !member) return;
    joined = true;
    self = member;
    state.roomId = roomId;

    const tokensCss = await fetch(chrome.runtime.getURL("ot/tokens.css")).then((r) => r.text());
    const root = window.OTShadow.getShadowRoot(tokensCss);
    window.OTOverlay.init();

    window.OTSidebar.mount(root, {
      onChat: (text) => send("chat", { text }),
      onReaction: (emoji) => send("reaction-burst", { emoji }),
      onRename: (name) => send("rename-room", { name }),
      onTyping: (t) => send(t ? "typing-start" : "typing-stop", {}),
      onProfileSave: saveProfile
    });
    window.OTSidebar.setCode(roomId);
    window.OTSidebar.setSelfProfile({ displayName: self.displayName, avatarColor: self.avatarColor });
    window.OTSidebar.setConnection("Connecting…");

    connectWs();
  }

  chrome.runtime.onMessage.addListener((msgText) => {
    let msg;
    try { msg = JSON.parse(msgText); } catch (e) { return; }
    if (msg && msg.type === 5003) joinRoom(msg.roomId, msg.member);
  });
})();
