// Pre-paint theme initialisation — shared by the landing (loaded from
// Landing.astro) and the docs (loaded from the ThemeProvider override).
//
// Deliberately a real file in public/ referenced via a classic, synchronous
// `<script is:inline src="/theme-init.js">` (no type=module/defer/async): the
// production CSP (`script-src 'self'`) blocks inline scripts, and a module
// script would run only after first paint and flash the wrong theme.
//
// Reads Starlight's own store (localStorage key `starlight-theme`, values
// light/dark or empty/absent = follow system) and sets `data-theme` on <html>
// before anything renders. Also defines `window.StarlightThemeProvider` so the
// docs theme picker keeps working — same key, same semantics as Starlight's
// default ThemeProvider (which we override for CSP), so the landing and the docs
// switch stay in sync.
window.StarlightThemeProvider = (() => {
  const storedTheme =
    typeof localStorage !== 'undefined' && localStorage.getItem('starlight-theme')
  const theme =
    storedTheme ||
    (window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark')
  document.documentElement.dataset.theme = theme === 'light' ? 'light' : 'dark'
  return {
    updatePickers(theme = storedTheme || 'auto') {
      document.querySelectorAll('starlight-theme-select').forEach((picker) => {
        const select = picker.querySelector('select')
        if (select) select.value = theme
        const tmpl = document.querySelector('#theme-icons')
        const newIcon = tmpl && tmpl.content.querySelector('.' + theme)
        if (newIcon) {
          const oldIcon = picker.querySelector('svg.label-icon')
          if (oldIcon) {
            oldIcon.replaceChildren(...newIcon.cloneNode(true).childNodes)
          }
        }
      })
    },
  }
})()
