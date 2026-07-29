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

// zaehleVersuche führt eine Lese-Query mit den echten Client-Defaults aus und
// meldet, wie oft die Query-Funktion dabei aufgerufen wurde. `retryDelay: 0`
// überschreibt nur die Wartezeit zwischen den Versuchen, nicht die Politik.
async function zaehleVersuche(fehler: Error): Promise<number> {
  let versuche = 0

  await expect(
    createQueryClient().fetchQuery({
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
  // Eine abgelehnte Buchung oder eine fehlende Berechtigung steht schon beim
  // ersten Versuch fest — die Helferin soll die Meldung sofort sehen.
  it('wiederholt einen 4xx-Fehler nicht', async () => {
    expect(await zaehleVersuche(new BackendError(409, 'conflict'))).toBe(1)
  })

  it('wiederholt einen ResponseBodyError nicht', async () => {
    expect(await zaehleVersuche(new ResponseBodyError('Schema verletzt'))).toBe(
      1,
    )
  })

  // Ein abgebrochener fetch wirft einen nackten TypeError; im Vereins-WLAN ist
  // das meist vorübergehend.
  it('wiederholt einen Netzfehler zweimal', async () => {
    expect(await zaehleVersuche(new TypeError('Failed to fetch'))).toBe(3)
  })

  it('wiederholt einen Serverfehler ab Status 500 zweimal', async () => {
    expect(
      await zaehleVersuche(new BackendError(500, 'internal_server_error')),
    ).toBe(3)
  })

  // Buchungen laufen über useActionSubmit an react-query vorbei, der
  // DSFinV-K-Export als einzige Mutation bleibt bei den Bibliotheks-Defaults:
  // Ein automatisch wiederholter Schreibvorgang würde doppelt buchen.
  it('setzt keine Politik für Mutations', () => {
    expect(createQueryClient().getDefaultOptions().mutations).toBeUndefined()
  })
})

describe('createQueryClient Fehler-Toast', () => {
  it('zeigt bei einem Query-Fehler einen globalen Fehler-Toast', async () => {
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

  it('zeigt die Korrelations-ID, wenn die Antwort eine liefert', async () => {
    const queryClient = createQueryClient()

    await expect(
      queryClient.fetchQuery({
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
      queryClient.fetchQuery({
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
