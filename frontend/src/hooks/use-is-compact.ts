import React from 'react'

const COMPACT_BREAKPOINT = 1024
const COMPACT_MAX = COMPACT_BREAKPOINT - 1

export function useIsCompact() {
  const [isCompact, setIsCompact] = React.useState(
    () => window.innerWidth < COMPACT_BREAKPOINT,
  )

  React.useEffect(() => {
    const mql = window.matchMedia(`(max-width: ${String(COMPACT_MAX)}px)`)
    const onChange = () => {
      setIsCompact(window.innerWidth < COMPACT_BREAKPOINT)
    }
    mql.addEventListener('change', onChange)
    return () => {
      mql.removeEventListener('change', onChange)
    }
  }, [])

  return isCompact
}
