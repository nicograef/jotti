import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  act,
  cleanup,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { BackendError } from '@/lib/Backend'

import type {
  DirektverkaufHistorieEintrag,
  DirektverkaufTaetigen,
} from './direktverkauf/Direktverkauf'
import { DIREKTVERKAUF_HISTORIE_KEY } from './direktverkauf/hooks'
import { DirektverkaufPage } from './DirektverkaufPage'
import { AKTIVE_PRODUKTE_KEY } from './product/hooks'
import type { Produkt } from './product/Produkt'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn(), info: vi.fn() },
}))

// Handy-Layout: Die Seite stellt den ServiceDock selbst, der Aktionsbutton des
// Direktverkaufs rendert per Portal hinein.
vi.mock('@/hooks/use-mobile', () => ({
  useIsMobile: () => true,
}))

vi.mock('@/lib/Auth', () => ({
  AuthSingleton: { userId: 1, canCancel: true },
}))

// Nur der Singleton wird ersetzt: Die Fehlerklassen bleiben echt, weil der
// Fehlerpfad der Aktion (use-action-submit) gegen sie prüft.
vi.mock('@/lib/Backend', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/Backend')>()),
  BackendSingleton: {},
}))

// Die Seite läuft gegen die echten Query-Hooks; nur die Backends sind ersetzt.
// Nur so ist prüfbar, dass ein gescheitertes Erstladen (leerer Cache) und ein
// gescheiterter Hintergrund-Refetch (gefüllter Cache) verschieden aussehen.
const { getAktiveProdukte, getDirektverkaufHistorie, direktverkaufTaetigen } =
  vi.hoisted(() => ({
    getAktiveProdukte: vi.fn<() => Promise<Produkt[]>>(),
    getDirektverkaufHistorie:
      vi.fn<() => Promise<DirektverkaufHistorieEintrag[]>>(),
    direktverkaufTaetigen: vi.fn<(v: DirektverkaufTaetigen) => Promise<void>>(),
  }))

vi.mock('./product/ProduktBackend', () => ({
  ProduktBackend: class {
    getAktiveProdukte = getAktiveProdukte
  },
}))

vi.mock('./direktverkauf/DirektverkaufBackend', () => ({
  DirektverkaufBackend: class {
    getDirektverkaufHistorie = getDirektverkaufHistorie
    direktverkaufTaetigen = direktverkaufTaetigen
  },
}))

const testProdukt: Produkt = {
  id: 1,
  name: 'Bratwurst',
  kategorie: 'essen',
  status: 'active',
  varianten: [
    {
      id: 1,
      name: 'Normal',
      preisCents: 350,
      status: 'active',
      createdAt: '2025-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z',
    },
  ],
  createdAt: '2025-01-01T00:00:00Z',
  updatedAt: '2025-01-01T00:00:00Z',
}

// Zweites Produkt derselben Kategorie, damit beide Varianten-Zeilen ohne
// Kategoriewechsel nebeneinander stehen.
const testBeilage: Produkt = {
  id: 2,
  name: 'Pommes',
  kategorie: 'essen',
  status: 'active',
  varianten: [
    {
      id: 2,
      name: 'Normal',
      preisCents: 250,
      status: 'active',
      createdAt: '2025-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z',
    },
  ],
  createdAt: '2025-01-01T00:00:00Z',
  updatedAt: '2025-01-01T00:00:00Z',
}

const verkauf: DirektverkaufHistorieEintrag = {
  verkaufId: '11111111-1111-1111-1111-111111111111',
  userName: 'Anna',
  getaetigtAm: '2026-06-08T10:00:00Z',
  positionen: [
    {
      positionId: '22222222-2222-2222-2222-222222222222',
      varianteId: 1,
      produktName: 'Cola',
      varianteName: '0,5l',
      kategorie: 'getraenk',
      steuersatz: 'regel',
      einzelpreisCents: 500,
      menge: 2,
    },
  ],
  gesamtbetragCents: 1000,
  kommentar: '',
  offenePositionen: [],
  gesamtStorniertCents: 0,
  stornierungen: [],
}

beforeEach(() => {
  getAktiveProdukte.mockResolvedValue([])
  getDirektverkaufHistorie.mockResolvedValue([])
})

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
})

function renderPage() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  render(
    <QueryClientProvider client={queryClient}>
      <DirektverkaufPage />
    </QueryClientProvider>,
  )
  return { queryClient }
}

