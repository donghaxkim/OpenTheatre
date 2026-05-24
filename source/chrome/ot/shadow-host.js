// Creates (once) a shadow-DOM host on document.body for OpenTheatre UI.
// Idempotent: repeated calls return the same shadow root.
// Runs as a content script (ISOLATED world); exposes window.OTShadow.
(function () {
  const HOST_ID = "opentheatre-root";

  function getShadowRoot(tokensCssText) {
    let host = document.getElementById(HOST_ID);
    if (host) return host.shadowRoot;

    host = document.createElement("div");
    host.id = HOST_ID;
    // The host element itself carries no layout; the sidebar inside is fixed-positioned.
    host.style.cssText = "all: initial;";
    document.body.appendChild(host);

    const root = host.attachShadow({ mode: "open" });
    const style = document.createElement("style");
    style.textContent = tokensCssText;
    root.appendChild(style);
    return root;
  }

  function removeShadowHost() {
    const host = document.getElementById(HOST_ID);
    if (host) host.remove();
  }

  if (typeof module !== "undefined" && module.exports) {
    module.exports = { getShadowRoot, removeShadowHost };
  }
  if (typeof window !== "undefined") {
    window.OTShadow = { getShadowRoot, removeShadowHost };
  }
})();
