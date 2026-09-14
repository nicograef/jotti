import { toast } from 'sonner'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { BackendError, ResponseBodyError } from './Backend'
import { createQueryClient } from './queryClient'

vi.mock('sonner', () => ({
  toast: { error: vi.fn() },
}))

beforeEach(() => {
  vi.spyOn(console, 'error').mockImplementation(() => undefined)
})

afterEach(() => {
  vi.clearAllMocks()
  vi.restoreAllMocks()
})

// Führt eine Lese-Query mit den echten Client-Defaults aus und zählt die
// Aufrufe. `retryDelay: 0` kürzt nur die Wartezeit, nicht die Politik.
async function zaehleVersuche(fehler: Error): Promise<number> {
  let versuche = 0

  await expect(
    createQueryClient().query({
      queryKey: ['versuche'],
      queryFn: () => {
        versuche += 1
        return Promise.reject(fehler)
      },
      retryDelay: 0,
    }),
  ).rejects.toBe(fehler)

  return versuche
}

describe('createQueryClient Wiederholungen', () => {
  // Steht schon beim ersten Versuch fest — die Meldung soll sofort kommen.
  it('wiederholt einen 4xx-Fehler nicht', async () => {
    expect(await zaehleVersuche(new BackendError(409, 'conflict'))).toBe(1)
  })

  it('wiederholt einen ResponseBodyError nicht', async () => {
    expect(await zaehleVersuche(new ResponseBodyError('Schema verletzt'))).toBe(
      1,
    )
  })

  // Ein abgebrochener fetch wirft einen nackten TypeError.
  it('wiederholt einen Netzfehler zweimal', async () => {
    expect(await zaehleVersuche(new TypeError('Failed to fetch'))).toBe(3)
  })

  it('wiederholt einen Serverfehler ab Status 500 zweimal', async () => {
    expect(
      await zaehleVersuche(new BackendError(500, 'internal_server_error')),
    ).toBe(3)
  })

  // Ein automatisch wiederholter Schreibvorgang würde doppelt buchen; Buchungen
  // laufen ohnehin über useActionSubmit an react-query vorbei.
  it('setzt keine Politik für Mutations', () => {
    expect(createQueryClient().getDefaultOptions().mutations).toBeUndefined()
  })
})

describe('createQueryClient Fehler-Toast', () => {
  it('zeigt bei einem Query-Fehler einen globalen Fehler-Toast', async () => {
    const queryClient = createQueryClient()

    await expect(
      queryClient.query({
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

  it('zeigt die Korrelations-ID, wenn die Antwort eine liefert', async () => {
    const queryClient = createQueryClient()

    await expect(
      queryClient.query({
        queryKey: ['test-query-referenz'],
        queryFn: () =>
          Promise.reject(
            new BackendError(500, 'internal_server_error', undefined, 'a1b2c3'),
          ),
        retry: false,
      }),
    ).rejects.toBeInstanceOf(BackendError)

    expect(toast.error).toHaveBeenCalledWith(
      'Daten konnten nicht geladen werden. Bitte Verbindung prüfen und erneut versuchen. Referenz: a1b2c3',
      { id: 'query-fehler' },
    )
  })

  it('bleibt ohne Korrelations-ID unverändert lesbar', async () => {
    const queryClient = createQueryClient()

    await expect(
      queryClient.query({
        queryKey: ['test-query-ohne-referenz'],
        queryFn: () =>
          Promise.reject(new BackendError(500, 'internal_server_error')),
        retry: false,
      }),
    ).rejects.toBeInstanceOf(BackendError)

    expect(toast.error).toHaveBeenCalledWith(
      'Daten konnten nicht geladen werden. Bitte Verbindung prüfen und erneut versuchen.',
      { id: 'query-fehler' },
    )
  })
})
