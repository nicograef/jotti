import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { toast } from 'sonner'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { FehlgeschlagenerDruckauftrag } from './DruckstationBackend'
import { DruckstationConfigPage } from './DruckstationConfigPage'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
}))

const { alleVerwerfen } = vi.hoisted(() => ({
  alleVerwerfen: vi.fn<() => Promise<number>>().mockResolvedValue(2),
}))

const fehlgeschlageneState = vi.hoisted(() => ({
  druckauftraege: [] as FehlgeschlagenerDruckauftrag[],
}))

vi.mock('./hooks', () => ({
  useDruckstationen: () => ({
    druckstationen: [],
    isPending: false,
    error: null,
    updateDruckstation: vi.fn(),
  }),
  useFehlgeschlageneDruckauftraege: () => ({
    druckauftraege: fehlgeschlageneState.druckauftraege,
    isPending: false,
    error: null,
    erneutVersuchen: vi.fn(),
    verwerfen: vi.fn(),
    alleVerwerfen,
  }),
}))

function makeAuftrag(id: number): FehlgeschlagenerDruckauftrag {
  return {
    id,
    bonArt: 'arbeitsbon',
    zielIp: '192.168.1.51',
    referenz: `bestellung-aufgenommen:${String(id)}`,
    versuche: 6,
    letzterFehler: 'drucker nicht erreichbar',
    erstelltAm: new Date().toISOString(),
  }
}

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
  fehlgeschlageneState.druckauftraege = []
})

describe('DruckstationConfigPage', () => {
  it('zeigt bei fehlgeschlagenen Aufträgen den Button "Alle verwerfen" und löst nach Bestätigung das Sammel-Verwerfen aus', async () => {
    fehlgeschlageneState.druckauftraege = [makeAuftrag(1), makeAuftrag(2)]
    const user = userEvent.setup()
    render(<DruckstationConfigPage />)

    const trigger = screen.getByRole('button', { name: 'Alle verwerfen' })
    expect(trigger).toBeInTheDocument()

    await user.click(trigger)

    expect(
      screen.getByText('Alle fehlgeschlagenen Druckaufträge verwerfen?'),
    ).toBeInTheDocument()
    expect(screen.getByText(/2 fehlgeschlagene Aufträge/)).toBeInTheDocument()

    const confirmButtons = screen.getAllByRole('button', {
      name: 'Alle verwerfen',
    })
    await user.click(confirmButtons[confirmButtons.length - 1])

    expect(alleVerwerfen).toHaveBeenCalled()
    expect(toast.success).toHaveBeenCalledWith('2 Aufträge verworfen.')
  })

  it('zeigt ohne fehlgeschlagene Aufträge keinen "Alle verwerfen"-Button', () => {
    fehlgeschlageneState.druckauftraege = []
    render(<DruckstationConfigPage />)

    expect(
      screen.queryByRole('button', { name: 'Alle verwerfen' }),
    ).not.toBeInTheDocument()
  })

  it('zeigt die Referenz fachlich und den Rohwert im title-Attribut', () => {
    fehlgeschlageneState.druckauftraege = [makeAuftrag(86)]
    const { container } = render(<DruckstationConfigPage />)

    expect(screen.getByText(/Bestellung Nr\. 86/)).toBeInTheDocument()
    expect(screen.queryByText(/bestellung-aufgenommen:86/)).toBeNull()

    const referenzZeile = container.querySelector('[title]')
    expect(referenzZeile).toHaveAttribute('title', 'bestellung-aufgenommen:86')
  })
})
