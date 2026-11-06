import { cleanup, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { BackendError } from '@/lib/Backend'
import type { Produkt } from '@/lib/produktSchemas'
import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'
import { signIn, signOut } from '@/test/auth'
import { FakeBackend } from '@/test/FakeBackend'
import { testPosition, testProdukt } from '@/test/fixtures'
import { renderWithBackend, setViewportWidth } from '@/test/render'

import type { Position } from './Bestellung'
import { TablePage } from './TablePage'
import type { TischSession } from './Tisch'

// Position ids are UUIDs; `nr` keeps them apart.
function position(nr: number): Position {
  return testPosition({
    positionId: `00000000-0000-4000-8000-${String(nr).padStart(12, '0')}`,
  })
}

const pfandProdukt: Produkt = {
  ...testProdukt,
  id: 9,
  name: 'Pfand',
  kategorie: 'sonstiges',
  steuersatz: 'regel',
  pfand: true,
  varianten: [
    { ...testProdukt.varianten[0], id: 19, name: 'Glas', preisCents: 200 },
  ],
}

// `tischId` bildet den :tischId-Param nach — Tischwechsel ohne Remount.
const testState = vi.hoisted(() => ({ tischId: '1' }))

vi.mock('react-router', () => ({
  useParams: () => ({ tischId: testState.tischId }),
}))

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

const getTischState = vi.fn<() => TischSession>()
const getTischHistorie = vi.fn<() => unknown[]>()

// Table state, history and products; without `produkte` the range is empty.
function backend(
  produkte: Produkt[] = [],
  pfandProdukte: Produkt[] = [],
): FakeBackend {
  return new FakeBackend()
    .respond('service/get-tisch-state', getTischState)
    .respond('service/get-tisch-historie', () => ({
      historie: getTischHistorie(),
    }))
    .respond('service/get-aktive-produkte', { produkte, pfandProdukte })
    .respond('service/get-pfandrueckgabe-historie', { historie: [] })
    .respond('serviceleitung/stornierung-erteilen', {})
}

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
  // Handy-Pfad: Kopfbereich und Fehlerzustand sind in beiden Layouts gleich;
  // der Split selbst ist manuelle Abnahme.
  setViewportWidth(375)
  // User 1 owns the positions ("Meine auswählen" in Zahlung); Serviceleitung
  // unlocks the Storno and Umbuchen paths of the history.
  signIn({ userId: 1, role: 'serviceleitung' })
})

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
  signOut()
  testState.tischId = '1'
})

function renderPage(fake: FakeBackend = backend()) {
  return renderWithBackend(<TablePage />, fake)
}

