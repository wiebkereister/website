// Mobile navigation toggle and the two-click map. No cookies, no tracking.
document.addEventListener("DOMContentLoaded", () => {
  const toggle = document.querySelector(".nav-toggle");
  const nav = document.getElementById("hauptnavigation");
  if (toggle && nav) {
    toggle.addEventListener("click", () => {
      const open = toggle.getAttribute("aria-expanded") !== "true";
      toggle.setAttribute("aria-expanded", String(open));
      nav.classList.toggle("open", open);
    });
  }

  document.querySelectorAll(".map[data-map-src]").forEach((map) => {
    const consent = map.querySelector(".map-consent");
    const frame = map.querySelector(".map-frame");
    const controls = map.querySelector("[data-map-controls]");
    map.querySelector("[data-map-show]").addEventListener("click", () => {
      const iframe = document.createElement("iframe");
      iframe.src = map.dataset.mapSrc;
      iframe.title = map.dataset.mapTitle;
      iframe.loading = "lazy";
      iframe.referrerPolicy = "no-referrer-when-downgrade";
      frame.replaceChildren(iframe);
      frame.hidden = false;
      controls.hidden = false;
      consent.hidden = true;
      controls.querySelector("button").focus();
    });
    map.querySelector("[data-map-hide]").addEventListener("click", () => {
      frame.replaceChildren();
      frame.hidden = true;
      controls.hidden = true;
      consent.hidden = false;
      map.querySelector("[data-map-show]").focus();
    });
  });
});
