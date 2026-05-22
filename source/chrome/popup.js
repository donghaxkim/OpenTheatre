(function () {
  // The 8 v2 avatar colours, mirrored from ot/tokens.css --p-* palette.
  const PALETTE = ["#cf9a52", "#c8695b", "#b08968", "#8f9f5e", "#5fa08f", "#6b91c4", "#a87fb0", "#cd7f9c"];
  const PROFILE_KEY = "ot.profile";

  function generateMemberId() {
    return "m_" + Math.random().toString(36).slice(2) + Date.now().toString(36);
  }
  function loadProfile() {
    try {
      const p = JSON.parse(localStorage.getItem(PROFILE_KEY));
      if (p && p.memberId && p.displayName && p.avatarColor) return p;
    } catch (e) { /* ignore */ }
    return null;
  }
  function initial(name) { return (name.trim()[0] || "?").toUpperCase(); }
  function send(msg) {
    return new Promise((res) => chrome.runtime.sendMessage(JSON.stringify(msg), res));
  }

  const $ = (id) => document.getElementById(id);
  let chosenColor = PALETTE[0];

  // --- first-run setup ---
  function renderSetup() {
    $("setup").hidden = false;
    const sw = $("swatches");
    PALETTE.forEach((c, i) => {
      const b = document.createElement("button");
      b.className = "sw" + (i === 0 ? " on" : "");
      b.style.background = c;
      b.onclick = () => {
        sw.querySelectorAll(".sw").forEach((x) => x.classList.remove("on"));
        b.classList.add("on");
        chosenColor = c;
        $("setupAva").style.background = c;
      };
      sw.appendChild(b);
    });
    $("setupAva").style.background = chosenColor;
    $("setupName").addEventListener("input", () => {
      $("setupAva").textContent = initial($("setupName").value);
    });
    $("saveProfile").onclick = () => {
      const name = $("setupName").value.trim();
      if (!name) return;
      const profile = { memberId: generateMemberId(), displayName: name, avatarColor: chosenColor };
      localStorage.setItem(PROFILE_KEY, JSON.stringify(profile));
      $("setup").hidden = true;
      renderMain(profile);
    };
  }

  // --- create / join ---
  function renderMain(profile) {
    $("main").hidden = false;
    $("meAva").style.background = profile.avatarColor;
    $("meAva").textContent = initial(profile.displayName);
    $("meName").textContent = profile.displayName;

    $("createBtn").onclick = async () => {
      $("err").textContent = "";
      const resp = await send({ type: 5001, member: {
        id: profile.memberId, displayName: profile.displayName, avatarColor: profile.avatarColor
      } });
      if (!resp || resp.error) { $("err").textContent = "Could not create a room."; return; }
      await send({ type: 5002, roomId: resp.roomId, member: {
        id: profile.memberId, displayName: profile.displayName, avatarColor: profile.avatarColor
      } });
      window.close();
    };

    $("joinBtn").onclick = async () => {
      $("err").textContent = "";
      const code = $("codeInput").value.trim().toUpperCase();
      if (code.length !== 6) { $("err").textContent = "Enter a 6-character code."; return; }
      await send({ type: 5002, roomId: code, member: {
        id: profile.memberId, displayName: profile.displayName, avatarColor: profile.avatarColor
      } });
      window.close();
    };
  }

  const profile = loadProfile();
  if (profile) renderMain(profile); else renderSetup();
})();
