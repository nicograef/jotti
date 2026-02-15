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
 * Calls the fetcher on mount (and whenever deps change), manages loading/error state,
 * and exposes a reload function for manual refetching.
 */
export function useFetch<T>(
  fetcher: () => Promise<T>,
  initialData: T,
  deps: unknown[] = [],
): UseFetchResult<T> {
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<T>(initialData)
  const [error, setError] = useState<Error | null>(null)

  // eslint-disable-next-line react-hooks/exhaustive-deps
  const stableFetcher = useCallback(fetcher, deps)

  const fetchData = useCallback(async () => {
    setLoading(true)
    setError(null)

    try {
      const result = await stableFetcher()
      setData(result)
    } catch (err) {
      console.error('Fetch failed:', err)
      setError(err instanceof Error ? err : new Error(String(err)))
    }

    setLoading(false)
  }, [stableFetcher])

  const reload = useCallback(() => {
    void fetchData()
  }, [fetchData])

  useEffect(() => {
    void fetchData()
  }, [fetchData])

  return { loading, data, error, reload, setData }
}
