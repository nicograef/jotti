import { cleanup, render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { KassenberichtePage } from './KassenberichtePage'
import type { Kassensitzung } from './types'

vi.mock('react-router', () => ({
  NavLink: ({ children, to }: { children?: ReactNode; to: string }) => (
    <a href={to}>{children}</a>
  ),
}))

const hookState = vi.hoisted(() => ({
  kassensitzungen: [] as Kassensitzung[],
  listLoading: false,
}))

vi.mock('./hooks', () => ({
  useAbgeschlosseneKassensitzungen: () => ({
    kassensitzungen: hookState.kassensitzungen,
    isPending: hookState.listLoading,
  }),
  useReport: () => ({ result: null, isPending: false }),
  useDsfinvkExport: () => ({ exportieren: vi.fn(), isPending: false }),
}))

afterEach(() => {
  cleanup()
  hookState.kassensitzungen = []
  hookState.listLoading = false
})

describe('KassenberichtePage', () => {
  it('zeigt ohne abgeschlossene Kassensitzung einen erklärenden leeren Zustand mit Link zur Kasse', () => {
    hookState.kassensitzungen = []
    render(<KassenberichtePage />)

    expect(
      screen.getByText('Noch keine abgeschlossene Kassensitzung'),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('link', { name: 'Zur Kassensitzungs-Seite' }),
    ).toHaveAttribute('href', '/admin/kasse')
    // Kein Filter und kein Export im leeren Zustand.
    expect(
      screen.queryByRole('button', { name: 'DSFinV-K-Export' }),
    ).not.toBeInTheDocument()
  })

  it('zeigt bei vorhandenen Kassensitzungen Filter und Export statt des leeren Zustands', () => {
    hookState.kassensitzungen = [
      {
        zNr: 2,
        datum: '2026-07-05',
        bezeichnung: 'Sommerfest Samstag',
        status: 'abgeschlossen',
      },
    ]
    render(<KassenberichtePage />)

    expect(
      screen.queryByText('Noch keine abgeschlossene Kassensitzung'),
    ).not.toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'DSFinV-K-Export' }),
    ).toBeInTheDocument()
  })
})
