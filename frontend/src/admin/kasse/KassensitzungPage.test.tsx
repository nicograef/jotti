import type { QueryClient } from '@tanstack/react-query'
import { cleanup, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { toast } from 'sonner'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { LIVE_REPORTING_KEY } from '@/admin/reporting/hooks'
import type { OffenerTisch } from '@/admin/reporting/types'
import { TSE_KONFIGURATION_KEY } from '@/admin/tse/hooks'
import { BackendError } from '@/lib/Backend'
import { VorgangsRegisterSingleton } from '@/lib/VorgangsRegister'
import { FakeBackend } from '@/test/FakeBackend'
import { renderWithBackend } from '@/test/render'

import { EroeffnenSection } from './EroeffnenSection'
import { KasseAbschliessenSection } from './KasseAbschliessenSection'
import type { AktiveKassensitzung } from './KasseBackend'
import type { GeldtransitBuchung } from './Kassensitzung'
import { KassensitzungPage } from './KassensitzungPage'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
}))

const sommerfest: AktiveKassensitzung = {
  zNr: 12,
  datum: '2026-07-11',
  bezeichnung: 'Sommerfest Tag 2',
  status: 'offen',
  eroeffnetAm: '2026-07-11T08:02:00Z',
}

function liveReporting(offeneTische: OffenerTisch[], offeneSaldiCents: number) {
  return {
    kassensitzungNr: 12,
    bezeichnung: 'Sommerfest Tag 2',
    datum: '2026-07-11',
    offeneTische,
    offeneSaldiCents,
    summary: {
      gesamtUmsatzCents: 12345,
      gesamtBestellungenCents: 12345,
      gesamtStornierungenCents: 300,
      geldtransitCents: 5000,
      anzahlBestellungen: 0,
      anzahlStornierungen: 0,
      anzahlDirektverkaeufe: 0,
      direktverkaufUmsatzCents: 0,
    },
    breakdowns: { servicekraefte: [] },
    stornierungen: [],
    produktStatistik: [],
  }
}

// Soll-Bestand 340,00 € with breakdown; every booking succeeds.
function backend({
  kassensitzung = null,
  buchungen = [],
  offeneTische = [],
  offeneSaldiCents = 0,
  tseKonfiguriert = false,
}: {
  kassensitzung?: AktiveKassensitzung | null
  buchungen?: GeldtransitBuchung[]
  offeneTische?: OffenerTisch[]
  offeneSaldiCents?: number
  tseKonfiguriert?: boolean
} = {}): FakeBackend {
  return new FakeBackend()
    .respond('admin/get-aktive-kassensitzung', kassensitzung)
    .respond('admin/get-kassenbestand', {
      sollBestandCents: 34000,
      anfangsbestandCents: 15000,
      bareinnahmenCents: 17000,
      einlagenCents: 3000,
      entnahmenCents: 1000,
    })
    .respond('admin/get-geldtransit-liste', buchungen)
    .respond(
      'admin/get-live-reporting',
      liveReporting(offeneTische, offeneSaldiCents),
    )
    .respond('admin/get-tse-konfiguration', {
      apiKeyGesetzt: tseKonfiguriert,
      apiSecretGesetzt: tseKonfiguriert,
      tssId: '',
      clientId: '',
      istKonfiguriert: tseKonfiguriert,
    })
    .respond('admin/kassensitzung-eroeffnen', { zNr: 1 })
    .respond('admin/kasse-abschliessen', {
      ausfallResteAnzahl: 0,
      ohneKonfigurationAnzahl: 0,
    })
    .respond('admin/geldtransit-buchen', {})
}

// Waits for a query whose data leaves no visible trace in the UI.
async function geladen(queryClient: QueryClient, key: string) {
  await waitFor(() => {
    expect(queryClient.getQueryState([key])?.status).toBe('success')
  })
}

function renderPage(fake: FakeBackend = backend()) {
  return renderWithBackend(<KassensitzungPage />, fake)
}

