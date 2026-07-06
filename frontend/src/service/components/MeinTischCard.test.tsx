import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Position } from '../table/Bestellung'
import type { TischSession } from '../table/Tisch'
import { MeinTischCard } from './MeinTischCard'

vi.mock('react-router', () => ({
  useNavigate: () => vi.fn(),
}))

vi.mock('@/lib/Auth', () => ({
  AuthSingleton: { userId: 1 },
}))

afterEach(() => {
  cleanup()
})

function position(positionId: string, bestellerUserId: number): Position {
  return {
    positionId,
    varianteId: 1,
    produktName: 'Bratwurst',
    varianteName: 'Normal',
    kategorie: 'essen',
    steuersatz: 'regel',
    einzelpreisCents: 350,
    menge: 1,
    bestellerUserId,
    bestellerName: 'Tester',
  }
}

function tischSession(overrides: Partial<TischSession>): TischSession {
  return {
    tischId: 1,
    tischName: 'Stammtisch',
    saldoCents: 0,
    unbezahltePositionen: [],
    ausstehendePositionen: [],
    gesamtZahlungenCents: 0,
    fuerMichErledigt: true,
    ...overrides,
  }
}

describe('MeinTischCard', () => {
  it('zählt offene Positionen als Vereinigung und hebt die eigenen hervor', () => {
    const state = tischSession({
      // p1 von mir (in beiden Listen → einmal gezählt), p2 von Kollegin.
      ausstehendePositionen: [position('p1', 1), position('p2', 2)],
      unbezahltePositionen: [position('p1', 1), position('p3', 1)],
      fuerMichErledigt: false,
    })

    render(<MeinTischCard state={state} />)

    // p1 ∪ p2 ∪ p3 = 3 offen, davon p1 und p3 von mir = 2.
    expect(screen.getByText('3 offen')).toBeInTheDocument()
    expect(screen.getByText('davon 2 von dir')).toBeInTheDocument()
    expect(screen.queryByText('Alles erledigt')).not.toBeInTheDocument()
  })

  it('zeigt "Alles erledigt" wenn keine offenen Positionen', () => {
    render(<MeinTischCard state={tischSession({})} />)

    expect(screen.getByText('Alles erledigt')).toBeInTheDocument()
    expect(screen.queryByText(/offen/)).not.toBeInTheDocument()
  })
})