describe('DirektverkaufPage', () => {
  it('zeigt bei gescheitertem Erstladen der Produkte den Ladefehler', async () => {
    getAktiveProdukte.mockRejectedValue(new Error('Netzabbruch'))
    renderPage()

    expect(
      await screen.findByText('Produkte konnten nicht geladen werden'),
    ).toBeInTheDocument()
    expect(
      screen.queryByText('Keine Produkte verfügbar'),
    ).not.toBeInTheDocument()
  })

  // Der Ladefehler gilt nur dem gescheiterten Erstladen. Scheitert ein
  // Hintergrund-Refetch, bleiben die zuletzt geladenen Daten gültig; die
  // Meldung trägt der zentrale Fehler-Toast aus queryClient.ts.
  it('lässt die Produkte stehen, wenn ein Hintergrund-Refetch scheitert', async () => {
    getAktiveProdukte
      .mockResolvedValueOnce([testProdukt])
      .mockRejectedValue(new Error('Netzabbruch'))
    const { queryClient } = renderPage()

    await screen.findByText('Bratwurst')
    await act(async () => {
      await queryClient.refetchQueries()
    })
    // Die Query steht auf „error", die Seite hat die Aktualisierung verarbeitet
    // — erst danach ist die Anzeige aussagekräftig.
    await waitFor(() => {
      expect(queryClient.getQueryState([AKTIVE_PRODUKTE_KEY])?.status).toBe(
        'error',
      )
    })

    expect(getAktiveProdukte).toHaveBeenCalledTimes(2)
    expect(screen.getByText('Bratwurst')).toBeInTheDocument()
    expect(
      screen.queryByText('Produkte konnten nicht geladen werden'),
    ).not.toBeInTheDocument()
  })

  it('zeigt bei gescheitertem Erstladen der Historie den Ladefehler', async () => {
    getDirektverkaufHistorie.mockRejectedValue(new Error('Netzabbruch'))
    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByRole('tab', { name: 'Historie' }))

    expect(
      await screen.findByText('Historie konnte nicht geladen werden'),
    ).toBeInTheDocument()
    expect(
      screen.queryByText(/Noch keine Direktverkäufe/),
    ).not.toBeInTheDocument()
  })

  // Der 409 `vorgang_daten_abweichend` belegt, dass der Verkauf unter diesem
  // Schlüssel gebucht ist — nur seine Antwort ging verloren. Bliebe die Auswahl
  // stehen, bliebe auch die verkaufId stehen, und die Meldung („nur die
  // Differenz erneut erfassen") führte in genau denselben 409 zurück.
  it('räumt Auswahl, Schlüssel und Historie ab, wenn der Server den Verkauf als gebucht meldet', async () => {
    getAktiveProdukte.mockResolvedValue([testProdukt, testBeilage])
    direktverkaufTaetigen
      .mockRejectedValueOnce(
        new BackendError(409, 'vorgang_daten_abweichend', ''),
      )
      // Der zweite Versuch scheitert am Netz, damit der Drawer offen bleibt und
      // kein Erfolgs-Pop dazwischenkommt — geprüft wird nur sein Schlüssel.
      .mockRejectedValue(new Error('Netzabbruch'))
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('Bratwurst')
    await user.click(
      screen.getAllByRole('button', { name: 'Variante hinzufügen' })[0],
    )
    await user.click(screen.getByRole('button', { name: /Kassieren/ }))
    const drawer = await screen.findByRole('dialog')
    // Vor dem Absenden gezählt: Das Abräumen läuft synchron im Fehlerpfad, der
    // Refetch wäre sonst schon gelaufen, bevor der Zähler steht.
    const historieAbrufe = getDirektverkaufHistorie.mock.calls.length
    await user.click(
      within(drawer).getByRole('button', { name: 'Verkauf abschließen' }),
    )
    await waitFor(() => {
      expect(direktverkaufTaetigen).toHaveBeenCalledTimes(1)
    })
    const ersterKey = direktverkaufTaetigen.mock.calls[0][0].verkaufId

    // Der abgeschlossene Vorgang schließt den Drawer und leert die Auswahl.
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })
    expect(screen.getByRole('button', { name: /Kassieren/ })).toBeDisabled()
    // Die Historie lädt neu — nur so sieht der Helfer, was tatsächlich gebucht
    // ist, und weiß, was die Differenz ist.
    await waitFor(() => {
      expect(getDirektverkaufHistorie.mock.calls.length).toBeGreaterThan(
        historieAbrufe,
      )
    })

    // Die Differenz nachverkaufen: eigene Zusammenstellung, eigener Schlüssel.
    await user.click(
      screen.getAllByRole('button', { name: 'Variante hinzufügen' })[1],
    )
    await user.click(screen.getByRole('button', { name: /Kassieren/ }))
    const zweiterDrawer = await screen.findByRole('dialog')
    await user.click(
      within(zweiterDrawer).getByRole('button', {
        name: 'Verkauf abschließen',
      }),
    )
    await waitFor(() => {
      expect(direktverkaufTaetigen).toHaveBeenCalledTimes(2)
    })

    const zweiterAufruf = direktverkaufTaetigen.mock.calls[1][0]
    expect(zweiterAufruf.positionen).toEqual([
      { produktId: 2, varianteId: 2, menge: 1 },
    ])
    expect(zweiterAufruf.verkaufId).not.toBe(ersterKey)
  })

  it('lässt die Historie stehen, wenn ein Hintergrund-Refetch scheitert', async () => {
    getDirektverkaufHistorie
      .mockResolvedValueOnce([verkauf])
      .mockRejectedValue(new Error('Netzabbruch'))
    const user = userEvent.setup()
    const { queryClient } = renderPage()

    await user.click(screen.getByRole('tab', { name: 'Historie' }))
    await screen.findByText('2× Cola 0,5l')

    await act(async () => {
      await queryClient.refetchQueries()
    })
    await waitFor(() => {
      expect(
        queryClient.getQueryState([DIREKTVERKAUF_HISTORIE_KEY])?.status,
      ).toBe('error')
    })

    expect(getDirektverkaufHistorie).toHaveBeenCalledTimes(2)
    expect(screen.getByText('2× Cola 0,5l')).toBeInTheDocument()
    expect(
      screen.queryByText('Historie konnte nicht geladen werden'),
    ).not.toBeInTheDocument()
  })
})
