import { createContext, use, useEffect, useState } from 'react'

type Theme = 'dark' | 'light' | 'system'

const THEME_STORAGE_KEY = 'vite-ui-theme'

interface ThemeProviderProps {
  children: React.ReactNode
}

interface ThemeProviderState {
  theme: Theme
  isDark: boolean
  setTheme: (theme: Theme) => void
}

const initialState: ThemeProviderState = {
  theme: 'system',
  isDark: false,
  setTheme: () => null,
}

const ThemeProviderContext = createContext<ThemeProviderState>(initialState)

export function ThemeProvider({ children }: ThemeProviderProps) {
  const [theme, setTheme] = useState<Theme>(() => {
    const stored = localStorage.getItem(THEME_STORAGE_KEY)
    return (stored ?? 'system') as Theme
  })

  const [prefersDark, setPrefersDark] = useState(
    () => window.matchMedia('(prefers-color-scheme: dark)').matches,
  )

  useEffect(() => {
    const mq = window.matchMedia('(prefers-color-scheme: dark)')
    const onChange = (e: MediaQueryListEvent) => {
      setPrefersDark(e.matches)
    }
    mq.addEventListener('change', onChange)
    return () => {
      mq.removeEventListener('change', onChange)
    }
  }, [])

  useEffect(() => {
    const root = window.document.documentElement

    root.classList.remove('light', 'dark')

    if (theme === 'system') {
      root.classList.add(prefersDark ? 'dark' : 'light')
      return
    }

    root.classList.add(theme)
  }, [theme, prefersDark])

  const isDark = theme === 'dark' || (theme === 'system' && prefersDark)

  const value = {
    theme,
    isDark,
    setTheme: (theme: Theme) => {
      localStorage.setItem(THEME_STORAGE_KEY, theme)
      setTheme(theme)
    },
  }

  return <ThemeProviderContext value={value}>{children}</ThemeProviderContext>
}

// eslint-disable-next-line react-refresh/only-export-components
export const useTheme = () => use(ThemeProviderContext)
