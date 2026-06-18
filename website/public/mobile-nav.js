// Mobile-Navigation umschalten.
//
// Bewusst eine statische Datei in public/ (kein gebuendeltes Astro-`<script>`):
// Die Produktiv-CSP erlaubt nur `script-src 'self'` ohne `'unsafe-inline'`, also
// werden Inline-Skripte vom Browser blockiert. Astro inlinet kleine `<script>`
// jedoch automatisch; eine eigene Datei wird dagegen verlaesslich als externe,
// gleichnamige Ressource ausgeliefert und von der CSP erlaubt.
const toggle = document.getElementById('menu-toggle')
const nav = document.getElementById('mobile-nav')
const iconOpen = toggle?.querySelector('.menu-icon-open')
const iconClose = toggle?.querySelector('.menu-icon-close')

if (toggle && nav) {
  const setOpen = (open) => {
    toggle.setAttribute('aria-expanded', String(open))
    toggle.setAttribute('aria-label', open ? 'Menü schließen' : 'Menü öffnen')
    nav.classList.toggle('hidden', !open)
    nav.classList.toggle('flex', open)
    iconOpen?.classList.toggle('hidden', open)
    iconClose?.classList.toggle('hidden', !open)
  }

  toggle.addEventListener('click', () => {
    setOpen(toggle.getAttribute('aria-expanded') !== 'true')
  })

  nav.querySelectorAll('a').forEach((link) => {
    link.addEventListener('click', () => setOpen(false))
  })

  document.addEventListener('keydown', (event) => {
    if (
      event.key === 'Escape' &&
      toggle.getAttribute('aria-expanded') === 'true'
    ) {
      setOpen(false)
      toggle.focus()
    }
  })
}
