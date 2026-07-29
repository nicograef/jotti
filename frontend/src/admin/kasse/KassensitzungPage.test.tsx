import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { toast } from 'sonner'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { BackendError } from '@/lib/Backend'
import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'

import type { GeldtransitBuchung } from './Kassensitzung'
import {
  EroeffnenSection,
  KasseAbschliessenSection,
  KassensitzungPage,
} from './KassensitzungPage'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
}))

const { kasseAbschliessen, kassensitzungEroeffnen, geldtransitBuchen } =
  vi.hoisted(() => ({
    kasseAbschliessen: vi
      .fn<
        (cents: number) => Promise<{
          ausfallResteAnzahl: number
          ohneKonfigurationAnzahl: number
        }>
      >()
      .mockResolvedValue({ ausfallResteAnzahl: 0, ohneKonfigurationAnzahl: 0 }),
    kassensitzungEroeffnen: vi
      .fn<(bezeichnung: string, betragCents: number) => Promise<number>>()
      .mockResolvedValue(1),
    geldtransitBuchen: vi
      .fn<
        (
          geldtransitId: string,
          richtung: string,
          betragCents: number,
          kommentar: string,
        ) => Promise<void>
      >()
      .mockResolvedValue(undefined),
  }))

type OffeneKassensitzungMock = {
  zNr: number
  datum: string
  bezeichnung: string
  status: 'offen'
  eroeffnetAm: string
} | null

const offeneKassensitzungState = vi.hoisted(
  (): { isError: boolean; kassensitzung: OffeneKassensitzungMock } => ({
    isError: false,
    kassensitzung: null,
  }),
)

const geldtransitListeState = vi.hoisted(() => ({
  buchungen: [] as GeldtransitBuchung[],
}))

vi.mock('./hooks', () => ({
  kasseBackend: {
    kasseAbschliessen,
    kassensitzungEroeffnen,
    geldtransitBuchen,
  },
  KASSENBESTAND_KEY: 'kassenbestand',
  GELDTRANSIT_LISTE_KEY: 'geldtransit-liste',
  useKassenbestand: () => ({
    kassenbestand: {
      sollBestandCents: 34000,
      anfangsbestandCents: 15000,
      bareinnahmenCents: 17000,
      einlagenCents: 3000,
      entnahmenCents: 1000,
    },
    dataUpdatedAt: 0,
  }),
  useGeldtransitListe: () => ({ buchungen: geldtransitListeState.buchungen }),
  useOffeneKassensitzung: () => ({
    kassensitzung: offeneKassensitzungState.kassensitzung,
    isPending: false,
    isError: offeneKassensitzungState.isError,
    refetch: () => Promise.resolve(),
  }),
}))

interface OffenerTischMock {
  tischId: number
  tischName: string
  saldoCents: number
}

const liveReportingState = vi.hoisted(
  (): { offeneTische: OffenerTischMock[]; offeneSaldiCents: number } => ({
    offeneTische: [],
    offeneSaldiCents: 0,
  }),
)

vi.mock('@/admin/reporting/hooks', () => ({
  useLiveReporting: () => ({
    liveData: {
      offeneTische: liveReportingState.offeneTische,
      offeneSaldiCents: liveReportingState.offeneSaldiCents,
      summary: {
        gesamtUmsatzCents: 12345,
        gesamtStornierungenCents: 300,
        geldtransitCents: 5000,
      },
    },
    isPending: false,
  }),
}))

const tseState = vi.hoisted(() => ({ istKonfiguriert: false }))

vi.mock('@/admin/tse/hooks', () => ({
  useTSEKonfiguration: () => ({
    tseKonfiguration: { istKonfiguriert: tseState.istKonfiguriert },
  }),
}))

function renderPage() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  render(
    <QueryClientProvider client={queryClient}>
      <KassensitzungPage />
    </QueryClientProvider>,
  )
}

beforeEach(() => {
  VorgangsRegisterSingleton.zuruecksetzen()
})

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
  offeneKassensitzungState.isError = false
  offeneKassensitzungState.kassensitzung = null
  geldtransitListeState.buchungen = []
  liveReportingState.offeneTische = []
  liveReportingState.offeneSaldiCents = 0
})

