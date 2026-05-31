import {
  type Dispatch,
  type SetStateAction,
  useCallback,
  useEffect,
  useState,
} from 'react'

interface UseFetchResult<T> {
  loading: boolean
  data: T
  error: Error | null
  reload: () => void
  setData: Dispatch<SetStateAction<T>>
}

/**
 * Generic data-fetching hook.
 * Calls the fetcher on mount (and whenever fetcher changes), manages loading/error state,
 * and exposes a reload function for manual refetching.
 */
export function useFetch<T>(
  fetcher: () => Promise<T>,
  initialData: T,
): UseFetchResult<T> {
  const [loading, setLoading] = useState(true)
  const [data, setData] = useState<T>(initialData)
  const [error, setError] = useState<Error | null>(null)
  const [trigger, setTrigger] = useState(0)

  const reload = useCallback(() => {
    setLoading(true)
    setError(null)
    setTrigger((t) => t + 1)
  }, [])

  useEffect(() => {
    let cancelled = false
    fetcher()
      .then((result) => {
        if (!cancelled) {
          setData(result)
          setError(null)
          setLoading(false)
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          console.error('Fetch failed:', err)
          setError(err instanceof Error ? err : new Error(String(err)))
          setLoading(false)
        }
      })
    return () => {
      cancelled = true
    }
  }, [fetcher, trigger])

  return { loading, data, error, reload, setData }
}