async function renderEroeffnen(fake: FakeBackend) {
  const { queryClient } = renderWithBackend(
    <EroeffnenSection onSuccess={vi.fn()} />,
    fake,
  )
  await geladen(queryClient, TSE_KONFIGURATION_KEY)
}

// Returns once Soll-Bestand and the open tables have loaded.
async function renderAbschluss(fake: FakeBackend = backend()) {
  const { queryClient } = renderWithBackend(
    <KasseAbschliessenSection kassensitzungNr={1} onSuccess={vi.fn()} />,
    fake,
  )
  await screen.findByText('340,00 €')
  await geladen(queryClient, LIVE_REPORTING_KEY)
}

beforeEach(() => {
  VorgangsRegisterSingleton.zuruecksetzen()
})

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
})

describe('KassensitzungPage', () => {
  it('zeigt bei Query-Fehler einen Fehlerzustand statt des Steppers', async () => {
    renderPage(backend().fail('admin/get-aktive-kassensitzung'))

    expect(
      await screen.findByText('Kassendaten konnten nicht geladen werden'),
    ).toBeInTheDocument()
    expect(screen.queryByText('2 · Laufender Betrieb')).not.toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Erneut versuchen' }),
    ).toBeInTheDocument()
  })

  it('zeigt im Leerzustand Schritt 1 als aktives Eröffnen-Formular, Schritte 2–3 ausgegraut', async () => {
    renderPage()

    expect(
      await screen.findByRole('button', { name: 'Kassensitzung eröffnen' }),
    ).toBeInTheDocument()
    expect(screen.getByText('2 · Laufender Betrieb')).toBeInTheDocument()
    expect(
      screen.getByText('3 · Am Ende des Tages: Kasse abschließen'),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'Kasse abschließen' }),
    ).not.toBeInTheDocument()
  })

  it('zeigt bei offener Sitzung den Stepper mit Titel, Soll-Bestand-Aufschlüsselung und Bewegungsliste', async () => {
    const buchungen: GeldtransitBuchung[] = [
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
    renderPage(backend({ kassensitzung: sommerfest, buchungen }))

    expect(
      await screen.findByText('Kassentag Nr. 12 — Sommerfest Tag 2'),
    ).toBeInTheDocument()
    // 340,00 € steht in Schritt 2 und in der Live-Rechnung von Schritt 3.
    expect(
      (await screen.findAllByText('340,00 €')).length,
    ).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('Anfangsbestand')).toBeInTheDocument()
    expect(screen.getByText('+ Bareinnahmen')).toBeInTheDocument()
    expect(screen.getByText('+ Einlagen')).toBeInTheDocument()
    expect(screen.getByText('− Entnahmen')).toBeInTheDocument()
    expect(
      await screen.findByText(/Abschöpfung in den Tresor/),
    ).toBeInTheDocument()
    expect(screen.getByText(/Wechselgeld Nachschub/)).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Geld einlegen' }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Geld entnehmen' }),
    ).toBeInTheDocument()
  })

  it('zeigt im Barrierestatus Schritt 3 mit dem Hinweis auf den unterbrochenen Abschluss', async () => {
    renderPage(
      backend({
        kassensitzung: { ...sommerfest, status: 'wird_abgeschlossen' },
      }),
    )

    expect(
      await screen.findByText('Abschluss unterbrochen — erneut abschließen'),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'Kassensitzung eröffnen' }),
    ).not.toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Kasse endgültig abschließen…' }),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'Geld einlegen' }),
    ).not.toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'Geld entnehmen' }),
    ).not.toBeInTheDocument()
    expect(screen.getByText('Soll-Bestand')).toBeInTheDocument()
    expect(screen.getByText('Heutige Kassenbewegungen')).toBeInTheDocument()
  })

  it('öffnet über „Geld entnehmen" den Dialog mit vorbelegter Richtung und bucht', async () => {
    const fake = backend({ kassensitzung: sommerfest })
    const user = userEvent.setup()
    renderPage(fake)

    await user.click(
      await screen.findByRole('button', { name: 'Geld entnehmen' }),
    )

    expect(
      screen.getByRole('heading', { name: 'Geld entnehmen' }),
    ).toBeInTheDocument()

    await user.type(screen.getByLabelText('Betrag'), '30,00')
    await user.type(screen.getByLabelText('Kommentar'), 'Getränke-Nachkauf')
    const dialogButtons = screen.getAllByRole('button', {
      name: 'Geld entnehmen',
    })
    await user.click(dialogButtons[dialogButtons.length - 1])

    await waitFor(() => {
      expect(fake.bodies('admin/geldtransit-buchen')).toEqual([
        {
          geldtransitId: expect.any(String) as unknown,
          richtung: 'entnahme',
          betragCents: 3000,
          kommentar: 'Getränke-Nachkauf',
        },
      ])
    })
  })
})

