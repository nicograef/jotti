import { cleanup, render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { AdminDashboardPage } from './AdminDashboardPage'

vi.mock('react-router', () => ({
  NavLink: ({ children, to }: { children?: ReactNode; to: string }) => (
    <a href={to}>{children}</a>
  ),
}))

vi.mock('./hooks', () => ({
  useLiveReporting: () => ({
    liveData: null,
    isPending: false,
    dataUpdatedAt: 0,
    refetch: vi.fn(),
  }),
}))

vi.mock('@/admin/tse/hooks', () => ({
  RUECKSTAND_WARN_SEKUNDEN: 60,
  useTSEStatus: () => ({
    tseStatus: { istKonfiguriert: true },
    isPending: false,
  }),
  useTSESignaturQueue: () => ({
    queue: {
      offeneAuftraege: 0,
      fehlgeschlageneAuftraege: 0,
      rueckstandSekunden: 0,
      letzterFehler: '',
    },
  }),
}))

const druckState = vi.hoisted(() => ({ anzahl: 0 }))

vi.mock('@/admin/settings/hooks', () => ({
  useFehlgeschlageneDruckauftraege: () => ({
    druckauftraege: Array.from({ length: druckState.anzahl }, (_, i) => ({
      id: i + 1,
    })),
  }),
}))

afterEach(() => {
  cleanup()
})

describe('AdminDashboardPage Drucker-Banner', () => {
  it('zeigt bei fehlgeschlagenen Druckaufträgen ein Banner mit Link zu den Druckstationen', () => {
    druckState.anzahl = 2
    render(<AdminDashboardPage />)

    expect(screen.getByRole('alert')).toBeInTheDocument()
    expect(
      screen.getByText(/2 Druckaufträge konnten nicht gedruckt werden\./),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('link', { name: 'Druckstationen' }),
    ).toHaveAttribute('href', '/admin/druckstationen')
  })

  it('formuliert das Banner bei genau einem Druckauftrag im Singular', () => {
    druckState.anzahl = 1
    render(<AdminDashboardPage />)

    expect(
      screen.getByText(/1 Druckauftrag konnte nicht gedruckt werden\./),
    ).toBeInTheDocument()
  })

  it('zeigt ohne fehlgeschlagene Druckaufträge kein Banner', () => {
    druckState.anzahl = 0
    render(<AdminDashboardPage />)

    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })
})
