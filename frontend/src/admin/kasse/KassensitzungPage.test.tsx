import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { EroeffnenSection, KasseAbschliessenSection } from './KassensitzungPage'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

const { kasseAbschliessen, kassensitzungEroeffnen } = vi.hoisted(() => ({
  kasseAbschliessen: vi
    .fn<(cents: number) => Promise<void>>()
    .mockResolvedValue(undefined),
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

vi.mock('@/admin/settings/hooks', () => ({
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
})
