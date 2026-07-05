import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { toast } from 'sonner'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { BackendError } from '@/lib/Backend'

import { EroeffnenSection, KasseAbschliessenSection } from './KassensitzungPage'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
}))

const { kasseAbschliessen, kassensitzungEroeffnen } = vi.hoisted(() => ({
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
}))

vi.mock('./hooks', () => ({
  kasseBackend: { kasseAbschliessen, kassensitzungEroeffnen },
  useKassenbestand: () => ({ kassenbestand: { sollBestandCents: 34000 } }),
  useOffeneKassensitzung: vi.fn(),
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

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
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