describe('TablePage', () => {
  // Schon eine der beiden Abfragen reicht für den Fehlerzustand.
  it.each(['service/get-tisch-state', 'service/get-tisch-historie'])(
    'zeigt bei Fehler von %s einen Fehlerzustand statt der Leer-Defaults',
    async (endpoint) => {
      getTischState.mockReturnValue(stammtisch)
      getTischHistorie.mockReturnValue([])
      renderPage(backend().fail(endpoint))

      expect(
        await screen.findByText('Tischdaten konnten nicht geladen werden'),
      ).toBeInTheDocument()
      expect(
        screen.getByRole('button', { name: 'Erneut versuchen' }),
      ).toBeInTheDocument()
      // Der Leer-Default (Saldo 0,00 €) darf bei einem Fehler nicht erscheinen —
      // der Tisch wirkt sonst fälschlich abgerechnet.
      expect(screen.queryByText('0,00 €')).not.toBeInTheDocument()
    },
  )

  it('zeigt bei Produkt-Fehler den Bestellen-Tab als Fehlerzustand statt leerer Liste', async () => {
    getTischState.mockReturnValue(stammtisch)
    getTischHistorie.mockReturnValue([])
    renderPage(backend().fail('service/get-aktive-produkte'))

    expect(
      await screen.findByText('Produkte konnten nicht geladen werden'),
    ).toBeInTheDocument()
    // Die Leer-Defaults des Bestellen-Tabs (Korb-Summe 0,00 €) dürfen bei
    // einem Fehler nicht erscheinen — das Sortiment wirkt sonst leer.
    expect(screen.queryByText(/0,00 €/)).not.toBeInTheDocument()
  })

  it('lädt die Tischdaten über „Erneut versuchen" nach einem Fehler neu', async () => {
    const abbruch = () => {
      throw new BackendError(400, 'test_fehler')
    }
    getTischState.mockImplementationOnce(abbruch).mockReturnValue(stammtisch)
    getTischHistorie.mockImplementationOnce(abbruch).mockReturnValue([])
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
    getTischState.mockReturnValue(stammtisch)
    getTischHistorie.mockReturnValue([])
    renderPage()

    expect(await screen.findByText('Stammtisch')).toBeInTheDocument()
    expect(screen.getByText('12,50 €')).toBeInTheDocument()
  })

  it('zeigt "Alles bezahlt" ohne unbezahlte Positionen', async () => {
    getTischState.mockReturnValue(stammtisch)
    getTischHistorie.mockReturnValue([])
    renderPage()

    expect(await screen.findByText('Alles bezahlt')).toBeInTheDocument()
  })

  it('zeigt die Anzahl unbezahlter Positionen als Badge', async () => {
    getTischState.mockReturnValue({
      ...stammtisch,
      unbezahltePositionen: [position(1), position(2)],
    })
    getTischHistorie.mockReturnValue([])
    renderPage()

    const badge = await screen.findByText('2 unbezahlt')
    expect(screen.queryByText('Alles bezahlt')).not.toBeInTheDocument()
    // „Unbezahlt" wartet auf die Servicekraft, ist kein Gefahrenzustand: Warn-Amber
    // statt destructive.
    expect(badge).toHaveAttribute('data-variant', 'warn')
  })

  // Radix hängt inaktive Tab-Inhalte aus; ohne den nach TablePage gehobenen
  // State ginge die Auswahl beim Tab-Wechsel verloren.
  it('behält den Bestell-Korb über einen Tab-Wechsel hinweg', async () => {
    getTischState.mockReturnValue(stammtisch)
    getTischHistorie.mockReturnValue([])
    const user = userEvent.setup()
    renderPage(backend([testProdukt]))

    await screen.findByText('Stammtisch')
    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    expect(
      screen.getByRole('button', { name: /Bestellung überprüfen/ }),
    ).toHaveTextContent('3,50')

    await user.click(screen.getByRole('tab', { name: 'Historie' }))
    await user.click(screen.getByRole('tab', { name: 'Bestellen' }))

    expect(
      screen.getByRole('button', { name: /Bestellung überprüfen/ }),
    ).toHaveTextContent('3,50')
  })

  it('behält die Kassieren-Auswahl über einen Tab-Wechsel hinweg', async () => {
    getTischState.mockReturnValue({
      ...stammtisch,
      unbezahltePositionen: [position(1)],
    })
    getTischHistorie.mockReturnValue([])
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('Stammtisch')
    await user.click(screen.getByRole('tab', { name: 'Kassieren' }))
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

  // Der useMengen-`max` deckelt nur beim `add`: schrumpft die unbezahlte Menge
  // einer ausgewählten Position (Storno-Refetch beim Schließen des Erfolgs-Pops),
  // muss die gehobene Auswahl sinken; eine verschwundene Position fällt heraus.
  it('deckelt die Kassieren-Auswahl, wenn ein Refetch kleinere unbezahlte Mengen liefert', async () => {
    const posMehr = { ...position(1), menge: 2 }
    const posWeg = position(2)
    getTischState
      .mockReturnValueOnce({
        ...stammtisch,
        unbezahltePositionen: [posMehr, posWeg],
      })
      .mockReturnValue({
        ...stammtisch,
        unbezahltePositionen: [{ ...posMehr, menge: 1 }],
      })
    getTischHistorie.mockReturnValue([
      {
        art: 'bestellung',
        id: '00000000-0000-4000-8000-000000000101',
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
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('Stammtisch')

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

    // p1 ist auf die neue Obergrenze (1) gedeckelt, p2 ist verschwunden.
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
    getTischState.mockReturnValue(stammtisch)
    getTischHistorie.mockReturnValue([])
    const user = userEvent.setup()
    const { rerender } = renderPage(backend([testProdukt]))

    await screen.findByText('Stammtisch')
    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    expect(
      screen.getByRole('button', { name: /Bestellung überprüfen/ }),
    ).toHaveTextContent('3,50')

    // Anderer Tisch: nur der :tischId-Param wechselt, TablePage bleibt gemountet.
    testState.tischId = '2'
    rerender(<TablePage />)

    await waitFor(() => {
      expect(
        screen.getByRole('button', { name: /Bestellung überprüfen/ }),
      ).toBeDisabled()
    })
  })

  // Die Bestellung ist gebucht, die Antwort geht verloren. Nach einem Tab-Wechsel
  // muss die Wiederholung denselben Inhalt und dieselbe bestellungId tragen, sonst
  // bucht sie doppelt.
  it('behält Kommentar und bestellungId über einen Tab-Wechsel hinweg', async () => {
    setViewportWidth(1024)
    getTischState.mockReturnValue(stammtisch)
    getTischHistorie.mockReturnValue([])
    const bestellungAufnehmen = vi
      .fn()
      .mockImplementationOnce(() => {
        throw new Error('Netzwerkfehler')
      })
      .mockReturnValue({})
    const fake = backend([testProdukt]).respond(
      'service/bestellung-aufnehmen',
      bestellungAufnehmen,
    )
    const user = userEvent.setup()
    renderPage(fake)

    await screen.findByRole('heading', { name: 'Stammtisch', level: 1 })
    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    await user.type(screen.getByPlaceholderText(/Kommentar/), 'ohne Zwiebeln')
    await user.click(
      screen.getByRole('button', { name: 'Bestellung aufnehmen' }),
    )
    await waitFor(() => {
      expect(bestellungAufnehmen).toHaveBeenCalledTimes(1)
    })

    await user.click(screen.getByRole('tab', { name: 'Historie' }))
    await user.click(screen.getByRole('tab', { name: 'Bestellen' }))
    await user.click(
      screen.getByRole('button', { name: 'Bestellung aufnehmen' }),
    )
    await waitFor(() => {
      expect(bestellungAufnehmen).toHaveBeenCalledTimes(2)
    })

    const [erster, wiederholung] = fake.bodies(
      'service/bestellung-aufnehmen',
    ) as [
      { bestellungId: string; kommentar: string },
      { bestellungId: string; kommentar: string },
    ]
    expect(wiederholung.kommentar).toBe('ohne Zwiebeln')
    expect(wiederholung.bestellungId).toBe(erster.bestellungId)
  })

  // Die Rückgabe ist gebucht, die Antwort geht verloren. Nach einem Tab-Wechsel
  // muss die Wiederholung dieselbe rueckgabeId tragen, sonst zahlt sie doppelt aus.
  it('behält Pfand-Auswahl und rueckgabeId über einen Tab-Wechsel hinweg', async () => {
    setViewportWidth(1024)
    getTischState.mockReturnValue(stammtisch)
    getTischHistorie.mockReturnValue([])
    const pfandZurueckgeben = vi
      .fn()
      .mockImplementationOnce(() => {
        throw new Error('Netzwerkfehler')
      })
      .mockReturnValue({})
    const fake = backend([], [pfandProdukt]).respond(
      'service/pfand-zurueckgeben',
      pfandZurueckgeben,
    )
    const user = userEvent.setup()
    renderPage(fake)

    await screen.findByRole('heading', { name: 'Stammtisch', level: 1 })
    await user.click(screen.getByRole('tab', { name: 'Kassieren' }))
    await user.click(
      await screen.findByRole('button', { name: 'Variante hinzufügen' }),
    )
    await user.click(screen.getByRole('button', { name: 'Pfand auszahlen' }))
    await waitFor(() => {
      expect(pfandZurueckgeben).toHaveBeenCalledTimes(1)
    })

    await user.click(screen.getByRole('tab', { name: 'Bestellen' }))
    await user.click(screen.getByRole('tab', { name: 'Kassieren' }))
    await user.click(screen.getByRole('button', { name: 'Pfand auszahlen' }))
    await waitFor(() => {
      expect(pfandZurueckgeben).toHaveBeenCalledTimes(2)
    })

    const [erster, wiederholung] = fake.bodies(
      'service/pfand-zurueckgeben',
    ) as [{ rueckgabeId: string }, { rueckgabeId: string }]
    expect(wiederholung.rueckgabeId).toBe(erster.rueckgabeId)
  })

  it('startet die Pfand-Auswahl bei einem Tischwechsel leer', async () => {
    setViewportWidth(1024)
    getTischState.mockReturnValue(stammtisch)
    getTischHistorie.mockReturnValue([])
    const user = userEvent.setup()
    const { rerender } = renderPage(backend([], [pfandProdukt]))

    await screen.findByRole('heading', { name: 'Stammtisch', level: 1 })
    await user.click(screen.getByRole('tab', { name: 'Kassieren' }))
    await user.click(
      await screen.findByRole('button', { name: 'Variante hinzufügen' }),
    )
    expect(
      screen.getByRole('button', { name: 'Pfand auszahlen' }),
    ).toBeInTheDocument()

    // Anderer Tisch: nur der :tischId-Param wechselt, TablePage bleibt gemountet.
    testState.tischId = '2'
    rerender(<TablePage />)

    // Erst den geladenen Tisch 2 abwarten: während des Ladens fehlt der Button ohnehin.
    await screen.findByRole('heading', { name: 'Stammtisch', level: 1 })
    expect(
      screen.queryByRole('button', { name: 'Pfand auszahlen' }),
    ).not.toBeInTheDocument()
  })

  // Tischwechsel: TablePage bleibt gemountet und setzt den Korb nur zurück. Ein
  // stehen gebliebener Vorgang blockierte den erzwungenen Reload dauerhaft.
  it('gibt den Bestell-Korb beim Tischwechsel im Vorgangs-Register frei', async () => {
    getTischState.mockReturnValue(stammtisch)
    getTischHistorie.mockReturnValue([])
    const user = userEvent.setup()
    const { rerender, unmount } = renderPage(backend([testProdukt]))

    await screen.findByText('Stammtisch')
    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)

    // Anderer Tisch: nur der :tischId-Param wechselt, TablePage bleibt gemountet.
    testState.tischId = '2'
    rerender(<TablePage />)
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)

    unmount()
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })

  // Der Refetch des Tisch-States läuft erst beim Schließen des Pops.
  it('zeigt nach der Stornierung den Erfolgs-Pop und lädt erst beim Schließen neu', async () => {
    getTischState.mockReturnValue(stammtisch)
    getTischHistorie.mockReturnValue([
      {
        art: 'bestellung',
        id: '00000000-0000-4000-8000-000000000101',
        userId: 1,
        userName: 'Tester',
        tischId: 1,
        positionen: [position(1)],
        gesamtPreisCents: 350,
        kommentar: '',
        aufgenommenAm: '2026-06-18T12:00:00Z',
        stornierbarePositionen: [position(1)],
        umbuchbarePositionen: [],
      },
    ])
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

    await screen.findByText('Stornierung gebucht.')
    expect(getTischState.mock.calls.length).toBe(ladeCalls)

    await user.click(screen.getByRole('status'))
    await waitFor(() => {
      expect(getTischState.mock.calls.length).toBeGreaterThan(ladeCalls)
    })
  })
})
