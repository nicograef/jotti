import { useEffect, useState } from 'react'

// The app's only desktop threshold: below lg (1024 px) mobile/tablet with drawers, from lg sidebar and two columns.
// See docs/decisions.md D07.
const MOBILE_BREAKPOINT = 1024
const MOBILE_MAX = MOBILE_BREAKPOINT - 1

export function useIsMobile(): boolean {
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
