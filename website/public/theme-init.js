// Pre-paint theme init, loaded from public/ via a classic synchronous
// `<script is:inline src>` (no type=module/defer/async): the production CSP
// (`script-src 'self'`) blocks inline scripts, and a module script would run
// only after first paint and flash the wrong theme. Reads Starlight's own store
// (localStorage `starlight-theme`: light/dark, empty = follow system) and
// defines `window.StarlightThemeProvider`, which the docs theme picker expects.
window.StarlightThemeProvider = (() => {
  const storedTheme =
    typeof localStorage !== 'undefined' &&
    localStorage.getItem('starlight-theme')
  const theme =
    storedTheme ||
    (window.matchMedia('(prefers-color-scheme: light)').matches
      ? 'light'
      : 'dark')
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