describe('KassensitzungPage', () => {
  it('zeigt bei Query-Fehler einen Fehlerzustand statt des Steppers', () => {
    offeneKassensitzungState.isError = true
    renderPage()

    expect(
      screen.getByText('Kassendaten konnten nicht geladen werden'),
    ).toBeInTheDocument()
    // Der Stepper darf bei einem Fehler nicht erscheinen — die Kasse wirkt sonst
    // fälschlich geschlossen.
    expect(screen.queryByText('2 · Laufender Betrieb')).not.toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Erneut versuchen' }),
    ).toBeInTheDocument()
  })

  it('zeigt im Leerzustand Schritt 1 als aktives Eröffnen-Formular, Schritte 2–3 ausgegraut', () => {
    offeneKassensitzungState.kassensitzung = null
    renderPage()

    // Schritt 1 ist das Eröffnen-Formular.
    expect(
      screen.getByRole('button', { name: 'Kassensitzung eröffnen' }),
    ).toBeInTheDocument()
    // Schritte 2 und 3 sind als ausgegraute Platzhalter da (kein Soll-Bestand,
    // kein Abschluss-Button).
    expect(screen.getByText('2 · Laufender Betrieb')).toBeInTheDocument()
    expect(
      screen.getByText('3 · Am Ende des Tages: Kasse abschließen'),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'Kasse abschließen' }),
    ).not.toBeInTheDocument()
  })

  it('zeigt bei offener Sitzung den Stepper mit Titel, Soll-Bestand-Aufschlüsselung und Bewegungsliste', () => {
    offeneKassensitzungState.kassensitzung = {
      zNr: 12,
      datum: '2026-07-11',
      bezeichnung: 'Sommerfest Tag 2',
      status: 'offen',
      eroeffnetAm: '2026-07-11T08:02:00Z',
    }
    geldtransitListeState.buchungen = [
      {
        zeitpunkt: '2026-07-11T18:15:00Z',
        richtung: 'entnahme',
        betragCents: 150000,
        kommentar: 'Abschöpfung in den Tresor',
        gebuchtVon: 'nico',
      },
      {
        zeitpunkt: '2026-07-11T12:05:00Z',
        richtung: 'einlage',
        betragCents: 20000,
        kommentar: 'Wechselgeld Nachschub',
        gebuchtVon: 'sophie',
      },
    ]
    renderPage()

    // Dynamischer Titel.
    expect(
      screen.getByText('Kassentag Nr. 12 — Sommerfest Tag 2'),
    ).toBeInTheDocument()
    // Soll-Bestand groß und die vier Aufschlüsselungs-Kacheln. 340,00 € steht in
    // Schritt 2 (Soll-Bestand groß) und in der Live-Rechnung von Schritt 3.
    expect(screen.getAllByText('340,00 €').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Anfangsbestand')).toBeInTheDocument()
    expect(screen.getByText('+ Bareinnahmen')).toBeInTheDocument()
    expect(screen.getByText('+ Einlagen')).toBeInTheDocument()
    expect(screen.getByText('− Entnahmen')).toBeInTheDocument()
    // Bewegungsliste mit beiden Buchungen.
    expect(screen.getByText(/Abschöpfung in den Tresor/)).toBeInTheDocument()
    expect(screen.getByText(/Wechselgeld Nachschub/)).toBeInTheDocument()
    // Buttons zum Buchen mit vorbelegter Richtung.
    expect(
      screen.getByRole('button', { name: 'Geld einlegen' }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Geld entnehmen' }),
    ).toBeInTheDocument()
  })

  it('öffnet über „Geld entnehmen" den Dialog mit vorbelegter Richtung und bucht', async () => {
    offeneKassensitzungState.kassensitzung = {
      zNr: 12,
      datum: '2026-07-11',
      bezeichnung: 'Sommerfest Tag 2',
      status: 'offen',
      eroeffnetAm: '2026-07-11T08:02:00Z',
    }
    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByRole('button', { name: 'Geld entnehmen' }))

    // Der Dialog trägt die Entnahme-Richtung im Titel.
    expect(
      screen.getByRole('heading', { name: 'Geld entnehmen' }),
    ).toBeInTheDocument()

    await user.type(screen.getByLabelText('Betrag'), '30,00')
    await user.type(screen.getByLabelText('Kommentar'), 'Getränke-Nachkauf')
    const dialogButtons = screen.getAllByRole('button', {
      name: 'Geld entnehmen',
    })
    await user.click(dialogButtons[dialogButtons.length - 1])

    expect(geldtransitBuchen).toHaveBeenCalledTimes(1)
    // Argumente: (geldtransitId, richtung, betragCents, kommentar). Die Richtung
    // ist durch den Button vorbelegt, der Betrag in Cent, der Kommentar wörtlich.
    const [geldtransitId, richtung, betragCents, kommentar] =
      geldtransitBuchen.mock.calls[0]
    expect(typeof geldtransitId).toBe('string')
    expect(richtung).toBe('entnahme')
    expect(betragCents).toBe(3000)
    expect(kommentar).toBe('Getränke-Nachkauf')
  })
})

