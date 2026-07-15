import { useEffect, useState } from 'react'

// Einheitliche „Desktop"-Schwelle der App: unter lg (1024px) gilt als
// mobil/Tablet — Drawer-Navigation, einspaltig, Bottom-Sheets; ab lg
// persistente Sidebar und zweispaltige Inhalte. Deckt sich mit dem
// Content-Zweispalt-Breakpoint und dem Service-Split (ADR 07).
const MOBILE_BREAKPOINT = 1024
const MOBILE_MAX = MOBILE_BREAKPOINT - 1

export function useIsMobile() {
  const [isMobile, setIsMobile] = useState(
    () => window.innerWidth < MOBILE_BREAKPOINT,
  )

  useEffect(() => {
    const mql = window.matchMedia(`(max-width: ${String(MOBILE_MAX)}px)`)
    const onChange = () => {
      setIsMobile(window.innerWidth < MOBILE_BREAKPOINT)
    }
    mql.addEventListener('change', onChange)
    return () => {
      mql.removeEventListener('change', onChange)
    }
  }, [])

  return isMobile
}