describe('EroeffnenSection', () => {
  it('fragt ohne TSE-Konfiguration nach; Abbrechen eröffnet nicht, Bestätigen eröffnet', async () => {
    const fake = backend({ tseKonfiguriert: false })
    const user = userEvent.setup()
    await renderEroeffnen(fake)

    await user.type(screen.getByLabelText('Bezeichnung'), 'Sommerfest Tag 1')
    await user.type(screen.getByLabelText('Anfangsbestand'), '150,00')
    await user.click(
      screen.getByRole('button', { name: 'Kassensitzung eröffnen' }),
    )

    expect(screen.getByText('Keine TSE konfiguriert')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Abbrechen' }))
    expect(fake.bodies('admin/kassensitzung-eroeffnen')).toEqual([])

    await user.click(
      screen.getByRole('button', { name: 'Kassensitzung eröffnen' }),
    )
    await user.click(screen.getByRole('button', { name: 'Trotzdem eröffnen' }))
    await waitFor(() => {
      expect(fake.bodies('admin/kassensitzung-eroeffnen')).toEqual([
        { bezeichnung: 'Sommerfest Tag 1', betragCents: 15000 },
      ])
    })
  })

  it('eröffnet mit konfigurierter TSE direkt ohne Dialog', async () => {
    const fake = backend({ tseKonfiguriert: true })
    const user = userEvent.setup()
    await renderEroeffnen(fake)

    await user.type(screen.getByLabelText('Bezeichnung'), 'Sommerfest Tag 1')
    await user.type(screen.getByLabelText('Anfangsbestand'), '150,00')
    await user.click(
      screen.getByRole('button', { name: 'Kassensitzung eröffnen' }),
    )

    expect(screen.queryByText('Keine TSE konfiguriert')).not.toBeInTheDocument()
    await waitFor(() => {
      expect(fake.bodies('admin/kassensitzung-eroeffnen')).toEqual([
        { bezeichnung: 'Sommerfest Tag 1', betragCents: 15000 },
      ])
    })
  })

  it('eröffnet mit 0 € Anfangsbestand (kein Wechselgeld)', async () => {
    const fake = backend({ tseKonfiguriert: true })
    const user = userEvent.setup()
    await renderEroeffnen(fake)

    await user.type(screen.getByLabelText('Bezeichnung'), 'Sommerfest')
    await user.type(screen.getByLabelText('Anfangsbestand'), '0,00')
    await user.click(
      screen.getByRole('button', { name: 'Kassensitzung eröffnen' }),
    )

    await waitFor(() => {
      expect(fake.bodies('admin/kassensitzung-eroeffnen')).toEqual([
        { bezeichnung: 'Sommerfest', betragCents: 0 },
      ])
    })
  })

  it('akzeptiert Standardwert 0 € (leeres Betrag-Feld) ohne Validierungsfehler', async () => {
    // Negativwerte kann EuroInput strukturell nicht erzeugen; deren
    // Schema-Absicherung prüft KasseBackend.test.ts.
    const fake = backend({ tseKonfiguriert: true })
    const user = userEvent.setup()
    await renderEroeffnen(fake)

    await user.type(screen.getByLabelText('Bezeichnung'), 'Sommerfest')
    await user.click(
      screen.getByRole('button', { name: 'Kassensitzung eröffnen' }),
    )

    expect(
      screen.queryByText('Betrag muss mindestens 0 Cent sein.'),
    ).not.toBeInTheDocument()
    await waitFor(() => {
      expect(fake.bodies('admin/kassensitzung-eroeffnen')).toEqual([
        { bezeichnung: 'Sommerfest', betragCents: 0 },
      ])
    })
  })
})

describe('KasseAbschliessenSection', () => {
  it('rechnet die Differenz live als Ist − Soll, Fehlbetrag negativ in Rot', async () => {
    const user = userEvent.setup()
    await renderAbschluss()

    // Soll 340,00 € comes from the faked Kassenbestand.
    expect(screen.getByText('340,00 €')).toBeInTheDocument()
    expect(screen.getByText('0,00 €')).toBeInTheDocument()
    const leerDifferenz = screen.getByText('-340,00 €')
    expect(leerDifferenz).toHaveClass('text-destructive')

    await user.type(screen.getByLabelText('Gezählter Ist-Bestand'), '337,50')
    const fehlbetrag = screen.getByText('-2,50 €')
    expect(fehlbetrag).toBeInTheDocument()
    expect(fehlbetrag).toHaveClass('text-destructive')
  })

  it('färbt einen Überschuss (Ist > Soll) nicht rot', async () => {
    const user = userEvent.setup()
    await renderAbschluss()

    await user.type(screen.getByLabelText('Gezählter Ist-Bestand'), '342,50')
    const ueberschuss = screen.getByText('+2,50 €')
    expect(ueberschuss).toBeInTheDocument()
    expect(ueberschuss).not.toHaveClass('text-destructive')
  })

  it('warnt bei offenen Tischen mit Anzahl und Betrag, ohne offene Tische fehlt die Warnung', async () => {
    await renderAbschluss(
      backend({
        offeneTische: [
          { tischId: 1, tischName: 'Tisch 1', saldoCents: 25000 },
          { tischId: 2, tischName: 'Tisch 2', saldoCents: 16200 },
        ],
        offeneSaldiCents: 41200,
      }),
    )

    expect(
      screen.getByText('2 Tische sind noch offen (412,00 €).'),
    ).toBeInTheDocument()
  })

  it('zeigt ohne offene Tische keine Warnung', async () => {
    await renderAbschluss()

    expect(screen.queryByText(/noch offen/)).not.toBeInTheDocument()
  })

  it('stellt Soll, Ist und Differenz im Bestätigungsdialog gegenüber', async () => {
    const user = userEvent.setup()
    await renderAbschluss()

    await user.type(screen.getByLabelText('Gezählter Ist-Bestand'), '342,50')
    await user.click(
      screen.getByRole('button', { name: 'Kasse endgültig abschließen…' }),
    )

    expect(screen.getByText('Kasse abschließen?')).toBeInTheDocument()
    // Der Ist-Bestand steht in der Live-Rechnung und im Dialog — zweimal.
    expect(screen.getAllByText('342,50 €').length).toBeGreaterThanOrEqual(2)
    expect(
      screen.getByRole('button', { name: 'Kasse abschließen' }),
    ).toBeInTheDocument()
  })

  it('bucht den Abschluss mit dem gezählten Ist-Bestand in Cent', async () => {
    const fake = backend()
    const user = userEvent.setup()
    await renderAbschluss(fake)

    await user.type(screen.getByLabelText('Gezählter Ist-Bestand'), '342,50')
    await user.click(
      screen.getByRole('button', { name: 'Kasse endgültig abschließen…' }),
    )
    await user.click(screen.getByRole('button', { name: 'Kasse abschließen' }))

    await waitFor(() => {
      expect(fake.bodies('admin/kasse-abschliessen')).toEqual([
        { istBestandCents: 34250 },
      ])
    })
  })

  it('übernimmt die Zählhilfe-Summe in das Ist-Bestand-Feld', async () => {
    const user = userEvent.setup()
    await renderAbschluss()

    await user.click(screen.getByRole('button', { name: /Zählhilfe öffnen/ }))
    // 3×100 € (30000) + 2×20 € (4000) = 34000 → 340,00 €.
    await user.type(screen.getByLabelText('100 €'), '3')
    await user.type(screen.getByLabelText('20 €'), '2')
    await user.click(screen.getByRole('button', { name: 'Übernehmen' }))

    expect(screen.getByLabelText('Gezählter Ist-Bestand')).toHaveValue('340,00')
  })

  it('weist Ausfall-Reste in der Erfolgsmeldung aus', async () => {
    const user = userEvent.setup()
    await renderAbschluss(
      backend().respond('admin/kasse-abschliessen', {
        ausfallResteAnzahl: 2,
        ohneKonfigurationAnzahl: 1,
      }),
    )

    await user.type(screen.getByLabelText('Gezählter Ist-Bestand'), '342,50')
    await user.click(
      screen.getByRole('button', { name: 'Kasse endgültig abschließen…' }),
    )
    await user.click(screen.getByRole('button', { name: 'Kasse abschließen' }))

    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith(
        expect.stringContaining('nachsigniert'),
      )
    })
    expect(toast.success).toHaveBeenCalledWith(
      expect.stringContaining('keine TSE konfiguriert'),
    )
  })

  it('zeigt bei ausstehenden Signaturen eine Meldung und lässt den Abschluss erneut anfordern', async () => {
    const user = userEvent.setup()
    await renderAbschluss(
      backend().fail(
        'admin/kasse-abschliessen',
        new BackendError(409, 'signaturen_ausstehend', {
          anzahl: 2,
        }),
      ),
    )

    await user.type(screen.getByLabelText('Gezählter Ist-Bestand'), '342,50')
    await user.click(
      screen.getByRole('button', { name: 'Kasse endgültig abschließen…' }),
    )
    await user.click(screen.getByRole('button', { name: 'Kasse abschließen' }))

    await waitFor(() => {
      expect(toast.warning).toHaveBeenCalledWith(
        expect.stringContaining('2 Vorgänge sind noch nicht signiert'),
      )
    })
    expect(screen.getByText('Kasse abschließen?')).toBeInTheDocument()
  })
})