describe('EroeffnenSection', () => {
  it('fragt ohne TSE-Konfiguration nach; Abbrechen eröffnet nicht, Bestätigen eröffnet', async () => {
    tseState.istKonfiguriert = false
    const user = userEvent.setup()
    render(<EroeffnenSection onSuccess={vi.fn()} />)

    await user.type(screen.getByLabelText('Bezeichnung'), 'Sommerfest Tag 1')
    await user.type(screen.getByLabelText('Anfangsbestand'), '150,00')
    await user.click(
      screen.getByRole('button', { name: 'Kassensitzung eröffnen' }),
    )

    expect(screen.getByText('Keine TSE konfiguriert')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Abbrechen' }))
    expect(kassensitzungEroeffnen).not.toHaveBeenCalled()

    await user.click(
      screen.getByRole('button', { name: 'Kassensitzung eröffnen' }),
    )
    await user.click(screen.getByRole('button', { name: 'Trotzdem eröffnen' }))
    expect(kassensitzungEroeffnen).toHaveBeenCalledWith(
      'Sommerfest Tag 1',
      15000,
    )
  })

  it('eröffnet mit konfigurierter TSE direkt ohne Dialog', async () => {
    tseState.istKonfiguriert = true
    const user = userEvent.setup()
    render(<EroeffnenSection onSuccess={vi.fn()} />)

    await user.type(screen.getByLabelText('Bezeichnung'), 'Sommerfest Tag 1')
    await user.type(screen.getByLabelText('Anfangsbestand'), '150,00')
    await user.click(
      screen.getByRole('button', { name: 'Kassensitzung eröffnen' }),
    )

    expect(screen.queryByText('Keine TSE konfiguriert')).not.toBeInTheDocument()
    expect(kassensitzungEroeffnen).toHaveBeenCalledWith(
      'Sommerfest Tag 1',
      15000,
    )
  })

  it('eröffnet mit 0 € Anfangsbestand (kein Wechselgeld)', async () => {
    tseState.istKonfiguriert = true
    const user = userEvent.setup()
    render(<EroeffnenSection onSuccess={vi.fn()} />)

    await user.type(screen.getByLabelText('Bezeichnung'), 'Sommerfest')
    await user.type(screen.getByLabelText('Anfangsbestand'), '0,00')
    await user.click(
      screen.getByRole('button', { name: 'Kassensitzung eröffnen' }),
    )

    expect(kassensitzungEroeffnen).toHaveBeenCalledWith('Sommerfest', 0)
  })

  it('akzeptiert Standardwert 0 € (leeres Betrag-Feld) ohne Validierungsfehler', async () => {
    // 0 € Anfangsbestand ist gültig — leeres EuroInput-Feld ergibt den Standardwert 0.
    // Negativwerte kann EuroInput strukturell nicht erzeugen; deren Schema-Absicherung
    // wird direkt in KasseBackend.test.ts geprüft.
    tseState.istKonfiguriert = true
    const user = userEvent.setup()
    render(<EroeffnenSection onSuccess={vi.fn()} />)

    await user.type(screen.getByLabelText('Bezeichnung'), 'Sommerfest')
    // Kein Betrag eingetragen — Formular-Standardwert ist 0, der jetzt gültig ist.
    await user.click(
      screen.getByRole('button', { name: 'Kassensitzung eröffnen' }),
    )

    // Mit dem Standardwert 0 soll kein Validierungsfehler erscheinen.
    expect(
      screen.queryByText('Betrag muss mindestens 0 Cent sein.'),
    ).not.toBeInTheDocument()
    expect(kassensitzungEroeffnen).toHaveBeenCalledWith('Sommerfest', 0)
  })
})

describe('KasseAbschliessenSection', () => {
  it('rechnet die Differenz live als Ist − Soll, Fehlbetrag negativ in Rot', async () => {
    const user = userEvent.setup()
    render(<KasseAbschliessenSection kassensitzungNr={1} onSuccess={vi.fn()} />)

    // Soll ist 340,00 € (aus dem Kassenbestand-Mock). Ohne Eingabe: Gezählt
    // 0,00 €, Differenz −340,00 € (Ist − Soll) als kompletter Fehlbetrag, rot.
    expect(screen.getByText('340,00 €')).toBeInTheDocument()
    expect(screen.getByText('0,00 €')).toBeInTheDocument()
    const leerDifferenz = screen.getByText('-340,00 €')
    expect(leerDifferenz).toHaveClass('text-destructive')

    // 337,50 € gezählt → Differenz −2,50 € (Fehlbetrag, fehlendes Geld), rot.
    await user.type(screen.getByLabelText('Gezählter Ist-Bestand'), '337,50')
    const fehlbetrag = screen.getByText('-2,50 €')
    expect(fehlbetrag).toBeInTheDocument()
    expect(fehlbetrag).toHaveClass('text-destructive')
  })

  it('färbt einen Überschuss (Ist > Soll) nicht rot', async () => {
    const user = userEvent.setup()
    render(<KasseAbschliessenSection kassensitzungNr={1} onSuccess={vi.fn()} />)

    // 342,50 € gezählt bei Soll 340,00 € → Differenz +2,50 € (Überschuss),
    // nicht rot.
    await user.type(screen.getByLabelText('Gezählter Ist-Bestand'), '342,50')
    // Überschuss trägt das Plus-Vorzeichen (+2,50 €), bleibt aber ohne Rot.
    const ueberschuss = screen.getByText('+2,50 €')
    expect(ueberschuss).toBeInTheDocument()
    expect(ueberschuss).not.toHaveClass('text-destructive')
  })

  it('warnt bei offenen Tischen mit Anzahl und Betrag, ohne offene Tische fehlt die Warnung', () => {
    liveReportingState.offeneTische = [
      { tischId: 1, tischName: 'Tisch 1', saldoCents: 25000 },
      { tischId: 2, tischName: 'Tisch 2', saldoCents: 16200 },
    ]
    liveReportingState.offeneSaldiCents = 41200
    render(<KasseAbschliessenSection kassensitzungNr={1} onSuccess={vi.fn()} />)

    expect(
      screen.getByText('2 Tische sind noch offen (412,00 €).'),
    ).toBeInTheDocument()
  })

  it('zeigt ohne offene Tische keine Warnung', () => {
    render(<KasseAbschliessenSection kassensitzungNr={1} onSuccess={vi.fn()} />)

    expect(screen.queryByText(/noch offen/)).not.toBeInTheDocument()
  })

  it('stellt Soll, Ist und Differenz im Bestätigungsdialog gegenüber', async () => {
    const user = userEvent.setup()
    render(<KasseAbschliessenSection kassensitzungNr={1} onSuccess={vi.fn()} />)

    await user.type(screen.getByLabelText('Gezählter Ist-Bestand'), '342,50')
    await user.click(
      screen.getByRole('button', { name: 'Kasse endgültig abschließen…' }),
    )

    expect(screen.getByText('Kasse abschließen?')).toBeInTheDocument()
    // Ist-Bestand steht in der Live-Rechnung (Gezählt) und im Dialog (Ist-Bestand
    // gezählt) — also mindestens zweimal.
    expect(screen.getAllByText('342,50 €').length).toBeGreaterThanOrEqual(2)
    // Der Bestätigungs-Button trägt weiterhin „Kasse abschließen".
    expect(
      screen.getByRole('button', { name: 'Kasse abschließen' }),
    ).toBeInTheDocument()
  })

  it('bucht den Abschluss mit dem gezählten Ist-Bestand in Cent', async () => {
    const user = userEvent.setup()
    render(<KasseAbschliessenSection kassensitzungNr={1} onSuccess={vi.fn()} />)

    await user.type(screen.getByLabelText('Gezählter Ist-Bestand'), '342,50')
    await user.click(
      screen.getByRole('button', { name: 'Kasse endgültig abschließen…' }),
    )
    await user.click(screen.getByRole('button', { name: 'Kasse abschließen' }))

    expect(kasseAbschliessen).toHaveBeenCalledWith(34250)
  })

  it('übernimmt die Zählhilfe-Summe in das Ist-Bestand-Feld', async () => {
    const user = userEvent.setup()
    render(<KasseAbschliessenSection kassensitzungNr={1} onSuccess={vi.fn()} />)

    await user.click(screen.getByRole('button', { name: /Zählhilfe öffnen/ }))
    // 3×100 € (30000) + 2×20 € (4000) = 34000 → 340,00 €.
    await user.type(screen.getByLabelText('100 €'), '3')
    await user.type(screen.getByLabelText('20 €'), '2')
    await user.click(screen.getByRole('button', { name: 'Übernehmen' }))

    expect(screen.getByLabelText('Gezählter Ist-Bestand')).toHaveValue('340,00')
  })

  it('weist Ausfall-Reste in der Erfolgsmeldung aus', async () => {
    kasseAbschliessen.mockResolvedValueOnce({
      ausfallResteAnzahl: 2,
      ohneKonfigurationAnzahl: 1,
    })
    const user = userEvent.setup()
    render(<KasseAbschliessenSection kassensitzungNr={1} onSuccess={vi.fn()} />)

    await user.type(screen.getByLabelText('Gezählter Ist-Bestand'), '342,50')
    await user.click(
      screen.getByRole('button', { name: 'Kasse endgültig abschließen…' }),
    )
    await user.click(screen.getByRole('button', { name: 'Kasse abschließen' }))

    expect(toast.success).toHaveBeenCalledWith(
      expect.stringContaining('nachsigniert'),
    )
    expect(toast.success).toHaveBeenCalledWith(
      expect.stringContaining('keine TSE konfiguriert'),
    )
  })

  it('zeigt bei ausstehenden Signaturen eine Meldung und lässt den Abschluss erneut anfordern', async () => {
    kasseAbschliessen.mockRejectedValueOnce(
      new BackendError(409, 'signaturen_ausstehend', {
        anzahl: 2,
      }),
    )
    const user = userEvent.setup()
    render(<KasseAbschliessenSection kassensitzungNr={1} onSuccess={vi.fn()} />)

    await user.type(screen.getByLabelText('Gezählter Ist-Bestand'), '342,50')
    await user.click(
      screen.getByRole('button', { name: 'Kasse endgültig abschließen…' }),
    )
    await user.click(screen.getByRole('button', { name: 'Kasse abschließen' }))

    expect(toast.warning).toHaveBeenCalledWith(
      expect.stringContaining('2 Vorgänge sind noch nicht signiert'),
    )
    // Dialog bleibt offen: der Abschluss kann erneut angefordert werden.
    expect(screen.getByText('Kasse abschließen?')).toBeInTheDocument()
  })
})

describe('GeldtransitDialog im Vorgangs-Register', () => {
  it('meldet das angefangene Formular und gibt es beim Schließen frei', async () => {
    offeneKassensitzungState.kassensitzung = {
      zNr: 12,
      datum: '2026-07-11',
      bezeichnung: 'Sommerfest Tag 2',
      status: 'offen',
      eroeffnetAm: '2026-07-11T08:02:00Z',
    }
    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByRole('button', { name: 'Geld einlegen' }))
    // Das frisch geöffnete, leere Formular ist noch kein offener Vorgang.
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)

    await user.type(screen.getByLabelText('Kommentar'), 'Wechselgeld')
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)

    // Beim Schließen bleiben die Werte im Formular stehen, werden aber beim
    // nächsten Öffnen verworfen — der Vorgang ist damit erledigt.
    await user.click(screen.getByRole('button', { name: 'Abbrechen' }))
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })
})

