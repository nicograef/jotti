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
    fuerMichErledigt: true,
    ...overrides,
  }
}

describe('MeinTischCard', () => {
  it('zählt offene (unbezahlte) Positionen und hebt die eigenen hervor', () => {
    const state = tischSession({
      saldoCents: 1050,
      // p1 und p3 von mir, p2 von Kollegin.
      unbezahltePositionen: [
        position('p1', 1),
        position('p2', 2),
        position('p3', 1),
      ],
      fuerMichErledigt: false,
    })

    render(<MeinTischCard state={state} />)

    // p1, p2, p3 = 3 offen, davon p1 und p3 von mir = 2.
    expect(screen.getByText('3 offen · 2 von dir')).toBeInTheDocument()
    expect(screen.queryByText('Alles bezahlt')).not.toBeInTheDocument()
    // Saldo trägt das "Offen"-Label über dem Betrag.
    expect(screen.getByText('Offen')).toBeInTheDocument()
    expect(screen.getByText(/10,50\s*€/)).toBeInTheDocument()
  })

  it('zeigt "Alles bezahlt" ohne unbezahlte Positionen', () => {
    const state = tischSession({ unbezahltePositionen: [] })

    render(<MeinTischCard state={state} />)

    expect(screen.getByText('Alles bezahlt')).toBeInTheDocument()
    expect(screen.queryByText(/offen/)).not.toBeInTheDocument()
  })
})
