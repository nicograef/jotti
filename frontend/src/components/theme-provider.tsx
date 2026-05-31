import { createContext, use, useEffect, useState } from 'react'

type Theme = 'dark' | 'light' | 'system'

interface ThemeProviderProps {
  children: React.ReactNode
  defaultTheme?: Theme
  storageKey?: string
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

export function ThemeProvider({
  children,
  defaultTheme = 'system',
  storageKey = 'vite-ui-theme',
  ...props
}: ThemeProviderProps) {
  const [theme, setTheme] = useState<Theme>(() => {
    const stored = localStorage.getItem(storageKey)
    return (stored ?? defaultTheme) as Theme
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
      localStorage.setItem(storageKey, theme)
      setTheme(theme)
    },
  }

  return (
    <ThemeProviderContext {...props} value={value}>
      {children}
    </ThemeProviderContext>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useTheme = () => {
  const context = use(ThemeProviderContext)
  return context
}
