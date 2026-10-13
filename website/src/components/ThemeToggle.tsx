import { useSyncExternalStore } from 'react'
import { Moon, Sun } from 'lucide-react'

// Schreibt in denselben Speicher wie der Doku-Schalter (localStorage
// `starlight-theme`) und setzt `data-theme` auf <html> — identisch zu
// public/theme-init.js, damit Landing und Doku synchron bleiben.

type Theme = 'light' | 'dark'

// Folgt `data-theme` auf <html>, egal welcher Schalter es setzt.
function subscribeTheme(onChange: () => void): () => void {
  const observer = new MutationObserver(onChange)
  observer.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['data-theme'],
  })
  return () => {
    observer.disconnect()
  }
}

function readTheme(): Theme {
  return document.documentElement.dataset.theme === 'light' ? 'light' : 'dark'
}

export default function ThemeToggle() {
  // Vor der Hydration ist das Theme unbekannt; `null` rendert einen Platzhalter
  // gleicher Größe — kein Hydration-Mismatch, kein Layout-Sprung.
  const theme = useSyncExternalStore<Theme | null>(
    subscribeTheme,
    readTheme,
    () => null,
  )

  function toggle() {
    const next: Theme = (theme ?? readTheme()) === 'dark' ? 'light' : 'dark'
    document.documentElement.dataset.theme = next
    try {
      localStorage.setItem('starlight-theme', next)
    } catch {
      // localStorage kann blockiert sein (Privatmodus) — Umschaltung wirkt
      // trotzdem für die aktuelle Sitzung.
    }
    window.StarlightThemeProvider.updatePickers(next)
  }

  const label =
    theme === 'dark'
      ? 'Zu hellem Design wechseln'
      : 'Zu dunklem Design wechseln'

  return (
    <button
      type="button"
      onClick={toggle}
      aria-label={theme === null ? 'Design wechseln' : label}
      className="flex h-10 w-10 items-center justify-center rounded-xl border border-card-border bg-card text-foreground transition-colors hover:border-brand hover:text-brand"
    >
      {theme === 'dark' ? (
        <Sun size={19} aria-hidden="true" />
      ) : theme === 'light' ? (
        <Moon size={19} aria-hidden="true" />
      ) : (
        <span className="h-[19px] w-[19px]" aria-hidden="true" />
      )}
    </button>
  )
}
