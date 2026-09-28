import { useEffect, useRef } from 'react'

/**
 * `true` exactly on the first render where `bereit` holds, then `false`, so the list entry animates on first build, never on refetch.
 * The flag lives in a ref because an extra render would cut the fresh animation; reading it during render (`react-hooks/refs`) is intended.
 */
export function useErstAufbau(bereit: boolean): boolean {
  const aufgebautRef = useRef(false)
  useEffect(() => {
    if (bereit) {
      aufgebautRef.current = true
    }
  }, [bereit])
  // eslint-disable-next-line react-hooks/refs
  return bereit && !aufgebautRef.current
}
