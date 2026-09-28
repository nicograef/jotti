import { cleanup, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { BackendError } from '@/lib/Backend'
import type { Produkt } from '@/lib/produktSchemas'
import { FakeBackend } from '@/test/FakeBackend'
import { renderWithBackend, setViewportWidth } from '@/test/render'

import { DirektverkaufPage } from './DirektverkaufPage'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

const testProdukt: Produkt = {
  id: 1,
  name: 'Bratwurst',
  kategorie: 'essen',
  steuersatz: 'ermaessigt',
  status: 'active',
  varianten: [
    {
      id: 1,
      name: 'Normal',
      preisCents: 350,
      status: 'active',
      createdAt: '2025-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z',
    },
  ],
  createdAt: '2025-01-01T00:00:00Z',
  updatedAt: '2025-01-01T00:00:00Z',
}

// Handy-Pfad: der Fehlerzustand ist in beiden Layouts derselbe.
beforeEach(() => {
  setViewportWidth(375)
})

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
  setViewportWidth(1024)
})

function backend(): FakeBackend {
  return new FakeBackend()
    .respond('service/get-aktive-produkte', { produkte: [] })
    .respond('service/get-direktverkauf-historie', { historie: [] })
}

describe('DirektverkaufPage', () => {
  it('zeigt bei Produkt-Fehler einen Fehlerzustand statt der Leer-Defaults', async () => {
    renderWithBackend(
      <DirektverkaufPage />,
      backend().fail('service/get-aktive-produkte'),
    )

    expect(
      await screen.findByText('Produkte konnten nicht geladen werden'),
    ).toBeInTheDocument()
    // Der Leer-Default (Verkaufs-Summe 0,00 €) darf bei einem Fehler nicht
    // erscheinen — das Sortiment wirkt sonst leer und der Verkauf abgerechnet.
    expect(screen.queryByText(/0,00 €/)).not.toBeInTheDocument()
  })

  it('lädt die Produkte über „Erneut versuchen" neu', async () => {
    const user = userEvent.setup()
    const getAktiveProdukte = vi
      .fn()
      .mockImplementationOnce(() => {
        throw new BackendError(400, 'test_fehler')
      })
      .mockReturnValue({ produkte: [testProdukt] })
    renderWithBackend(
      <DirektverkaufPage />,
      backend().respond('service/get-aktive-produkte', getAktiveProdukte),
    )

    await user.click(
      await screen.findByRole('button', { name: 'Erneut versuchen' }),
    )

    expect(await screen.findByText('Bratwurst')).toBeInTheDocument()
    expect(
      screen.queryByText('Produkte konnten nicht geladen werden'),
    ).not.toBeInTheDocument()
  })

  it('zeigt bei Historie-Fehler einen Fehlerzustand im Historie-Reiter', async () => {
    const user = userEvent.setup()
    renderWithBackend(
      <DirektverkaufPage />,
      backend().fail('service/get-direktverkauf-historie'),
    )

    await user.click(screen.getByRole('tab', { name: 'Historie' }))

    expect(
      await screen.findByText('Historie konnte nicht geladen werden'),
    ).toBeInTheDocument()
  })
})
