import React from "react"

const MOBILE_BREAKPOINT = 768
const MOBILE_MAX = MOBILE_BREAKPOINT - 1

export function useIsMobile() {
  const [isMobile, setIsMobile] = React.useState(
    () => window.innerWidth < MOBILE_BREAKPOINT,
  )

  React.useEffect(() => {
    const mql = window.matchMedia(`(max-width: ${String(MOBILE_MAX)}px)`)
    const onChange = () => { setIsMobile(window.innerWidth < MOBILE_BREAKPOINT); }
    mql.addEventListener('change', onChange)
    return () => { mql.removeEventListener('change', onChange); }
  }, [])

  return isMobile
}
