import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { DirektverkaufPage } from './DirektverkaufPage'

// Steuerbarer Testzustand der beiden Lese-Hooks der Seite.
const testState = vi.hoisted(() => ({
  produkteError: false,
  historieError: false,
}))
const reloadProdukte = vi.hoisted(() => vi.fn())

vi.mock('@/lib/Backend', () => ({
  BackendSingleton: {},
}))

// Handy-Pfad: der Fehlerzustand ist in beiden Layouts derselbe.
vi.mock('@/hooks/use-mobile', () => ({
  useIsMobile: () => true,
}))

vi.mock('./product/hooks', () => ({
  useAktiveProdukte: () => ({
    produkte: [],
    isPending: false,
    isError: testState.produkteError,
    refetch: reloadProdukte,
  }),
}))

vi.mock('./direktverkauf/hooks', () => ({
  useDirektverkaufHistorie: () => ({
    historie: [],
    isPending: false,
    isError: testState.historieError,
    refetch: vi.fn(),
  }),
}))

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
  testState.produkteError = false
  testState.historieError = false
})

describe('DirektverkaufPage', () => {
  it('zeigt bei Produkt-Fehler einen Fehlerzustand statt der Leer-Defaults', () => {
    testState.produkteError = true
    render(<DirektverkaufPage />)

    expect(
      screen.getByText('Produkte konnten nicht geladen werden'),
    ).toBeInTheDocument()
    // Der Leer-Default (Verkaufs-Summe 0,00 €) darf bei einem Fehler nicht
    // erscheinen — das Sortiment wirkt sonst leer und der Verkauf abgerechnet.
    expect(screen.queryByText(/0,00 €/)).not.toBeInTheDocument()
  })

  it('lädt die Produkte über „Erneut versuchen" neu', async () => {
    testState.produkteError = true
    const user = userEvent.setup()
    render(<DirektverkaufPage />)

    await user.click(screen.getByRole('button', { name: 'Erneut versuchen' }))

    expect(reloadProdukte).toHaveBeenCalled()
  })

  it('zeigt bei Historie-Fehler einen Fehlerzustand im Historie-Reiter', async () => {
    testState.historieError = true
    const user = userEvent.setup()
    render(<DirektverkaufPage />)

    await user.click(screen.getByRole('tab', { name: 'Historie' }))

    expect(
      screen.getByText('Historie konnte nicht geladen werden'),
    ).toBeInTheDocument()
  })
})
