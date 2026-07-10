import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { TischSession } from './table/Tisch'
import { TablePage } from './TablePage'

vi.mock('react-router', () => ({
  useParams: () => ({ tischId: '1' }),
}))

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

vi.mock('@/lib/Backend', () => ({
  BackendSingleton: {},
}))

vi.mock('@/hooks/use-mobile', () => ({
  useIsMobile: () => false,
}))

vi.mock('./product/hooks', () => ({
  useAktiveProdukte: () => ({ produkte: [], isPending: false }),
}))

// vaul's Drawer braucht Browser-APIs, die jsdom nicht bereitstellt — der
// Drawer-Inhalt ist für diese Tests irrelevant.
vi.mock('@/components/ui/drawer', () => {
  const Passthrough = ({ children }: { children?: ReactNode }) => children
  return {
    Drawer: Passthrough,
    DrawerTrigger: Passthrough,
    DrawerContent: () => null,
    DrawerHeader: Passthrough,
    DrawerTitle: Passthrough,
    DrawerDescription: Passthrough,
    DrawerFooter: Passthrough,
    DrawerClose: Passthrough,
  }
})

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
  ausstehendePositionen: [],
  gesamtZahlungenCents: 0,
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
})
