import { onlineManager, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, renderHook, waitFor } from '@testing-library/react'
import { createElement, type ReactNode } from 'react'
import { toast } from 'sonner'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useAllTische } from '@/admin/tables/hooks'
import type { Tisch } from '@/admin/tables/Tisch'
import { useAktiveProdukte } from '@/service/product/hooks'
import { useAktiveTische, useTischState } from '@/service/table/hooks'
import type { TischSession } from '@/service/table/Tisch'

import { BackendError, NetzwerkFehler, ResponseBodyError } from './Backend'
import { createQueryClient } from './queryClient'

vi.mock('sonner', () => ({
  toast: { error: vi.fn() },
}))

// Die beiden Hooks der Aktualitäts-Suite laufen gegen echte Query-Verdrahtung,
// aber ohne Netz: nur ihre Backend-Klassen sind ersetzt.
const { getTischState, getAktiveTische, getAktiveProdukte, getAllTische } =
  vi.hoisted(() => ({
    getTischState: vi.fn<() => Promise<TischSession>>(),
    getAktiveTische: vi.fn<() => Promise<never[]>>(),
    getAktiveProdukte: vi.fn<() => Promise<never[]>>(),
    getAllTische: vi.fn<() => Promise<Tisch[]>>(),
  }))

vi.mock('@/service/table/TischBackend', () => ({
  TischBackend: class {
    getTischState = getTischState
    getAktiveTische = getAktiveTische
  },
}))

vi.mock('@/admin/tables/TischBackend', () => ({
  TischBackend: class {
    getAllTische = getAllTische
  },
}))

vi.mock('@/service/product/ProduktBackend', () => ({
  ProduktBackend: class {
    getAktiveProdukte = getAktiveProdukte
  },
}))

afterEach(() => {
  cleanup()
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

const stammtisch: TischSession = {
  tischId: 1,
  tischName: 'Stammtisch',
  saldoCents: 1250,
  unbezahltePositionen: [],
  fuerMichErledigt: true,
}

// Ein Provider über einem gemeinsamen Client, damit zwei aufeinanderfolgende
// Mounts denselben Cache sehen.
function erzeugeWrapper() {
  const queryClient = createQueryClient()
  return function Wrapper({ children }: { children: ReactNode }) {
    return createElement(QueryClientProvider, { client: queryClient }, children)
  }
}

describe('Aktualität der Queries', () => {
  it('setzt keine Aktualitätsschwelle als Voreinstellung', () => {
    expect(
      createQueryClient().getDefaultOptions().queries?.staleTime,
    ).toBeUndefined()
  })

  // Kassen- und Tischzustand ist ohne Schwelle: Wer einen Tisch öffnet, sieht
  // nie einen Saldo aus dem Cache und kassiert dadurch zu wenig.
  it('lädt einen Tischzustand beim zweiten Mount neu', async () => {
    getTischState.mockResolvedValue(stammtisch)
    const wrapper = erzeugeWrapper()

    const ersteMontage = renderHook(() => useTischState(1), { wrapper })
    await waitFor(() => {
      expect(getTischState).toHaveBeenCalledTimes(1)
    })
    ersteMontage.unmount()

    renderHook(() => useTischState(1), { wrapper })

    await waitFor(() => {
      expect(getTischState).toHaveBeenCalledTimes(2)
    })
  })

  // Die Tischliste der Verwaltung trägt `saldoCents` — den offenen Saldo der
  // laufenden Kassensitzung. Er steuert den Löschen-/Deaktivieren-Guard; aus
  // dem Cache gäbe er einen soeben bebuchten Tisch zum Deaktivieren frei.
  it('lädt die Tischliste der Verwaltung beim zweiten Mount neu', async () => {
    getAllTische.mockResolvedValue([])
    const wrapper = erzeugeWrapper()

    const ersteMontage = renderHook(() => useAllTische(), { wrapper })
    await waitFor(() => {
      expect(getAllTische).toHaveBeenCalledTimes(1)
    })
    ersteMontage.unmount()

    renderHook(() => useAllTische(), { wrapper })

    await waitFor(() => {
      expect(getAllTische).toHaveBeenCalledTimes(2)
    })
  })

  // Auch die aktiven Tische des Service tragen `saldoCents`. Sie werden heute
  // nur als Ziel-Tisch-Auswahl benutzt, aber die Regel gilt der Nutzlast, nicht
  // dem aktuellen Verwendungszweck: Keine Tisch-Query ist ein Stammdatum.
  it('lädt die aktiven Tische des Service beim zweiten Mount neu', async () => {
    getAktiveTische.mockResolvedValue([])
    const wrapper = erzeugeWrapper()

    const ersteMontage = renderHook(() => useAktiveTische(), { wrapper })
    await waitFor(() => {
      expect(getAktiveTische).toHaveBeenCalledTimes(1)
    })
    ersteMontage.unmount()

    renderHook(() => useAktiveTische(), { wrapper })

    await waitFor(() => {
      expect(getAktiveTische).toHaveBeenCalledTimes(2)
    })
  })

  // Stammdaten dagegen kommen innerhalb der Schwelle aus dem Cache: Die
  // Produktliste ändert sich während einer Veranstaltung nicht.
  it('lädt die Produktliste innerhalb der Stammdaten-Schwelle beim zweiten Mount nicht neu', async () => {
    getAktiveProdukte.mockResolvedValue([])
    const wrapper = erzeugeWrapper()

    const ersteMontage = renderHook(() => useAktiveProdukte(), { wrapper })
    await waitFor(() => {
      expect(getAktiveProdukte).toHaveBeenCalledTimes(1)
    })
    ersteMontage.unmount()

    const zweiteMontage = renderHook(() => useAktiveProdukte(), { wrapper })

    // Die Daten stehen sofort aus dem Cache bereit, ohne zweiten Abruf.
    await waitFor(() => {
      expect(zweiteMontage.result.current.isPending).toBe(false)
    })
    expect(getAktiveProdukte).toHaveBeenCalledTimes(1)
  })
})
