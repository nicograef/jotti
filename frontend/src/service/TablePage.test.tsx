import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'

import type { Produkt } from './product/Produkt'
import type { Position } from './table/Bestellung'
import type { TischSession } from './table/Tisch'
import { TablePage } from './TablePage'

// Mit der Produktebene liegt die Variantenliste eine Navigationsebene tiefer:
// erst das Produkt oeffnen, dann ist der Hinzufuegen-Knopf da.
async function produktOeffnen(
  user: ReturnType<typeof userEvent.setup>,
  name = 'Bratwurst',
) {
  await user.click(screen.getByRole('button', { name: new RegExp(name) }))
}

function position(positionId: string): Position {
  return {
    positionId,
    varianteId: 1,
    produktName: 'Bratwurst',
    varianteName: 'Normal',
    kategorie: 'essen',
    steuersatz: 'regel',
    einzelpreisCents: 350,
    menge: 1,
    bestellerUserId: 1,
    bestellerName: 'Tester',
  }
}

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

// Steuerbarer Testzustand: `tischId` bildet den :tischId-Param nach (Tischwechsel
// ohne Remount), `produkte` speist die Bestell-Tab-Auswahl.
const testState = vi.hoisted(() => ({
  tischId: '1',
  produkte: [] as Produkt[],
}))

vi.mock('react-router', () => ({
  useParams: () => ({ tischId: testState.tischId }),
}))

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

// Diese Suite prüft Kopfbereich und Fehlerzustand (identisch in beiden
// Layouts). Auf dem Handy-Pfad trägt nur der Kopf den Tischnamen; ab lg zeigt
// ihn zusätzlich die Abschluss-Spalte — der Split selbst ist manuelle Abnahme.
vi.mock('@/hooks/use-mobile', () => ({
  useIsMobile: () => true,
}))

vi.mock('@/lib/Backend', () => ({
  BackendSingleton: {},
}))

// Die eigene Servicekraft (für die „Meine Positionen"-Filterung in Zahlung);
// canCancel/canRebook, damit der Storno-/Umbuchen-Pfad der Historie greift.
vi.mock('@/lib/Auth', () => ({
  AuthSingleton: { userId: 1, canCancel: true, canRebook: true },
}))

vi.mock('./product/hooks', () => ({
  useAktiveProdukte: () => ({ produkte: testState.produkte, isPending: false }),
}))

const { getTischState, getTischHistorie, stornierungErteilen } = vi.hoisted(
  () => ({
    getTischState: vi.fn<() => Promise<TischSession>>(),
    getTischHistorie: vi.fn<() => Promise<unknown[]>>(),
    stornierungErteilen: vi.fn<() => Promise<void>>(),
  }),
)

vi.mock('./table/TischBackend', () => ({
  TischBackend: class {
    getTischState = getTischState
    getTischHistorie = getTischHistorie
    stornierungErteilen = stornierungErteilen
  },
}))

// Tischzustand mit offenem Saldo. Der Saldo ist bewusst ungleich 0, damit er
// sich im DOM eindeutig von den 0,00-€-Summen der Bestell-Leiste unterscheidet.
const stammtisch: TischSession = {
  tischId: 1,
  tischName: 'Stammtisch',
  saldoCents: 1250,
  unbezahltePositionen: [],
  fuerMichErledigt: true,
}

beforeEach(() => {
  VorgangsRegisterSingleton.zuruecksetzen()
})

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
  testState.tischId = '1'
  testState.produkte = []
})

function renderPage() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  render(
    <QueryClientProvider client={queryClient}>
      <TablePage />
    </QueryClientProvider>,
  )
}