describe('GeldtransitDialog im Vorgangs-Register', () => {
  it('meldet das angefangene Formular und gibt es beim Schließen frei', async () => {
    const user = userEvent.setup()
    renderPage(backend({ kassensitzung: sommerfest }))

    await user.click(
      await screen.findByRole('button', { name: 'Geld einlegen' }),
    )
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)

    await user.type(screen.getByLabelText('Kommentar'), 'Wechselgeld')
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(1)

    // Beim Schließen bleiben die Werte stehen, das nächste Öffnen verwirft sie
    // — der Vorgang ist damit erledigt.
    await user.click(screen.getByRole('button', { name: 'Abbrechen' }))
    expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
  })
})

describe('KasseAbschliessenSection im Vorgangs-Register', () => {
  it('meldet den eingetippten Ist-Bestand samt offener Rückfrage als einen Vorgang', async () => {
    const user = userEvent.setup()
    await renderAbschluss()

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
    const fake = backend()
    const user = userEvent.setup()
    await renderAbschluss(fake)

    await user.type(screen.getByLabelText('Gezählter Ist-Bestand'), '342,50')
    await user.click(
      screen.getByRole('button', { name: 'Kasse endgültig abschließen…' }),
    )
    await user.click(screen.getByRole('button', { name: 'Kasse abschließen' }))

    await waitFor(() => {
      expect(fake.bodies('admin/kasse-abschliessen')).toEqual([
        { istBestandCents: 34250 },
      ])
    })
    await waitFor(() => {
      expect(VorgangsRegisterSingleton.anzahlOffen()).toBe(0)
    })
  })
})
