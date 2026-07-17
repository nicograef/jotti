import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Position } from './table/Bestellung'
import type { TischSession } from './table/Tisch'
import { TablePage } from './TablePage'

function position(positionId: string): Position {
  return {
    positionId,
    varianteId: 1,
    produktName: 'Bratwurst',
    varianteName: 'Normal',
    kategorie: 'essen',
    steuersatz: 'regel',
    einzelpreisCents: 350,
    menge: 1,
    bestellerUserId: 1,
    bestellerName: 'Tester',
  }
}

vi.mock('react-router', () => ({
  useParams: () => ({ tischId: '1' }),
}))

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

// Diese Suite prüft Kopfbereich und Fehlerzustand (identisch in beiden
// Layouts). Auf dem Handy-Pfad trägt nur der Kopf den Tischnamen; ab lg zeigt
// ihn zusätzlich die Abschluss-Spalte — der Split selbst ist manuelle Abnahme.
vi.mock('@/hooks/use-mobile', () => ({
  useIsMobile: () => true,
}))

vi.mock('@/lib/Backend', () => ({
  BackendSingleton: {},
}))

vi.mock('./product/hooks', () => ({
  useAktiveProdukte: () => ({ produkte: [], isPending: false }),
}))

const { getTischState, getTischHistorie } = vi.hoisted(() => ({
  getTischState: vi.fn<() => Promise<TischSession>>(),
  getTischHistorie: vi.fn<() => Promise<unknown[]>>(),
}))

vi.mock('./table/TischBackend', () => ({
  TischBackend: class {
    getTischState = getTischState
    getTischHistorie = getTischHistorie
  },
}))

// Tischzustand mit offenem Saldo. Der Saldo ist bewusst ungleich 0, damit er
// sich im DOM eindeutig von den 0,00-€-Summen der Bestell-Leiste unterscheidet.
const stammtisch: TischSession = {
  tischId: 1,
  tischName: 'Stammtisch',
  saldoCents: 1250,
  unbezahltePositionen: [],
  fuerMichErledigt: true,
}

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
})

function renderPage() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  render(
    <QueryClientProvider client={queryClient}>
      <TablePage />
    </QueryClientProvider>,
  )
}

describe('TablePage', () => {
  it('zeigt bei Query-Fehler einen Fehlerzustand statt der Leer-Defaults', async () => {
    getTischState.mockRejectedValue(new Error('Netzabbruch'))
    getTischHistorie.mockRejectedValue(new Error('Netzabbruch'))
    renderPage()

    expect(
      await screen.findByText('Tischdaten konnten nicht geladen werden'),
    ).toBeInTheDocument()
    // Der Leer-Default (Saldo 0,00 €) darf bei einem Fehler nicht erscheinen —
    // der Tisch wirkt sonst fälschlich abgerechnet.
    expect(screen.queryByText('0,00 €')).not.toBeInTheDocument()
  })

  it('lädt die Tischdaten über „Erneut versuchen" nach einem Fehler neu', async () => {
    getTischState
      .mockRejectedValueOnce(new Error('Netzabbruch'))
      .mockResolvedValue(stammtisch)
    getTischHistorie
      .mockRejectedValueOnce(new Error('Netzabbruch'))
      .mockResolvedValue([])
    const user = userEvent.setup()
    renderPage()

    await user.click(
      await screen.findByRole('button', { name: 'Erneut versuchen' }),
    )

    expect(await screen.findByText('Stammtisch')).toBeInTheDocument()
    expect(
      screen.queryByText('Tischdaten konnten nicht geladen werden'),
    ).not.toBeInTheDocument()
  })

  it('zeigt ohne Fehler den Tischzustand mit Saldo', async () => {
    getTischState.mockResolvedValue(stammtisch)
    getTischHistorie.mockResolvedValue([])
    renderPage()

    expect(await screen.findByText('Stammtisch')).toBeInTheDocument()
    expect(screen.getByText('12,50 €')).toBeInTheDocument()
  })

  it('zeigt "Alles bezahlt" ohne unbezahlte Positionen', async () => {
    getTischState.mockResolvedValue(stammtisch)
    getTischHistorie.mockResolvedValue([])
    renderPage()

    expect(await screen.findByText('Alles bezahlt')).toBeInTheDocument()
  })

  it('zeigt die Anzahl unbezahlter Positionen als Badge', async () => {
    getTischState.mockResolvedValue({
      ...stammtisch,
      unbezahltePositionen: [position('p1'), position('p2')],
    })
    getTischHistorie.mockResolvedValue([])
    renderPage()

    expect(await screen.findByText('2 unbezahlt')).toBeInTheDocument()
    expect(screen.queryByText('Alles bezahlt')).not.toBeInTheDocument()
  })
})
