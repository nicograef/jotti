import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { toast } from 'sonner'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { BackendError } from '@/lib/Backend'

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

vi.mock('@/admin/reporting/hooks', () => ({
  useLiveReporting: () => ({
    liveData: {
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

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
  offeneKassensitzungState.isError = false
  offeneKassensitzungState.kassensitzung = null
  geldtransitListeState.buchungen = []
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
    // Soll-Bestand groß und die vier Aufschlüsselungs-Kacheln.
    expect(screen.getByText('340,00 €')).toBeInTheDocument()
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
  it('nimmt den Ist-Bestand auf und stellt Soll, Ist und Differenz im Dialog gegenüber', async () => {
    const user = userEvent.setup()
    render(<KasseAbschliessenSection kassensitzungNr={1} onSuccess={vi.fn()} />)

    const istInput = screen.getByLabelText('Gezählter Ist-Bestand')
    await user.type(istInput, '342,50')
    expect(istInput).toHaveValue('342,50')

    await user.click(screen.getByRole('button', { name: 'Kasse abschließen' }))

    expect(screen.getByText('Kasse abschließen?')).toBeInTheDocument()
    expect(screen.getByText('340,00 €')).toBeInTheDocument() // Soll
    expect(screen.getByText('342,50 €')).toBeInTheDocument() // Ist
    expect(screen.getByText('-2,50 €')).toBeInTheDocument() // Differenz (Soll − Ist)
  })

  it('bucht den Abschluss mit dem gezählten Ist-Bestand in Cent', async () => {
    const user = userEvent.setup()
    render(<KasseAbschliessenSection kassensitzungNr={1} onSuccess={vi.fn()} />)

    await user.type(screen.getByLabelText('Gezählter Ist-Bestand'), '342,50')
    await user.click(screen.getByRole('button', { name: 'Kasse abschließen' }))

    const buttons = screen.getAllByRole('button', { name: 'Kasse abschließen' })
    await user.click(buttons[buttons.length - 1])

    expect(kasseAbschliessen).toHaveBeenCalledWith(34250)
  })

  it('weist Ausfall-Reste in der Erfolgsmeldung aus', async () => {
    kasseAbschliessen.mockResolvedValueOnce({
      ausfallResteAnzahl: 2,
      ohneKonfigurationAnzahl: 1,
    })
    const user = userEvent.setup()
    render(<KasseAbschliessenSection kassensitzungNr={1} onSuccess={vi.fn()} />)

    await user.type(screen.getByLabelText('Gezählter Ist-Bestand'), '342,50')
    await user.click(screen.getByRole('button', { name: 'Kasse abschließen' }))
    const buttons = screen.getAllByRole('button', { name: 'Kasse abschließen' })
    await user.click(buttons[buttons.length - 1])

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
        alterSekunden: 25,
      }),
    )
    const user = userEvent.setup()
    render(<KasseAbschliessenSection kassensitzungNr={1} onSuccess={vi.fn()} />)

    await user.type(screen.getByLabelText('Gezählter Ist-Bestand'), '342,50')
    await user.click(screen.getByRole('button', { name: 'Kasse abschließen' }))
    const buttons = screen.getAllByRole('button', { name: 'Kasse abschließen' })
    await user.click(buttons[buttons.length - 1])

    expect(toast.warning).toHaveBeenCalledWith(
      expect.stringContaining('2 Vorgänge sind noch nicht signiert'),
    )
    // Dialog bleibt offen: der Abschluss kann erneut angefordert werden.
    expect(screen.getByText('Kasse abschließen?')).toBeInTheDocument()
  })
})
