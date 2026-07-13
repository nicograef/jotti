import { useEffect, useState } from 'react'
import { Moon, Sun } from 'lucide-react'

// Hell/Dunkel-Umschalter der Landing. Schreibt in denselben Speicher wie der
// Doku-Schalter (Starlight-Key `starlight-theme`, Werte `light`/`dark`) und
// setzt `data-theme` auf <html> — identisch zur Pre-Paint-Logik in
// public/theme-init.js. Dadurch bleiben Landing und Doku über Navigation hinweg
// konsistent. `window.StarlightThemeProvider.updatePickers` hält den
// Doku-Select synchron, wo er existiert.

type Theme = 'light' | 'dark'

declare global {
  interface Window {
    StarlightThemeProvider?: { updatePickers?: (theme: string) => void }
  }
}

function readTheme(): Theme {
  return document.documentElement.dataset.theme === 'light' ? 'light' : 'dark'
}

export default function ThemeToggle() {
  // Vor der Hydration ist das Theme (aus localStorage/Systempräferenz) im
  // Server-HTML nicht bekannt; `null` rendert einen Platzhalter gleicher Größe,
  // sodass es weder Hydration-Mismatch noch Layout-Sprung gibt.
  const [theme, setTheme] = useState<Theme | null>(null)

  useEffect(() => {
    setTheme(readTheme())
  }, [])

  function toggle() {
    const next: Theme = (theme ?? readTheme()) === 'dark' ? 'light' : 'dark'
    document.documentElement.dataset.theme = next
    try {
      localStorage.setItem('starlight-theme', next)
    } catch {
      // localStorage kann blockiert sein (Privatmodus) — Umschaltung wirkt
      // trotzdem für die aktuelle Sitzung.
    }
    window.StarlightThemeProvider?.updatePickers?.(next)
    setTheme(next)
  }

  const label =
    theme === 'dark' ? 'Zu hellem Design wechseln' : 'Zu dunklem Design wechseln'

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