describe('KasseAbschliessenSection im Vorgangs-Register', () => {
  it('meldet den eingetippten Ist-Bestand samt offener Rückfrage als einen Vorgang', async () => {
    const user = userEvent.setup()
    render(<KasseAbschliessenSection kassensitzungNr={1} onSuccess={vi.fn()} />)

    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)

    await user.type(screen.getByLabelText('Gezählter Ist-Bestand'), '342,50')
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)

    await user.click(
      screen.getByRole('button', { name: 'Kasse endgültig abschließen…' }),
    )
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)

    // Abbrechen schließt nur die Rückfrage; der gezählte Betrag steht weiter da.
    await user.click(screen.getByRole('button', { name: 'Abbrechen' }))
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)
  })

  it('gibt den Vorgang frei, sobald der Abschluss gebucht ist', async () => {
    const user = userEvent.setup()
    render(<KasseAbschliessenSection kassensitzungNr={1} onSuccess={vi.fn()} />)

    await user.type(screen.getByLabelText('Gezählter Ist-Bestand'), '342,50')
    await user.click(
      screen.getByRole('button', { name: 'Kasse endgültig abschließen…' }),
    )
    await user.click(screen.getByRole('button', { name: 'Kasse abschließen' }))

    expect(kasseAbschliessen).toHaveBeenCalledWith(34250)
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })
})
