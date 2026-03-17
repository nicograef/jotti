// Mobile menu toggle
const toggle = document.querySelector(".mobile-menu-toggle");
const overlay = document.querySelector(".mobile-nav-overlay");

if (toggle && overlay) {
  toggle.addEventListener("click", () => {
    const open = toggle.getAttribute("aria-expanded") === "true";
    toggle.setAttribute("aria-expanded", String(!open));
    overlay.setAttribute("aria-hidden", String(open));
    document.body.classList.toggle("nav-open", !open);
  });

  overlay.querySelectorAll("a").forEach((link) => {
    link.addEventListener("click", () => {
      toggle.setAttribute("aria-expanded", "false");
      overlay.setAttribute("aria-hidden", "true");
      document.body.classList.remove("nav-open");
    });
  });
}
