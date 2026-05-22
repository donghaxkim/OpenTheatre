// Floating reaction overlay — emoji rise over the video when anyone in the
// room reacts. A separate fixed layer, always pointer-events:none so it can
// never block the host page. Exposed as window.OTOverlay.
(function () {
  let layer = null;

  function ensureKeyframes() {
    if (document.getElementById("ot-overlay-kf")) return;
    const s = document.createElement("style");
    s.id = "ot-overlay-kf";
    s.textContent =
      "@keyframes ot-rise{" +
      "0%{opacity:0;transform:translateY(16px) scale(.6)}" +
      "12%{opacity:1;transform:translateY(0) scale(1)}" +
      "100%{opacity:0;transform:translateY(-62vh) scale(.9) translateX(var(--ot-drift,0))}}";
    document.head.appendChild(s);
  }

  function init() {
    if (layer) return;
    ensureKeyframes();
    layer = document.createElement("div");
    layer.id = "opentheatre-overlay";
    layer.style.cssText =
      "position:fixed;top:0;left:0;right:0;bottom:0;" +
      "pointer-events:none;overflow:hidden;z-index:2147482000;";
    document.body.appendChild(layer);
  }

  // Spawn a short burst of one emoji rising from the bottom with drift.
  function burst(emoji) {
    if (!layer || !emoji) return;
    for (let i = 0; i < 6; i++) {
      const r = document.createElement("div");
      r.textContent = emoji;
      r.style.cssText =
        "position:absolute;bottom:6%;left:" + (14 + Math.random() * 68) + "%;" +
        "font-size:" + (22 + Math.random() * 14) + "px;will-change:transform,opacity;" +
        "animation:ot-rise 3.6s cubic-bezier(.32,.72,0,1) forwards;" +
        "animation-delay:" + (Math.random() * 0.45) + "s;";
      r.style.setProperty("--ot-drift", (Math.random() * 60 - 30) + "px");
      layer.appendChild(r);
      setTimeout(() => r.remove(), 4400);
    }
  }

  function remove() {
    if (layer) { layer.remove(); layer = null; }
  }

  if (typeof window !== "undefined") window.OTOverlay = { init, burst, remove };
})();