describe('TablePage', () => {
  it('zeigt bei Query-Fehler einen Fehlerzustand statt der Leer-Defaults', async () => {
    getTischState.mockRejectedValue(new Error('Netzabbruch'))
    getTischHistorie.mockRejectedValue(new Error('Netzabbruch'))
    renderPage()

    expect(
      await screen.findByText('Tischdaten konnten nicht geladen werden'),
    ).toBeInTheDocument()
    // Der Leer-Default (Saldo 0,00 €) darf bei einem Fehler nicht erscheinen —
    // der Tisch wirkt sonst fälschlich abgerechnet.
    expect(screen.queryByText('0,00 €')).not.toBeInTheDocument()
  })

  it('lädt die Tischdaten über „Erneut versuchen" nach einem Fehler neu', async () => {
    getTischState
      .mockRejectedValueOnce(new Error('Netzabbruch'))
      .mockResolvedValue(stammtisch)
    getTischHistorie
      .mockRejectedValueOnce(new Error('Netzabbruch'))
      .mockResolvedValue([])
    const user = userEvent.setup()
    renderPage()

    await user.click(
      await screen.findByRole('button', { name: 'Erneut versuchen' }),
    )

    expect(await screen.findByText('Stammtisch')).toBeInTheDocument()
    expect(
      screen.queryByText('Tischdaten konnten nicht geladen werden'),
    ).not.toBeInTheDocument()
  })

  it('zeigt ohne Fehler den Tischzustand mit Saldo', async () => {
    getTischState.mockResolvedValue(stammtisch)
    getTischHistorie.mockResolvedValue([])
    renderPage()

    expect(await screen.findByText('Stammtisch')).toBeInTheDocument()
    expect(screen.getByText('12,50 €')).toBeInTheDocument()
  })

  it('zeigt "Alles bezahlt" ohne unbezahlte Positionen', async () => {
    getTischState.mockResolvedValue(stammtisch)
    getTischHistorie.mockResolvedValue([])
    renderPage()

    expect(await screen.findByText('Alles bezahlt')).toBeInTheDocument()
  })

  it('zeigt die Anzahl unbezahlter Positionen als Badge', async () => {
    getTischState.mockResolvedValue({
      ...stammtisch,
      unbezahltePositionen: [position('p1'), position('p2')],
    })
    getTischHistorie.mockResolvedValue([])
    renderPage()

    const badge = await screen.findByText('2 unbezahlt')
    expect(screen.queryByText('Alles bezahlt')).not.toBeInTheDocument()
    // "Unbezahlt" wartet auf die Servicekraft, ist aber kein Gefahrenzustand:
    // Amber-Warn-Variante statt des Gefahren-Rots (destructive).
    expect(badge).toHaveAttribute('data-variant', 'warn')
  })

  // A1: Der gehobene Auswahl-State überlebt das Aus- und Wiedereinhängen der
  // Radix-Tab-Inhalte (inaktive Tabs werden ausgehängt). Ohne das Heben nach
  // TablePage ginge die Auswahl beim Tab-Wechsel verloren.
  it('behält den Bestell-Korb über einen Tab-Wechsel hinweg', async () => {
    testState.produkte = [testProdukt]
    getTischState.mockResolvedValue(stammtisch)
    getTischHistorie.mockResolvedValue([])
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('Stammtisch')
    // Bestellen ist der Default-Tab: eine Variante in den Korb legen.
    await produktOeffnen(user)
    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    expect(
      screen.getByRole('button', { name: /Bestellung überprüfen/ }),
    ).toHaveTextContent('3,50')

    // Zur Historie und zurück — der Bestellen-Tab wird zwischenzeitlich ausgehängt.
    await user.click(screen.getByRole('tab', { name: 'Historie' }))
    await user.click(screen.getByRole('tab', { name: 'Bestellen' }))

    expect(
      screen.getByRole('button', { name: /Bestellung überprüfen/ }),
    ).toHaveTextContent('3,50')
  })

  it('behält die Kassieren-Auswahl über einen Tab-Wechsel hinweg', async () => {
    getTischState.mockResolvedValue({
      ...stammtisch,
      unbezahltePositionen: [position('p1')],
    })
    getTischHistorie.mockResolvedValue([])
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('Stammtisch')
    await user.click(screen.getByRole('tab', { name: 'Kassieren' }))
    // Die eigene Position auswählen (Auth-userId 1).
    await user.click(screen.getByRole('button', { name: 'Produkt hinzufügen' }))
    expect(screen.getByRole('button', { name: /Kassieren/ })).toHaveTextContent(
      '3,50',
    )

    await user.click(screen.getByRole('tab', { name: 'Historie' }))
    await user.click(screen.getByRole('tab', { name: 'Kassieren' }))

    expect(screen.getByRole('button', { name: /Kassieren/ })).toHaveTextContent(
      '3,50',
    )
  })

  // A1: Der useMengen-`max` deckelt nur beim `add`. Schrumpft die unbezahlte
  // Menge einer bereits ausgewählten Position (Storno-Refetch, der erst beim
  // Schließen des Erfolgs-Pops eintrifft), muss die gehobene Auswahl auf die
  // neue Obergrenze sinken; eine verschwundene Position fällt heraus.
  it('deckelt die Kassieren-Auswahl, wenn ein Refetch kleinere unbezahlte Mengen liefert', async () => {
    const posMehr = { ...position('p1'), menge: 2 }
    const posWeg = position('p2')
    getTischState
      .mockResolvedValueOnce({
        ...stammtisch,
        unbezahltePositionen: [posMehr, posWeg],
      })
      .mockResolvedValue({
        ...stammtisch,
        unbezahltePositionen: [{ ...posMehr, menge: 1 }],
      })
    getTischHistorie.mockResolvedValue([
      {
        art: 'bestellung',
        id: '00000000-0000-0000-0000-000000000001',
        userId: 1,
        userName: 'Tester',
        tischId: 1,
        positionen: [posMehr],
        gesamtPreisCents: 700,
        kommentar: '',
        aufgenommenAm: '2026-06-18T12:00:00Z',
        stornierbarePositionen: [posMehr],
        umbuchbarePositionen: [],
      },
    ])
    stornierungErteilen.mockResolvedValue(undefined)
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('Stammtisch')

    // Kassieren: p1 voll (2 von 2) und p2 (1 von 1) auswählen.
    await user.click(screen.getByRole('tab', { name: 'Kassieren' }))
    await user.click(
      screen.getAllByRole('button', { name: 'Produkt hinzufügen' })[0],
    )
    await user.click(
      screen.getAllByRole('button', { name: 'Produkt hinzufügen' })[0],
    )
    await user.click(
      screen.getAllByRole('button', { name: 'Produkt hinzufügen' })[1],
    )
    expect(screen.getByRole('button', { name: /Kassieren/ })).toHaveTextContent(
      '10,50',
    )

    // Storno auf der Historie; der Refetch läuft erst beim Schließen des Pops.
    await user.click(screen.getByRole('tab', { name: 'Historie' }))
    await user.click(screen.getByRole('button', { name: /Bestellung/ }))
    await user.click(screen.getByRole('button', { name: /Stornieren…/ }))
    await user.click(screen.getByRole('button', { name: /hinzufügen/ }))
    await user.type(
      screen.getByPlaceholderText('Kommentar (erforderlich)'),
      'Falsch gebucht',
    )
    await user.click(
      screen.getByRole('button', { name: 'Stornierung erteilen' }),
    )
    await screen.findByText('Stornierung gebucht.')
    await user.click(screen.getByRole('status'))

    // Zurück auf Kassieren: p1 ist auf die neue Obergrenze (1) gedeckelt statt
    // der kaputten Über-Deckelung „2 von 1", p2 ist ganz verschwunden.
    await user.click(screen.getByRole('tab', { name: 'Kassieren' }))
    expect(await screen.findByText(/1 von 1 ausgewählt/)).toBeInTheDocument()
    expect(screen.queryByText(/2 von 1 ausgewählt/)).not.toBeInTheDocument()
    expect(
      screen.getAllByRole('button', { name: 'Produkt hinzufügen' }),
    ).toHaveLength(1)
    expect(screen.getByRole('button', { name: /Kassieren/ })).toHaveTextContent(
      '3,50',
    )
  })

  it('startet die Auswahl bei einem Tischwechsel leer', async () => {
    testState.produkte = [testProdukt]
    getTischState.mockResolvedValue(stammtisch)
    getTischHistorie.mockResolvedValue([])
    const user = userEvent.setup()
    // Eigener QueryClient, damit Re-Renders die Provider-Instanz teilen; jeder
    // Aufruf liefert ein frisches Element, sonst überspringt React das
    // Neurendern (referenzgleiche Props).
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    const renderUi = () => (
      <QueryClientProvider client={queryClient}>
        <TablePage />
      </QueryClientProvider>
    )
    const { rerender } = render(renderUi())

    await screen.findByText('Stammtisch')
    await produktOeffnen(user)
    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    expect(
      screen.getByRole('button', { name: /Bestellung überprüfen/ }),
    ).toHaveTextContent('3,50')

    // Anderer Tisch: nur der :tischId-Param wechselt, TablePage bleibt gemountet.
    testState.tischId = '2'
    rerender(renderUi())

    // Nach dem Laden des neuen Tisches ist der Korb leer (Aktionsbutton deaktiviert).
    await waitFor(() => {
      expect(
        screen.getByRole('button', { name: /Bestellung überprüfen/ }),
      ).toBeDisabled()
    })
  })

  // Der Tischwechsel ist einer der beiden realen Auslöser für ein Zähler-Leck
  // im Vorgangs-Register: TablePage bleibt gemountet und setzt den Korb nur
  // zurück. Ein stehen gebliebener Vorgang blockierte den erzwungenen Reload
  // dauerhaft, ohne dass es jemandem auffiele.
  it('gibt den Bestell-Korb beim Tischwechsel im Vorgangs-Register frei', async () => {
    testState.produkte = [testProdukt]
    getTischState.mockResolvedValue(stammtisch)
    getTischHistorie.mockResolvedValue([])
    const user = userEvent.setup()
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    const renderUi = () => (
      <QueryClientProvider client={queryClient}>
        <TablePage />
      </QueryClientProvider>
    )
    const { rerender, unmount } = render(renderUi())

    await screen.findByText('Stammtisch')
    await produktOeffnen(user)
    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)

    // Anderer Tisch: nur der :tischId-Param wechselt, TablePage bleibt gemountet.
    testState.tischId = '2'
    rerender(renderUi())
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)

    // Und auch das Verlassen der Seite hinterlässt keinen offenen Vorgang.
    unmount()
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })

  // A2: Eine Stornierung bestätigt über den Erfolgs-Pop; der Refetch des
  // Tisch-States läuft erst beim Schließen des Pops, nicht schon beim Erfolg.
  it('zeigt nach der Stornierung den Erfolgs-Pop und lädt erst beim Schließen neu', async () => {
    getTischState.mockResolvedValue(stammtisch)
    getTischHistorie.mockResolvedValue([
      {
        art: 'bestellung',
        id: '00000000-0000-0000-0000-000000000001',
        userId: 1,
        userName: 'Tester',
        tischId: 1,
        positionen: [position('p1')],
        gesamtPreisCents: 350,
        kommentar: '',
        aufgenommenAm: '2026-06-18T12:00:00Z',
        stornierbarePositionen: [position('p1')],
        umbuchbarePositionen: [],
      },
    ])
    stornierungErteilen.mockResolvedValue(undefined)
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('Stammtisch')
    const ladeCalls = getTischState.mock.calls.length

    await user.click(screen.getByRole('tab', { name: 'Historie' }))
    await user.click(screen.getByRole('button', { name: /Bestellung/ }))
    await user.click(screen.getByRole('button', { name: /Stornieren…/ }))
    await user.click(screen.getByRole('button', { name: /hinzufügen/ }))
    await user.type(
      screen.getByPlaceholderText('Kommentar (erforderlich)'),
      'Falsch gebucht',
    )
    await user.click(
      screen.getByRole('button', { name: 'Stornierung erteilen' }),
    )

    // Der Pop erscheint; bis zum Schließen läuft kein Refetch des Tisch-States.
    await screen.findByText('Stornierung gebucht.')
    expect(getTischState.mock.calls.length).toBe(ladeCalls)

    // Pop schließen (Tap auf das Overlay) → jetzt lädt der Tisch-State neu.
    await user.click(screen.getByRole('status'))
    await waitFor(() => {
      expect(getTischState.mock.calls.length).toBeGreaterThan(ladeCalls)
    })
  })
})
