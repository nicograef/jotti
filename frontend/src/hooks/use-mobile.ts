import * as React from "react"

const MOBILE_BREAKPOINT = 768
const MOBILE_MAX = MOBILE_BREAKPOINT - 1

const COMPACT_BREAKPOINT = 1024
const COMPACT_MAX = COMPACT_BREAKPOINT - 1

export function useIsMobile() {
  const [isMobile, setIsMobile] = React.useState<boolean | undefined>(undefined)

  React.useEffect(() => {
    const mql = window.matchMedia(`(max-width: ${String(MOBILE_MAX)}px)`)
    const onChange = () => {
      setIsMobile(window.innerWidth < MOBILE_BREAKPOINT)
    }
    mql.addEventListener("change", onChange)
    setIsMobile(window.innerWidth < MOBILE_BREAKPOINT)
    return () => { mql.removeEventListener("change", onChange); }
  }, [])

  return !!isMobile
}

export function useIsCompact() {
  const [isCompact, setIsCompact] = React.useState<boolean | undefined>(undefined)

  React.useEffect(() => {
    const mql = window.matchMedia(`(max-width: ${String(COMPACT_MAX)}px)`)
    const onChange = () => {
      setIsCompact(window.innerWidth < COMPACT_BREAKPOINT)
    }
    mql.addEventListener("change", onChange)
    setIsCompact(window.innerWidth < COMPACT_BREAKPOINT)
    return () => { mql.removeEventListener("change", onChange); }
  }, [])

  return !!isCompact
}
