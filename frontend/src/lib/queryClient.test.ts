import { toast } from 'sonner'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { createQueryClient } from './queryClient'

vi.mock('sonner', () => ({
  toast: { error: vi.fn() },
}))

afterEach(() => {
  vi.restoreAllMocks()
})

describe('createQueryClient', () => {
  it('zeigt bei einem Query-Fehler einen globalen Fehler-Toast', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    const queryClient = createQueryClient()

    await expect(
      queryClient.fetchQuery({
        queryKey: ['test-query'],
        queryFn: () => Promise.reject(new Error('Netzabbruch')),
        retry: false,
      }),
    ).rejects.toThrow('Netzabbruch')

    expect(toast.error).toHaveBeenCalledWith(
      'Daten konnten nicht geladen werden. Bitte Verbindung prüfen und erneut versuchen.',
      { id: 'query-fehler' },
    )
  })
})
