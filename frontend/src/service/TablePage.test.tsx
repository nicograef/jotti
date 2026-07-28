import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  act,
  cleanup,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import userEvent, { type UserEvent } from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { BackendError } from '@/lib/Backend'

import { AKTIVE_PRODUKTE_KEY } from './product/hooks'
import type { Produkt } from './product/Produkt'
import type { BestellungAufnehmen, Position } from './table/Bestellung'
import {
  AKTIVE_TISCHE_MIT_FAVORITEN_KEY,
  EIGENE_UEBERSICHT_KEY,
  MEINE_TISCHE_STATE_KEY,
  TISCH_HISTORIE_KEY,
  TISCH_STATE_KEY,
} from './table/hooks'
import type { Tisch, TischSession } from './table/Tisch'
import type { ZahlungKassieren } from './table/Zahlung'
import { TablePage } from './TablePage'

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

// Steuerbarer Testzustand: `tischId` bildet den :tischId-Param nach
// (Tischwechsel ohne Remount).
const testState = vi.hoisted(() => ({
  tischId: '1',
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

// Nur der Singleton wird ersetzt: Die Fehlerklassen bleiben echt, weil die
// Fehlermeldungen der Aktionen (use-action-submit) gegen sie prüfen.
vi.mock('@/lib/Backend', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/Backend')>()),
  BackendSingleton: {},
}))

// Die eigene Servicekraft (für die „Meine Positionen"-Filterung in Zahlung);
// canCancel/canRebook, damit der Storno-/Umbuchen-Pfad der Historie greift.
vi.mock('@/lib/Auth', () => ({
  AuthSingleton: { userId: 1, canCancel: true, canRebook: true },
}))

// Die Seite läuft gegen die echten Query-Hooks; nur die Backends sind ersetzt.
// Nur so ist prüfbar, dass ein gescheitertes Erstladen (leerer Cache) und ein
// gescheiterter Hintergrund-Refetch (gefüllter Cache) verschieden aussehen.
const {
  getTischState,
  getTischHistorie,
  getAktiveTische,
  getAktiveProdukte,
  stornierungErteilen,
  bestellungUmbuchen,
  bestellungAufnehmen,
  zahlungKassieren,
} = vi.hoisted(() => ({
  getTischState: vi.fn<() => Promise<TischSession>>(),
  getTischHistorie: vi.fn<() => Promise<unknown[]>>(),
  getAktiveTische: vi.fn<() => Promise<Tisch[]>>(),
  getAktiveProdukte: vi.fn<() => Promise<Produkt[]>>(),
  stornierungErteilen: vi.fn<() => Promise<void>>(),
  bestellungUmbuchen: vi.fn<() => Promise<void>>(),
  bestellungAufnehmen: vi.fn<(b: BestellungAufnehmen) => Promise<void>>(),
  zahlungKassieren: vi.fn<(z: ZahlungKassieren) => Promise<void>>(),
}))

vi.mock('./product/ProduktBackend', () => ({
  ProduktBackend: class {
    getAktiveProdukte = getAktiveProdukte
  },
}))

vi.mock('./table/TischBackend', () => ({
  TischBackend: class {
    getTischState = getTischState
    getTischHistorie = getTischHistorie
    getAktiveTische = getAktiveTische
    stornierungErteilen = stornierungErteilen
    bestellungUmbuchen = bestellungUmbuchen
    bestellungAufnehmen = bestellungAufnehmen
    zahlungKassieren = zahlungKassieren
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
  getAktiveProdukte.mockResolvedValue([])
})

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
  testState.tischId = '1'
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
  return { queryClient }
}

// Historien-Eintrag mit genau einer stornierbaren Position.
const bestellungMitPosition = {
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
}

// Umbuchung über die Historie: Bestellung öffnen, Position wählen, Ziel-Tisch
// wählen, buchen. Wie beim Storno steht danach der Erfolgs-Pop offen.
async function bucheUeberHistorieUm(user: UserEvent) {
  await user.click(screen.getByRole('tab', { name: 'Historie' }))
  await user.click(screen.getByRole('button', { name: /Bestellung/ }))
  await user.click(screen.getByRole('button', { name: 'Umbuchen' }))
  await user.click(screen.getByRole('button', { name: /hinzufügen/ }))
  await user.selectOptions(screen.getByRole('combobox'), '2')
  await user.click(screen.getByRole('button', { name: 'Umbuchung ausführen' }))
}

// Bestellen auf dem Handy-Pfad: Dock-Aktionsbutton öffnet den Drawer, darin
// liegt der Aufnehmen-Button. Danach wird der Drawer geschlossen, damit der
// nächste Schritt (Tab-Wechsel) nicht am modalen Overlay hängen bleibt.
async function bestelleUeberDenDrawer(user: UserEvent) {
  await user.click(
    screen.getByRole('button', { name: /Bestellung überprüfen/ }),
  )
  const dialog = await screen.findByRole('dialog')
  await user.click(
    within(dialog).getByRole('button', { name: 'Bestellung aufnehmen' }),
  )
  await user.click(within(dialog).getByRole('button', { name: 'Abbrechen' }))
  await waitFor(() => {
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })
}

// Kassieren auf dem Handy-Pfad, analog zum Bestellen. Der Dock-Trigger trägt
// zusätzlich Anzahl und Summe im Namen, der Submit-Button im Drawer nur
// „Kassieren".
async function kassiereUeberDenDrawer(user: UserEvent) {
  await user.click(screen.getByRole('button', { name: /Kassieren/ }))
  const dialog = await screen.findByRole('dialog')
  await user.click(within(dialog).getByRole('button', { name: 'Kassieren' }))
  await user.click(within(dialog).getByRole('button', { name: 'Abbrechen' }))
  await waitFor(() => {
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })
}

// Storno über die Historie: Bestellung öffnen, Position wählen, Kommentar
// setzen, buchen. Danach steht der Erfolgs-Pop offen — der Refetch läuft erst
// beim Schließen.
async function storniereUeberHistorie(user: UserEvent) {
  await user.click(screen.getByRole('tab', { name: 'Historie' }))
  await user.click(screen.getByRole('button', { name: /Bestellung/ }))
  await user.click(screen.getByRole('button', { name: /Stornieren…/ }))
  await user.click(screen.getByRole('button', { name: /hinzufügen/ }))
  await user.type(
    screen.getByPlaceholderText('Kommentar (erforderlich)'),
    'Falsch gebucht',
  )
  await user.click(screen.getByRole('button', { name: 'Stornierung erteilen' }))
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

  it('zeigt im Bestellen-Tab einen Ladefehler statt einer leeren Produktliste', async () => {
    getAktiveProdukte.mockRejectedValue(new Error('Netzabbruch'))
    getTischState.mockResolvedValue(stammtisch)
    getTischHistorie.mockResolvedValue([])
    renderPage()

    expect(
      await screen.findByText('Produkte konnten nicht geladen werden'),
    ).toBeInTheDocument()
    // Der Leerzustand der Produktliste behauptet, es gebe nichts zu bestellen.
    expect(
      screen.queryByText('Keine Produkte verfügbar'),
    ).not.toBeInTheDocument()
  })

  // Der Ladefehler gilt nur dem gescheiterten Erstladen. Scheitert ein
  // Hintergrund-Refetch, sind die zuletzt geladenen Daten weiter gültig — sie
  // wegzureißen nähme der Servicekraft mitten im Betrieb den offenen Tisch.
  // Die Meldung trägt der zentrale Fehler-Toast aus queryClient.ts.
  it('lässt den Tischzustand stehen, wenn ein Hintergrund-Refetch scheitert', async () => {
    getTischState
      .mockResolvedValueOnce(stammtisch)
      .mockRejectedValue(new Error('Netzabbruch'))
    getTischHistorie.mockResolvedValue([])
    const { queryClient } = renderPage()

    await screen.findByText('Stammtisch')
    await act(async () => {
      await queryClient.refetchQueries()
    })
    // Die Query steht auf „error", die Seite hat die Aktualisierung verarbeitet
    // — erst danach ist die Anzeige aussagekräftig.
    await waitFor(() => {
      expect(queryClient.getQueryState([TISCH_STATE_KEY, 1])?.status).toBe(
        'error',
      )
    })

    expect(getTischState).toHaveBeenCalledTimes(2)
    expect(screen.getByText('Stammtisch')).toBeInTheDocument()
    expect(
      screen.queryByText('Tischdaten konnten nicht geladen werden'),
    ).not.toBeInTheDocument()
  })

  it('lässt die Produktliste stehen, wenn ein Hintergrund-Refetch scheitert', async () => {
    getAktiveProdukte
      .mockResolvedValueOnce([testProdukt])
      .mockRejectedValue(new Error('Netzabbruch'))
    getTischState.mockResolvedValue(stammtisch)
    getTischHistorie.mockResolvedValue([])
    const { queryClient } = renderPage()

    await screen.findByText('Bratwurst')
    await act(async () => {
      await queryClient.refetchQueries()
    })
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
    getAktiveProdukte.mockResolvedValue([testProdukt])
    getTischState.mockResolvedValue(stammtisch)
    getTischHistorie.mockResolvedValue([])
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('Stammtisch')
    // Bestellen ist der Default-Tab: eine Variante in den Korb legen.
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

  // Der Idempotenz-Schlüssel muss dieselbe Lebensdauer haben wie die Auswahl,
  // zu der er gehört. Lag er im Tab-Inhalt, gab ihn das Aus- und
  // Wiedereinhängen der Radix-Tabs neu aus: Der Helfer bucht, die Antwort geht
  // verloren, er prüft auf der Historie nach und bucht erneut — mit neuem
  // Schlüssel bucht der Server ein zweites Mal.
  it('behält die bestellungId über einen Tab-Wechsel hinweg', async () => {
    getAktiveProdukte.mockResolvedValue([testProdukt])
    getTischState.mockResolvedValue(stammtisch)
    getTischHistorie.mockResolvedValue([])
    // Die verlorene Antwort: Der Korb bleibt gefüllt, weil nur der Erfolg ihn
    // leert.
    bestellungAufnehmen.mockRejectedValue(new Error('Netzabbruch'))
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('Stammtisch')
    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    await bestelleUeberDenDrawer(user)
    await waitFor(() => {
      expect(bestellungAufnehmen).toHaveBeenCalledTimes(1)
    })
    const ersterKey = bestellungAufnehmen.mock.calls[0][0].bestellungId

    // Auf die Historie schauen, ob die Bestellung angekommen ist, und zurück.
    await user.click(screen.getByRole('tab', { name: 'Historie' }))
    await user.click(screen.getByRole('tab', { name: 'Bestellen' }))

    await bestelleUeberDenDrawer(user)
    await waitFor(() => {
      expect(bestellungAufnehmen).toHaveBeenCalledTimes(2)
    })
    const zweiterAufruf = bestellungAufnehmen.mock.calls[1][0]
    expect(zweiterAufruf.bestellungId).toBe(ersterKey)
    expect(zweiterAufruf.positionen).toEqual(
      bestellungAufnehmen.mock.calls[0][0].positionen,
    )
  })

  // Ein Admin kann ein Produkt deaktivieren, während es im Korb liegt. Bliebe
  // der Korb-Eintrag stehen, wäre er unsichtbar (die Zeile ist weg) und nicht
  // mehr herunterzählbar; der Korb gälte nie wieder als leer, und die
  // bestellungId rotierte für die restliche Lebensdauer der Seite nicht mehr —
  // die nächste Bestellung liefe unter dem Schlüssel der vorherigen.
  it('vergibt eine neue bestellungId, wenn die gewählte Variante aus der Produktliste fällt', async () => {
    getAktiveProdukte
      .mockResolvedValueOnce([testProdukt, testBeilage])
      .mockResolvedValue([testBeilage])
    getTischState.mockResolvedValue(stammtisch)
    getTischHistorie.mockResolvedValue([])
    // Die verlorene Antwort: Der Korb bleibt gefüllt, weil nur der Erfolg ihn
    // leert.
    bestellungAufnehmen.mockRejectedValue(new Error('Netzabbruch'))
    const user = userEvent.setup()
    const { queryClient } = renderPage()

    await screen.findByText('Bratwurst')
    await user.click(
      screen.getAllByRole('button', { name: 'Variante hinzufügen' })[0],
    )
    await bestelleUeberDenDrawer(user)
    await waitFor(() => {
      expect(bestellungAufnehmen).toHaveBeenCalledTimes(1)
    })
    const ersterKey = bestellungAufnehmen.mock.calls[0][0].bestellungId

    // Die Bratwurst wird deaktiviert; der nächste Abruf liefert sie nicht mehr.
    await act(async () => {
      await queryClient.refetchQueries({ queryKey: [AKTIVE_PRODUKTE_KEY] })
    })
    await waitFor(() => {
      expect(screen.queryByText('Bratwurst')).not.toBeInTheDocument()
    })
    // Der Korb ist wirklich leer, nicht nur unsichtbar leer.
    expect(
      screen.getByRole('button', { name: /Bestellung überprüfen/ }),
    ).toBeDisabled()

    // Neue Zusammenstellung: eigener Vorgang, eigener Schlüssel.
    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    await bestelleUeberDenDrawer(user)
    await waitFor(() => {
      expect(bestellungAufnehmen).toHaveBeenCalledTimes(2)
    })
    const zweiterAufruf = bestellungAufnehmen.mock.calls[1][0]
    expect(zweiterAufruf.positionen).toEqual([
      { produktId: 2, varianteId: 2, menge: 1 },
    ])
    expect(zweiterAufruf.bestellungId).not.toBe(ersterKey)
  })

  // Der 409 `vorgang_daten_abweichend` belegt, dass der Vorgang unter diesem
  // Schlüssel gebucht ist — nur seine Antwort ging verloren. Bliebe der Korb
  // stehen, bliebe auch der Schlüssel stehen, und die Meldung („nur die
  // Differenz erneut erfassen") führte in genau denselben 409 zurück: Die
  // ergänzte Position wäre nie zu buchen.
  it('räumt Korb, Schlüssel und Tischzustand ab, wenn der Server den Vorgang als gebucht meldet', async () => {
    getAktiveProdukte.mockResolvedValue([testProdukt, testBeilage])
    getTischState.mockResolvedValue(stammtisch)
    getTischHistorie.mockResolvedValue([])
    bestellungAufnehmen
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
    await user.click(
      screen.getByRole('button', { name: /Bestellung überprüfen/ }),
    )
    const drawer = await screen.findByRole('dialog')
    // Vor dem Absenden gezählt: Das Abräumen läuft synchron im Fehlerpfad, der
    // Refetch wäre sonst schon gelaufen, bevor der Zähler steht.
    const tischAbrufe = getTischState.mock.calls.length
    await user.click(
      within(drawer).getByRole('button', { name: 'Bestellung aufnehmen' }),
    )
    await waitFor(() => {
      expect(bestellungAufnehmen).toHaveBeenCalledTimes(1)
    })
    const ersterKey = bestellungAufnehmen.mock.calls[0][0].bestellungId

    // Der abgeschlossene Vorgang schließt den Drawer und leert den Korb.
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })
    expect(
      screen.getByRole('button', { name: /Bestellung überprüfen/ }),
    ).toBeDisabled()
    // Der Tischzustand lädt neu — nur so sieht der Helfer, was tatsächlich
    // gebucht ist, und weiß, was die Differenz ist.
    await waitFor(() => {
      expect(getTischState.mock.calls.length).toBeGreaterThan(tischAbrufe)
    })

    // Die Differenz nachbuchen: eigene Zusammenstellung, eigener Schlüssel.
    await user.click(
      screen.getAllByRole('button', { name: 'Variante hinzufügen' })[1],
    )
    await user.click(
      screen.getByRole('button', { name: /Bestellung überprüfen/ }),
    )
    const zweiterDrawer = await screen.findByRole('dialog')
    await user.click(
      within(zweiterDrawer).getByRole('button', {
        name: 'Bestellung aufnehmen',
      }),
    )
    await waitFor(() => {
      expect(bestellungAufnehmen).toHaveBeenCalledTimes(2)
    })

    const zweiterAufruf = bestellungAufnehmen.mock.calls[1][0]
    expect(zweiterAufruf.positionen).toEqual([
      { produktId: 2, varianteId: 2, menge: 1 },
    ])
    expect(zweiterAufruf.bestellungId).not.toBe(ersterKey)
  })

  it('behält die vorgangId der Zahlung über einen Tab-Wechsel hinweg', async () => {
    getTischState.mockResolvedValue({
      ...stammtisch,
      unbezahltePositionen: [position('p1')],
    })
    getTischHistorie.mockResolvedValue([])
    zahlungKassieren.mockRejectedValue(new Error('Netzabbruch'))
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('Stammtisch')
    await user.click(screen.getByRole('tab', { name: 'Kassieren' }))
    await user.click(screen.getByRole('button', { name: 'Produkt hinzufügen' }))
    await kassiereUeberDenDrawer(user)
    await waitFor(() => {
      expect(zahlungKassieren).toHaveBeenCalledTimes(1)
    })
    const ersterKey = zahlungKassieren.mock.calls[0][0].vorgangId

    await user.click(screen.getByRole('tab', { name: 'Historie' }))
    await user.click(screen.getByRole('tab', { name: 'Kassieren' }))

    await kassiereUeberDenDrawer(user)
    await waitFor(() => {
      expect(zahlungKassieren).toHaveBeenCalledTimes(2)
    })
    expect(zahlungKassieren.mock.calls[1][0].vorgangId).toBe(ersterKey)
  })

  // Dieselbe Sackgasse auf dem Kassieren-Pfad: Ohne geleerte Auswahl bliebe der
  // Schlüssel stehen, und die verbliebene Position wäre nie zu kassieren.
  it('räumt Kassieren-Auswahl und Schlüssel ab, wenn der Server den Vorgang als gebucht meldet', async () => {
    getTischState.mockResolvedValue({
      ...stammtisch,
      unbezahltePositionen: [position('p1'), position('p2')],
    })
    getTischHistorie.mockResolvedValue([])
    zahlungKassieren
      .mockRejectedValueOnce(
        new BackendError(409, 'vorgang_daten_abweichend', ''),
      )
      .mockRejectedValue(new Error('Netzabbruch'))
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('Stammtisch')
    await user.click(screen.getByRole('tab', { name: 'Kassieren' }))
    await user.click(
      screen.getAllByRole('button', { name: 'Produkt hinzufügen' })[0],
    )
    await user.click(screen.getByRole('button', { name: /Kassieren/ }))
    const drawer = await screen.findByRole('dialog')
    await user.click(within(drawer).getByRole('button', { name: 'Kassieren' }))
    await waitFor(() => {
      expect(zahlungKassieren).toHaveBeenCalledTimes(1)
    })
    const ersterKey = zahlungKassieren.mock.calls[0][0].vorgangId

    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })
    expect(screen.getByRole('button', { name: /Kassieren/ })).toBeDisabled()

    await user.click(
      screen.getAllByRole('button', { name: 'Produkt hinzufügen' })[1],
    )
    await user.click(screen.getByRole('button', { name: /Kassieren/ }))
    const zweiterDrawer = await screen.findByRole('dialog')
    await user.click(
      within(zweiterDrawer).getByRole('button', { name: 'Kassieren' }),
    )
    await waitFor(() => {
      expect(zahlungKassieren).toHaveBeenCalledTimes(2)
    })

    const zweiterAufruf = zahlungKassieren.mock.calls[1][0]
    expect(zweiterAufruf.positionen).toEqual([{ positionId: 'p2', menge: 1 }])
    expect(zweiterAufruf.vorgangId).not.toBe(ersterKey)
  })

  // Die Historie ist eine eigene Query mit eigenem Schlüssel. Scheitert nur
  // sie, sind Bestellen und Kassieren nicht betroffen — der Tischzustand ist
  // geladen.
  it('ersetzt bei einem Ladefehler der Historie nur den Historie-Tab', async () => {
    getAktiveProdukte.mockResolvedValue([testProdukt])
    getTischState.mockResolvedValue(stammtisch)
    getTischHistorie.mockRejectedValue(new Error('Netzabbruch'))
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('Stammtisch')
    expect(
      screen.queryByText('Tischdaten konnten nicht geladen werden'),
    ).not.toBeInTheDocument()

    // Bestellen bleibt bedienbar.
    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    expect(
      screen.getByRole('button', { name: /Bestellung überprüfen/ }),
    ).toBeEnabled()

    // Der Ladefehler steht im Historie-Tab.
    await user.click(screen.getByRole('tab', { name: 'Historie' }))
    expect(
      await screen.findByText('Historie konnte nicht geladen werden'),
    ).toBeInTheDocument()
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
    await storniereUeberHistorie(user)
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
    getAktiveProdukte.mockResolvedValue([testProdukt])
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

  // A2: Eine Stornierung bestätigt über den Erfolgs-Pop; der Refetch des
  // Tisch-States läuft erst beim Schließen des Pops, nicht schon beim Erfolg.
  it('zeigt nach der Stornierung den Erfolgs-Pop und lädt erst beim Schließen neu', async () => {
    getTischState.mockResolvedValue(stammtisch)
    getTischHistorie.mockResolvedValue([bestellungMitPosition])
    stornierungErteilen.mockResolvedValue(undefined)
    const user = userEvent.setup()
    renderPage()

    await screen.findByText('Stammtisch')
    const ladeCalls = getTischState.mock.calls.length

    await storniereUeberHistorie(user)

    // Der Pop erscheint; bis zum Schließen läuft kein Refetch des Tisch-States.
    await screen.findByText('Stornierung gebucht.')
    expect(getTischState.mock.calls.length).toBe(ladeCalls)

    // Pop schließen (Tap auf das Overlay) → jetzt lädt der Tisch-State neu.
    await user.click(screen.getByRole('status'))
    await waitFor(() => {
      expect(getTischState.mock.calls.length).toBeGreaterThan(ladeCalls)
    })
  })

  // Die drei Queries der Tischübersicht hängen an keiner Komponente dieser
  // Seite; ohne Invalidierung liefert ihr Cache bei der Rückkehr innerhalb der
  // Aktualitätsschwelle die Werte von vor der Buchung — ein soeben kassierter
  // Tisch stünde dort weiter unter „Noch offen".
  it('invalidiert nach einer Buchung die Queries der Tischübersicht', async () => {
    getTischState.mockResolvedValue(stammtisch)
    getTischHistorie.mockResolvedValue([bestellungMitPosition])
    stornierungErteilen.mockResolvedValue(undefined)
    const user = userEvent.setup()
    const { queryClient } = renderPage()
    const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries')

    await screen.findByText('Stammtisch')
    await storniereUeberHistorie(user)
    await screen.findByText('Stornierung gebucht.')

    // Wie der Refetch läuft auch die Invalidierung erst beim Schließen des Pops.
    await user.click(screen.getByRole('status'))

    await waitFor(() => {
      expect(invalidateQueries).toHaveBeenCalledWith({
        queryKey: [MEINE_TISCHE_STATE_KEY],
      })
    })
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: [AKTIVE_TISCHE_MIT_FAVORITEN_KEY],
    })
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: [EIGENE_UEBERSICHT_KEY],
    })
  })

  // Eine Umbuchung ändert auch den Ziel-Tisch. Dessen Queries sind hier nicht
  // gemountet; ohne Invalidierung über das Präfix zeigte er beim Öffnen den
  // Cache-Stand von vor der Umbuchung — die umgebuchten Positionen fehlten.
  it('markiert nach einer Umbuchung auch den Ziel-Tisch als veraltet', async () => {
    getTischState.mockResolvedValue(stammtisch)
    getTischHistorie.mockResolvedValue([
      { ...bestellungMitPosition, umbuchbarePositionen: [position('p1')] },
    ])
    getAktiveTische.mockResolvedValue([
      { id: 1, name: 'Stammtisch', saldoCents: 1250 },
      { id: 2, name: 'Nebentisch', saldoCents: 0 },
    ])
    bestellungUmbuchen.mockResolvedValue(undefined)
    const user = userEvent.setup()
    const { queryClient } = renderPage()

    // Der Ziel-Tisch liegt im Cache, ohne dass eine Komponente an ihm hängt.
    queryClient.setQueryData([TISCH_STATE_KEY, 2], {
      ...stammtisch,
      tischId: 2,
      tischName: 'Nebentisch',
    })
    queryClient.setQueryData([TISCH_HISTORIE_KEY, 2], [])

    await screen.findByText('Stammtisch')
    await bucheUeberHistorieUm(user)
    await screen.findByText('Auf Nebentisch umgebucht.')
    await user.click(screen.getByRole('status'))

    await waitFor(() => {
      expect(
        queryClient.getQueryState([TISCH_STATE_KEY, 2])?.isInvalidated,
      ).toBe(true)
    })
    expect(
      queryClient.getQueryState([TISCH_HISTORIE_KEY, 2])?.isInvalidated,
    ).toBe(true)
  })
})
