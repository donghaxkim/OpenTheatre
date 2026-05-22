// Renders the OpenTheatre sidebar shell into a shadow root.
// Markup + classes mirror the .sidebar section of docs/design/mockup.html.
// Exposed as window.OTSidebar.
(function () {
  function renderSidebar(root) {
    const aside = document.createElement("aside");
    aside.className = "ot-sidebar";
    aside.innerHTML = `
      <div class="head">
        <div class="head-top"><span class="room-name" id="otRoomName">…</span></div>
        <div class="head-meta"><span class="code" id="otCode"></span>
          <span class="count" id="otCount"></span></div>
      </div>
      <div class="presence" id="otPresence"></div>
      <div class="chat" id="otChat"></div>`;
    // Sidebar layout CSS — ported from mockup.html .sidebar/.head/.presence
    // rules, plus a fixed-position frame so it sits on the host page's edge.
    const style = document.createElement("style");
    style.textContent = `
      .ot-sidebar { position: fixed; top: 0; right: 0; width: 320px; height: 100vh;
        z-index: 2147483000; display: flex; flex-direction: column;
        background: var(--bg); border-left: 1px solid var(--line-2);
        font-family: var(--font), system-ui, sans-serif; color: var(--text);
        box-sizing: border-box; }
      .ot-sidebar * { box-sizing: border-box; }
      .ot-sidebar .head { padding: 15px 14px 13px; border-bottom: 1px solid var(--line); }
      .ot-sidebar .room-name { font-size: 14px; font-weight: 600; }
      .ot-sidebar .head-meta { display: flex; gap: 8px; margin-top: 7px; align-items: center; }
      .ot-sidebar .code { font-family: var(--mono), monospace; font-size: 11px;
        letter-spacing: .14em; color: var(--text-dim); background: var(--surface);
        border: 1px solid var(--line); border-radius: 5px; padding: 2px 7px; }
      .ot-sidebar .count { font-size: 11px; color: var(--text-faint); }
      .ot-sidebar .presence { display: flex; gap: 6px; padding: 16px 14px;
        justify-content: center; border-bottom: 1px solid var(--line); }
      .ot-sidebar .person { display: flex; flex-direction: column; align-items: center;
        gap: 7px; width: 60px; }
      .ot-sidebar .ava { width: 44px; height: 44px; border-radius: 50%; display: grid;
        place-items: center; font-size: 16px; font-weight: 600; color: #181613; }
      .ot-sidebar .nm { font-size: 11px; color: var(--text-dim); }
      .ot-sidebar .chat { flex: 1; }`;
    root.appendChild(style);
    root.appendChild(aside);
    return aside;
  }

  // Fills the header + member strip from a room-state snapshot.
  function renderRoomState(root, state) {
    root.querySelector("#otRoomName").textContent = state.name || "Room";
    root.querySelector("#otCode").textContent = state.id || "";
    const members = state.members || [];
    root.querySelector("#otCount").textContent =
      members.length === 1 ? "just you" : members.length + " here";
    const strip = root.querySelector("#otPresence");
    strip.innerHTML = "";
    members.forEach((m) => {
      const el = document.createElement("div");
      el.className = "person";
      const ava = document.createElement("div");
      ava.className = "ava";
      ava.style.background = m.avatarColor || "#6b91c4";
      ava.textContent = ((m.displayName || "").trim()[0] || "?").toUpperCase();
      const nm = document.createElement("span");
      nm.className = "nm";
      nm.textContent = m.displayName || "";
      el.appendChild(ava);
      el.appendChild(nm);
      strip.appendChild(el);
    });
  }

  const api = { renderSidebar, renderRoomState };
  if (typeof module !== "undefined" && module.exports) module.exports = api;
  if (typeof window !== "undefined") window.OTSidebar = api;
})();
