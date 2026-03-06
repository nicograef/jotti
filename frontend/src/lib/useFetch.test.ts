import { renderHook, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { useFetch } from './useFetch'

describe('useFetch', () => {
  it('starts in loading state then transitions to success', async () => {
    const data = { name: 'Tisch 1' }
    const fetcher = vi.fn().mockResolvedValue(data)

    const { result } = renderHook(() => useFetch(fetcher, null))

    // Initially loading
    expect(result.current.loading).toBe(true)
    expect(result.current.data).toBeNull()
    expect(result.current.error).toBeNull()

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.data).toEqual(data)
    expect(result.current.error).toBeNull()
    expect(fetcher).toHaveBeenCalledTimes(1)
  })

  it('transitions to error state on fetch failure', async () => {
    const error = new Error('Network error')
    const fetcher = vi.fn().mockRejectedValue(error)

    const { result } = renderHook(() => useFetch(fetcher, null))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.data).toBeNull()
    expect(result.current.error).toEqual(error)
  })

  it('wraps non-Error rejections in Error', async () => {
    const fetcher = vi.fn().mockRejectedValue('string error')

    const { result } = renderHook(() => useFetch(fetcher, null))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.error).toBeInstanceOf(Error)
    expect(result.current.error?.message).toBe('string error')
  })

  it('uses initialData as default', async () => {
    const fetcher = vi.fn().mockResolvedValue([1, 2, 3])

    const { result } = renderHook(() => useFetch(fetcher, []))

    expect(result.current.data).toEqual([])

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.data).toEqual([1, 2, 3])
  })

  it('reload refetches data', async () => {
    let callCount = 0
    const fetcher = vi.fn().mockImplementation(() => {
      callCount++
      return Promise.resolve({ count: callCount })
    })

    const { result } = renderHook(() => useFetch(fetcher, null))

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.data).toEqual({ count: 1 })

    result.current.reload()

    await waitFor(() => {
      expect(result.current.data).toEqual({ count: 2 })
    })

    expect(fetcher).toHaveBeenCalledTimes(2)
  })
})
