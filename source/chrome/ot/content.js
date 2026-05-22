// OpenTheatre v2 content script. Waits for the background's 5003 "joinRoom"
// message, injects the sidebar shadow UI, opens the WS, renders room-state.
// Runs in the ISOLATED content-script world, sharing window.OT* with the
// other ot/*.js content scripts.
(function () {
  const SERVER_WS = "ws://localhost:5001"; // mirror of ot/config.js OT_SERVER_WS

  let injected = false;

  async function joinRoom(roomId, member) {
    if (injected) return;
    injected = true;

    // load design tokens, then build the shadow UI
    const tokensUrl = chrome.runtime.getURL("ot/tokens.css");
    const tokensCss = await fetch(tokensUrl).then((r) => r.text());
    const root = window.OTShadow.getShadowRoot(tokensCss);
    window.OTSidebar.renderSidebar(root);

    const ws = new window.OTWsClient.OTWsClient(SERVER_WS, roomId);
    ws.on("room-state", (state) => window.OTSidebar.renderRoomState(root, state));
    // member-joined / member-left drive incremental strip updates in a later
    // slice; the walking skeleton renders presence from the room-state snapshot.
    ws.connect(member);
  }

  chrome.runtime.onMessage.addListener((msgText) => {
    let msg;
    try { msg = JSON.parse(msgText); } catch (e) { return; }
    if (msg && msg.type === 5003) joinRoom(msg.roomId, msg.member);
  });
})();
