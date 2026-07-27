import { onlineManager } from '@tanstack/react-query'
import { toast } from 'sonner'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { BackendError, NetzwerkFehler, ResponseBodyError } from './Backend'
import { createQueryClient } from './queryClient'

vi.mock('sonner', () => ({
  toast: { error: vi.fn() },
}))

afterEach(() => {
  onlineManager.setOnline(true)
  vi.clearAllMocks()
  vi.restoreAllMocks()
})

// wiederholungsEntscheidung liest die tatsächlich konfigurierte retry-Funktion
// aus dem Client, damit der Test die Verdrahtung mitprüft.
function wiederholungsEntscheidung(
  anzahlFehlversuche: number,
  error: Error,
): boolean {
  const retry = createQueryClient().getDefaultOptions().queries?.retry
  if (typeof retry !== 'function') {
    throw new Error('retry ist nicht als Funktion konfiguriert')
  }
  return retry(anzahlFehlversuche, error)
}

describe('createQueryClient Fehler-Toast', () => {
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

  it('zeigt die Korrelations-ID im Toast, wenn der Fehler eine trägt', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
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

  // Im Funkloch meldet das Offline-Banner die Ursache; ein zusätzlicher
  // Ladefehler-Toast wäre eine falsche Aussage über den Server.
  it('unterdrückt den Toast, solange das Gerät offline ist', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    const queryClient = createQueryClient()
    onlineManager.setOnline(false)

    await expect(
      queryClient.fetchQuery({
        queryKey: ['test-query-offline'],
        queryFn: () => Promise.reject(new NetzwerkFehler('verbindungsabbruch')),
        retry: false,
        networkMode: 'always',
      }),
    ).rejects.toBeInstanceOf(NetzwerkFehler)

    expect(toast.error).not.toHaveBeenCalled()
  })
})

describe('createQueryClient Wiederholungen', () => {
  it('wiederholt einen NetzwerkFehler', () => {
    expect(
      wiederholungsEntscheidung(0, new NetzwerkFehler('zeitueberschreitung')),
    ).toBe(true)
  })

  it('wiederholt einen BackendError ab Status 500', () => {
    expect(
      wiederholungsEntscheidung(
        0,
        new BackendError(500, 'internal_server_error'),
      ),
    ).toBe(true)
  })

  it('wiederholt einen BackendError mit Status 4xx nicht', () => {
    expect(
      wiederholungsEntscheidung(0, new BackendError(409, 'conflict')),
    ).toBe(false)
  })

  it('wiederholt einen ResponseBodyError nicht', () => {
    expect(
      wiederholungsEntscheidung(0, new ResponseBodyError('Schema verletzt')),
    ).toBe(false)
  })

  it('wiederholt höchstens zweimal', () => {
    const fehler = new NetzwerkFehler('verbindungsabbruch')

    expect(wiederholungsEntscheidung(1, fehler)).toBe(true)
    expect(wiederholungsEntscheidung(2, fehler)).toBe(false)
  })
})

describe('createQueryClient Aktualität', () => {
  it('setzt eine Aktualitätsschwelle von 30 Sekunden', () => {
    expect(createQueryClient().getDefaultOptions().queries?.staleTime).toBe(
      30_000,
    )
  })

  // Nach einer Buchung muss der Tisch-Saldo sofort neu geladen werden — die
  // Schwelle darf das nicht verzögern.
  it('lädt nach invalidateQueries trotz der Schwelle sofort neu', async () => {
    const queryClient = createQueryClient()
    let aufrufe = 0
    const optionen = {
      queryKey: ['tisch-state'],
      queryFn: () => {
        aufrufe += 1
        return Promise.resolve(aufrufe)
      },
    }

    await queryClient.fetchQuery(optionen)
    await queryClient.fetchQuery(optionen)
    expect(aufrufe).toBe(1)

    await queryClient.invalidateQueries({ queryKey: ['tisch-state'] })
    await queryClient.fetchQuery(optionen)

    expect(aufrufe).toBe(2)
  })
})
