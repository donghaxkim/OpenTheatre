// OpenTheatre v2 sidebar — full UI, ported from docs/design/mockup.html.
// A "dumb" renderer: it builds the DOM, emits user-intent events via the
// handlers passed to mount(), and exposes render functions. All room state
// lives in content.js. Exposed as window.OTSidebar.
(function () {
  // 8 avatar colours (hex), mirrored from the tokens.css --p-* palette and
  // the popup's PALETTE. Profiles store hex, so the sidebar works in hex too.
  const PALETTE = ["#cf9a52", "#c8695b", "#b08968", "#8f9f5e",
                   "#5fa08f", "#6b91c4", "#a87fb0", "#cd7f9c"];

  const CSS = `
    .sidebar { position: fixed; top: 0; right: 0; width: 320px; height: 100vh;
      z-index: 2147483000; display: flex; flex-direction: column;
      background: var(--bg); border-left: 1px solid var(--line-2);
      font-family: var(--font), system-ui, sans-serif; color: var(--text);
      transition: transform .42s var(--ease); }
    .sidebar, .sidebar * { box-sizing: border-box; }
    .sidebar.collapsed { transform: translateX(100%); }
    .reopen { position: fixed; right: 14px; bottom: 14px; z-index: 2147483000;
      width: 44px; height: 44px; border-radius: 50%; display: none;
      place-items: center; cursor: pointer; color: var(--text);
      background: var(--surface-2); border: 1px solid var(--line-2);
      box-shadow: 0 8px 22px -10px rgba(0,0,0,.7); }
    .reopen.show { display: grid; }
    .reopen svg { width: 20px; height: 20px; }

    .conn { font-size: 11px; text-align: center; padding: 6px;
      color: var(--text-faint); background: var(--surface);
      border-bottom: 1px solid var(--line); display: none; }
    .conn.show { display: block; }

    .head { position: relative; padding: 15px 14px 13px; border-bottom: 1px solid var(--line); }
    .head-top { display: flex; align-items: center; gap: 8px; }
    .room-name-box { margin: -3px auto -3px -6px; min-width: 0; display: flex;
      align-items: center; gap: 4px; padding: 3px 6px; border-radius: 6px;
      border: 1px solid transparent; cursor: pointer;
      transition: background .15s var(--ease), border-color .15s var(--ease); }
    .room-name-box:hover { background: var(--surface); border-color: var(--line-2); }
    .room-name-box.editing { background: var(--bg); border-color: var(--text-faint); cursor: text; }
    .room-name { min-width: 0; font-size: 14px; font-weight: 600; letter-spacing: -.01em;
      white-space: nowrap; overflow: hidden; text-overflow: ellipsis; outline: none; }
    .room-name.editing { text-overflow: clip; }
    .rename-hint { width: 13px; height: 13px; flex: none; color: var(--text-faint);
      opacity: 0; transform: scale(.7);
      transition: opacity .15s var(--ease), transform .15s var(--ease); }
    .room-name-box:hover .rename-hint { opacity: 1; transform: scale(1); }
    .room-name-box.editing .rename-hint { opacity: 0; }
    .ghost { width: 28px; height: 28px; flex: none; border: none; background: transparent;
      border-radius: 7px; color: var(--text-faint); display: grid; place-items: center;
      cursor: pointer; transition: all .2s var(--ease); }
    .ghost:hover { background: var(--surface-2); color: var(--text); }
    .ghost:active { transform: scale(.94); }
    .ghost.copied { color: var(--text); }
    .ghost svg { width: 16px; height: 16px; }
    #inviteBtn { position: relative; }
    #inviteBtn > svg { transition: opacity .16s var(--ease), transform .16s var(--ease); }
    #inviteBtn.copied > svg { opacity: 0; transform: scale(.6); }
    .copied-flag { position: absolute; top: 50%; right: 0; z-index: 5;
      display: flex; align-items: center; gap: 5px; white-space: nowrap;
      padding: 5px 9px; border-radius: 7px; background: var(--surface-2);
      border: 1px solid var(--line-2); color: var(--text); font-size: 11px;
      font-weight: 500; pointer-events: none;
      opacity: 0; transform: translateY(-50%) scale(.85); transform-origin: right center;
      transition: opacity .2s var(--ease), transform .2s var(--ease); }
    #inviteBtn.copied .copied-flag { opacity: 1; transform: translateY(-50%) scale(1); }
    .copied-flag svg { width: 12px; height: 12px; flex: none; }
    .head-meta { display: flex; align-items: center; gap: 8px; margin-top: 7px; }
    .code-chip { display: inline-grid; align-items: center; justify-items: center;
      font-family: var(--mono), monospace; font-size: 11px; letter-spacing: .14em;
      color: var(--text-dim); background: var(--surface); border: 1px solid var(--line);
      border-radius: 5px; padding: 3px 8px; cursor: pointer;
      transition: color .18s var(--ease), border-color .18s var(--ease), transform .12s var(--ease); }
    .code-chip:hover { color: var(--text); border-color: var(--line-2); }
    .code-chip:active { transform: scale(.96); }
    .code-chip.copied { color: var(--text); border-color: var(--text-faint); }
    .code-chip svg { width: 12px; height: 12px; flex: none; }
    .cc-idle, .cc-done { grid-area: 1 / 1; display: flex; align-items: center; gap: 5px;
      transition: opacity .15s var(--ease); }
    .cc-done { opacity: 0; }
    .code-chip.copied .cc-idle { opacity: 0; }
    .code-chip.copied .cc-done { opacity: 1; }
    .count { font-size: 11px; color: var(--text-faint); }

    .settings { position: absolute; top: calc(100% + 6px); left: 8px; right: 8px;
      z-index: 40; background: var(--surface); border: 1px solid var(--line-2);
      border-radius: 10px; padding: 14px; box-shadow: 0 16px 40px -16px rgba(0,0,0,.7);
      opacity: 0; visibility: hidden; transform: translateY(-6px);
      transition: opacity .18s var(--ease), transform .18s var(--ease), visibility .18s; }
    .settings.open { opacity: 1; visibility: visible; transform: translateY(0); }
    .settings-top { display: flex; align-items: center; justify-content: space-between;
      margin-bottom: 13px; }
    .settings-title { font-size: 13px; font-weight: 600; }
    .settings-label { display: block; font-size: 10px; font-weight: 500;
      letter-spacing: .08em; text-transform: uppercase; color: var(--text-faint);
      margin-bottom: 8px; }
    .name-field { display: flex; align-items: center; gap: 10px; margin-bottom: 15px; }
    .name-field .ava { width: 38px; height: 38px; font-size: 14px; flex: none; }
    .name-field input { flex: 1; min-width: 0; background: var(--bg);
      border: 1px solid var(--line-2); border-radius: 8px; padding: 9px 11px;
      color: var(--text); font-family: var(--font), system-ui, sans-serif;
      font-size: 13px; outline: none; transition: border-color .2s ease; }
    .name-field input:focus { border-color: var(--text-faint); }
    .swatches { display: flex; justify-content: space-between; }
    .swatch { width: 28px; height: 28px; border-radius: 50%; border: none; padding: 0;
      cursor: pointer; transition: transform .15s var(--ease); }
    .swatch:hover { transform: scale(1.12); }
    .swatch:active { transform: scale(.94); }
    .swatch.is-on { box-shadow: 0 0 0 2px var(--surface), 0 0 0 4px var(--text); }

    .presence { display: flex; gap: 6px; padding: 16px 14px; justify-content: center;
      border-bottom: 1px solid var(--line); overflow-x: auto; }
    .person { flex: none; width: 60px; display: flex; flex-direction: column;
      align-items: center; gap: 7px; text-align: center; }
    .ava { width: 44px; height: 44px; border-radius: 50%; display: grid;
      place-items: center; font-size: 16px; font-weight: 600; color: #181613; }
    .person .nm { font-size: 11px; color: var(--text-dim); max-width: 60px;
      overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .person .role { font-size: 9px; letter-spacing: .11em; text-transform: uppercase;
      color: var(--text-faint); }

    .chat { flex: 1; overflow-y: auto; padding: 16px 14px; display: flex;
      flex-direction: column; gap: 15px; }
    .msg { display: flex; gap: 9px; }
    .msg .ava { width: 28px; height: 28px; font-size: 11px; }
    .msg-b { min-width: 0; }
    .msg-h { display: flex; align-items: baseline; gap: 7px; margin-bottom: 2px; }
    .msg-h .who { font-size: 12px; font-weight: 600; }
    .msg-h .t { font-size: 10px; color: var(--text-faint); }
    .msg-x { font-size: 13px; line-height: 1.5; color: var(--text); word-wrap: break-word;
      overflow-wrap: anywhere; }
    .joined { font-size: 11px; color: var(--text-faint); text-align: center; }
    .joined b { color: var(--text-dim); font-weight: 500; }

    .empty { flex: 1; display: none; flex-direction: column; align-items: center;
      justify-content: center; gap: 14px; padding: 30px 26px; text-align: center; }
    .empty.show { display: flex; }
    .empty .ring { width: 52px; height: 52px; border-radius: 50%;
      border: 1.5px dashed var(--line-2); }
    .empty p { font-size: 13px; line-height: 1.55; color: var(--text-dim); }
    .empty .ecode { font-family: var(--mono), monospace; font-size: 13px;
      letter-spacing: .14em; color: var(--text-dim); background: var(--surface);
      border: 1px solid var(--line); border-radius: 5px; padding: 7px 12px; }
    .empty .share { display: flex; align-items: center; gap: 8px; }
    .empty .copy { display: inline-flex; align-items: center; gap: 7px; font-size: 12px;
      font-weight: 500; color: var(--text-dim); background: var(--surface-2);
      border: 1px solid var(--line-2); border-radius: 7px; padding: 8px 12px;
      cursor: pointer; transition: all .2s var(--ease); }
    .empty .copy:hover { color: var(--text); border-color: var(--text-faint); }
    .empty .copy.copied { color: var(--text); border-color: var(--text-faint); }

    .typing { display: none; align-items: center; gap: 7px; padding: 0 14px 8px; }
    .typing.show { display: flex; }
    .typing .d { display: flex; gap: 3px; }
    .typing .d i { width: 4px; height: 4px; border-radius: 50%;
      background: var(--text-faint); animation: ot-bob 1.25s var(--ease) infinite; }
    .typing .d i:nth-child(2) { animation-delay: .15s; }
    .typing .d i:nth-child(3) { animation-delay: .3s; }
    @keyframes ot-bob { 0%,55%,100%{opacity:.3;transform:translateY(0)}
      28%{opacity:.9;transform:translateY(-3px)} }
    .typing span { font-size: 11px; color: var(--text-faint); }

    .reacts { display: flex; gap: 3px; padding: 8px 14px; border-top: 1px solid var(--line); }
    .reacts button { flex: 1; height: 32px; border: 1px solid transparent;
      border-radius: 8px; background: transparent; font-size: 16px; cursor: pointer;
      transition: all .2s var(--ease); }
    .reacts button:hover { background: var(--surface-2); }
    .reacts button:active { transform: scale(.9); }

    .composer { display: flex; align-items: center; gap: 8px; padding: 10px 14px 13px;
      border-top: 1px solid var(--line); }
    .composer .f { flex: 1; display: flex; align-items: center; gap: 6px;
      background: var(--surface); border: 1px solid var(--line-2); border-radius: 9px;
      padding: 0 6px 0 12px; transition: border-color .2s ease; }
    .composer .f:focus-within { border-color: var(--text-faint); }
    .composer input { flex: 1; background: transparent; border: none; outline: none;
      color: var(--text); font-family: var(--font), system-ui, sans-serif;
      font-size: 13px; padding: 9px 0; }
    .composer input::placeholder { color: var(--text-faint); }
    .send { width: 32px; height: 32px; flex: none; border: 1px solid var(--line-2);
      border-radius: 8px; background: var(--surface-2); color: var(--text-dim);
      display: grid; place-items: center; cursor: pointer; transition: all .2s var(--ease); }
    .send:hover { color: var(--text); border-color: var(--text-faint); }
    .send:active { transform: scale(.92); }
    .send svg { width: 15px; height: 15px; }`;

  const HTML = `
    <aside class="sidebar" id="sb">
      <div class="conn" id="conn"></div>
      <div class="head">
        <div class="head-top">
          <div class="room-name-box" id="roomNameBox" title="Double-click to rename">
            <span class="room-name" id="roomName">Room</span>
            <svg class="rename-hint" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 20h9"/><path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4z"/></svg>
          </div>
          <button class="ghost" id="inviteBtn" title="Copy room code">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M19 8v6M22 11h-6"/></svg>
            <span class="copied-flag">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4"><path d="M20 6L9 17l-5-5"/></svg>
              Code copied
            </span>
          </button>
          <button class="ghost" id="gearBtn" title="Settings">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.6 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.6a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9z"/></svg>
          </button>
          <button class="ghost" id="collapseBtn" title="Collapse">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 18l6-6-6-6"/></svg>
          </button>
        </div>
        <div class="head-meta">
          <button class="code-chip" id="copyCode" title="Copy room code">
            <span class="cc-idle" id="ccCode">------</span>
            <span class="cc-done" aria-hidden="true">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4"><path d="M20 6L9 17l-5-5"/></svg>
              Copied
            </span>
          </button>
          <span class="count" id="count"></span>
        </div>
        <div class="settings" id="settings">
          <div class="settings-top">
            <span class="settings-title">Settings</span>
            <button class="ghost" id="settingsClose" title="Close">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 6L6 18M6 6l12 12"/></svg>
            </button>
          </div>
          <label class="settings-label" for="nameInput">Display name</label>
          <div class="name-field">
            <div class="ava" id="profilePreview">?</div>
            <input id="nameInput" maxlength="20" autocomplete="off" />
          </div>
          <span class="settings-label">Color</span>
          <div class="swatches" id="swatches"></div>
        </div>
      </div>
      <div class="presence" id="presence"></div>
      <div class="chat" id="chat"></div>
      <div class="empty" id="empty">
        <div class="ring"></div>
        <p>No one else here yet.<br/>Share the code to pull friends in.</p>
        <div class="share">
          <span class="ecode" id="ecode">------</span>
          <button class="copy" id="copyEmpty">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:13px;height:13px"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
            <span id="copyEmptyLabel">Copy room code</span>
          </button>
        </div>
      </div>
      <div class="typing" id="typing">
        <div class="d"><i></i><i></i><i></i></div>
        <span id="typingText"></span>
      </div>
      <div class="reacts" id="reacts">
        <button data-e="\u{1F602}">\u{1F602}</button>
        <button data-e="❤️">❤️</button>
        <button data-e="\u{1F525}">\u{1F525}</button>
        <button data-e="\u{1F44F}">\u{1F44F}</button>
        <button data-e="\u{1F62D}">\u{1F62D}</button>
      </div>
      <div class="composer">
        <div class="f">
          <input id="composerInput" placeholder="Message the room" maxlength="500" />
        </div>
        <button class="send" id="sendBtn" title="Send">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2"><path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z"/></svg>
        </button>
      </div>
    </aside>
    <button class="reopen" id="reopen" title="Open OpenTheatre">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M15 18l-6-6 6-6"/></svg>
    </button>`;

  let root, handlers = {}, els = {}, code = "", colour = "#cf9a52";
  let typingSent = false, typingTimer = null;

  function initial(name) { return ((name || "").trim()[0] || "?").toUpperCase(); }
  function fmtTime(ts) {
    const d = new Date(ts || Date.now());
    return d.getHours() + ":" + String(d.getMinutes()).padStart(2, "0");
  }
  function $(id) { return root.getElementById ? root.getElementById(id) : root.querySelector("#" + id); }

  function mount(shadowRoot, h) {
    root = shadowRoot;
    handlers = h || {};
    const style = document.createElement("style");
    style.textContent = CSS;
    root.appendChild(style);
    const wrap = document.createElement("div");
    wrap.innerHTML = HTML;
    while (wrap.firstChild) root.appendChild(wrap.firstChild);
    cacheEls();
    buildSwatches();
    wireEvents();
  }

  function cacheEls() {
    ["sb", "conn", "roomName", "roomNameBox", "inviteBtn", "gearBtn", "collapseBtn",
     "copyCode", "ccCode", "count", "settings", "settingsClose", "nameInput",
     "profilePreview", "swatches", "presence", "chat", "empty", "ecode", "copyEmpty",
     "copyEmptyLabel", "typing", "typingText", "reacts", "composerInput", "sendBtn",
     "reopen"].forEach((id) => { els[id] = root.querySelector("#" + id); });
  }

  function buildSwatches() {
    PALETTE.forEach((c, i) => {
      const b = document.createElement("button");
      b.className = "swatch" + (i === 0 ? " is-on" : "");
      b.style.background = c;
      b.dataset.c = c;
      b.addEventListener("click", () => {
        els.swatches.querySelectorAll(".swatch").forEach((s) => s.classList.remove("is-on"));
        b.classList.add("is-on");
        colour = c;
        applySelfColour(c);
      });
      els.swatches.appendChild(b);
    });
  }

  // --- copy feedback (shared by invite button, code chip, empty button) ---
  function flashCopied(btn) {
    if (!btn) return;
    if (navigator.clipboard && code) navigator.clipboard.writeText(code).catch(() => {});
    btn.classList.add("copied");
    clearTimeout(btn._t);
    btn._t = setTimeout(() => btn.classList.remove("copied"), 1600);
  }

  function wireEvents() {
    els.inviteBtn.addEventListener("click", () => flashCopied(els.inviteBtn));
    els.copyCode.addEventListener("click", () => flashCopied(els.copyCode));
    els.copyEmpty.addEventListener("click", () => {
      flashCopied(els.copyEmpty);
      els.copyEmptyLabel.textContent = "Copied";
      setTimeout(() => { els.copyEmptyLabel.textContent = "Copy room code"; }, 1600);
    });

    // settings open/close
    els.gearBtn.addEventListener("click", (e) => {
      e.stopPropagation();
      els.settings.classList.toggle("open");
    });
    els.settingsClose.addEventListener("click", () => closeSettings());
    root.addEventListener("click", (e) => {
      if (els.settings.classList.contains("open") && !els.settings.contains(e.target)
          && e.target !== els.gearBtn && !els.gearBtn.contains(e.target)) {
        closeSettings();
      }
    });
    els.nameInput.addEventListener("input", () => applySelfName(els.nameInput.value));

    // collapse / reopen
    els.collapseBtn.addEventListener("click", () => {
      els.sb.classList.add("collapsed");
      els.reopen.classList.add("show");
    });
    els.reopen.addEventListener("click", () => {
      els.sb.classList.remove("collapsed");
      els.reopen.classList.remove("show");
    });

    // reactions
    els.reacts.querySelectorAll("button[data-e]").forEach((b) => {
      b.addEventListener("click", () => {
        if (handlers.onReaction) handlers.onReaction(b.dataset.e);
      });
    });

    // composer
    els.sendBtn.addEventListener("click", sendChat);
    els.composerInput.addEventListener("keydown", (e) => {
      if (e.key === "Enter") { e.preventDefault(); sendChat(); }
    });
    els.composerInput.addEventListener("input", onComposerInput);

    wireRename();
  }

  function closeSettings() {
    els.settings.classList.remove("open");
    if (handlers.onProfileSave) {
      handlers.onProfileSave({
        displayName: els.nameInput.value.trim() || "you",
        avatarColor: colour
      });
    }
  }

  function sendChat() {
    const text = els.composerInput.value.trim();
    if (!text) return;
    els.composerInput.value = "";
    stopTyping();
    if (handlers.onChat) handlers.onChat(text);
  }

  function onComposerInput() {
    if (!els.composerInput.value) { stopTyping(); return; }
    if (!typingSent) {
      typingSent = true;
      if (handlers.onTyping) handlers.onTyping(true);
    }
    clearTimeout(typingTimer);
    typingTimer = setTimeout(stopTyping, 3000);
  }
  function stopTyping() {
    clearTimeout(typingTimer);
    if (typingSent) {
      typingSent = false;
      if (handlers.onTyping) handlers.onTyping(false);
    }
  }

  function wireRename() {
    const box = els.roomNameBox, nm = els.roomName;
    let saved = "";
    box.addEventListener("dblclick", () => {
      if (nm.isContentEditable) return;
      saved = nm.textContent.trim();
      box.classList.add("editing");
      nm.classList.add("editing");
      nm.setAttribute("contenteditable", "true");
      nm.focus();
      const range = document.createRange();
      range.selectNodeContents(nm);
      const sel = window.getSelection();
      sel.removeAllRanges();
      sel.addRange(range);
    });
    nm.addEventListener("keydown", (e) => {
      if (e.key === "Enter") { e.preventDefault(); nm.blur(); }
      else if (e.key === "Escape") { e.preventDefault(); nm.textContent = saved; nm.blur(); }
    });
    nm.addEventListener("blur", () => {
      if (!box.classList.contains("editing")) return;
      nm.removeAttribute("contenteditable");
      box.classList.remove("editing");
      nm.classList.remove("editing");
      const next = nm.textContent.trim();
      if (next && next !== saved) {
        if (handlers.onRename) handlers.onRename(next);
      } else {
        nm.textContent = saved;
      }
    });
  }

  // --- live "you" updates from the settings panel ---
  function applySelfColour(hex) {
    els.profilePreview.style.background = hex;
    root.querySelectorAll('[data-you="ava"]').forEach((a) => { a.style.background = hex; });
    root.querySelectorAll('.who[data-you="name"]').forEach((w) => { w.style.color = hex; });
  }
  function applySelfName(raw) {
    const name = (raw || "").trim();
    els.profilePreview.textContent = initial(name);
    root.querySelectorAll('[data-you="ava"]').forEach((a) => { a.textContent = initial(name); });
    root.querySelectorAll('[data-you="name"]').forEach((n) => { n.textContent = name || "you"; });
  }

  // --- render API (called by content.js) ---
  function setSelfProfile(p) {
    els.nameInput.value = p.displayName || "you";
    colour = p.avatarColor || "#cf9a52";
    els.swatches.querySelectorAll(".swatch").forEach((s) => {
      s.classList.toggle("is-on", s.dataset.c === colour);
    });
    els.profilePreview.style.background = colour;
    els.profilePreview.textContent = initial(p.displayName);
  }

  function setRoomName(name) { els.roomName.textContent = name || "Room"; }

  function setCode(c) {
    code = c || "";
    els.ccCode.textContent = code || "------";
    els.ecode.textContent = code || "------";
  }

  function setConnection(text) {
    els.conn.textContent = text || "";
    els.conn.classList.toggle("show", !!text);
  }

  // members: [{id, displayName, avatarColor, isHost, isSelf}]
  function renderPresence(members) {
    els.presence.innerHTML = "";
    members.forEach((m) => {
      const person = document.createElement("div");
      person.className = "person";
      const ava = document.createElement("div");
      ava.className = "ava";
      ava.style.background = m.avatarColor || "var(--p-denim)";
      ava.textContent = initial(m.displayName);
      if (m.isSelf) ava.setAttribute("data-you", "ava");
      const nm = document.createElement("span");
      nm.className = "nm";
      nm.textContent = m.isSelf ? "you" : m.displayName;
      person.appendChild(ava);
      person.appendChild(nm);
      if (m.isHost) {
        const role = document.createElement("span");
        role.className = "role";
        role.textContent = "host";
        person.appendChild(role);
      }
      els.presence.appendChild(person);
    });
    els.count.textContent = members.length <= 1 ? "just you" : members.length + " here";
    showEmpty(members.length <= 1);
  }

  function showEmpty(isEmpty) {
    els.empty.classList.toggle("show", isEmpty);
    els.chat.style.display = isEmpty ? "none" : "flex";
  }

  // msg: {text, timestamp, author:{displayName,avatarColor}, isSelf}
  function appendChat(msg) {
    const el = document.createElement("div");
    el.className = "msg";
    const ava = document.createElement("div");
    ava.className = "ava";
    ava.style.background = msg.author.avatarColor || "var(--p-denim)";
    ava.textContent = initial(msg.author.displayName);
    if (msg.isSelf) ava.setAttribute("data-you", "ava");
    const body = document.createElement("div");
    body.className = "msg-b";
    const head = document.createElement("div");
    head.className = "msg-h";
    const who = document.createElement("span");
    who.className = "who";
    who.textContent = msg.isSelf ? "you" : msg.author.displayName;
    who.style.color = msg.author.avatarColor || "var(--text)";
    if (msg.isSelf) who.setAttribute("data-you", "name");
    const t = document.createElement("span");
    t.className = "t";
    t.textContent = fmtTime(msg.timestamp);
    head.appendChild(who);
    head.appendChild(t);
    const x = document.createElement("div");
    x.className = "msg-x";
    x.textContent = msg.text; // textContent — never innerHTML for user text
    body.appendChild(head);
    body.appendChild(x);
    el.appendChild(ava);
    el.appendChild(body);
    els.chat.appendChild(el);
    els.chat.scrollTop = els.chat.scrollHeight;
  }

  function appendSystem(text) {
    const el = document.createElement("div");
    el.className = "joined";
    const b = document.createElement("b");
    b.textContent = text.name;
    el.appendChild(b);
    el.appendChild(document.createTextNode(" " + text.rest));
    els.chat.appendChild(el);
    els.chat.scrollTop = els.chat.scrollHeight;
  }

  function clearChat() { els.chat.innerHTML = ""; }

  function setTyping(text) {
    els.typingText.textContent = text || "";
    els.typing.classList.toggle("show", !!text);
  }

  const api = {
    mount, setSelfProfile, setRoomName, setCode, setConnection,
    renderPresence, appendChat, appendSystem, clearChat, setTyping, showEmpty
  };
  if (typeof module !== "undefined" && module.exports) module.exports = api;
  if (typeof window !== "undefined") window.OTSidebar = api;
})();
