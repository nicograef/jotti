import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest'

import { TooltipProvider } from '@/components/ui/tooltip'

// Radix Popper (DropdownMenu-Positionierung) misst seinen Anker über
// ResizeObserver, den jsdom nicht kennt. Ein No-op-Stub reicht für den Test.
class ResizeObserverStub {
  observe(): void {
    // no-op
  }
  unobserve(): void {
    // no-op
  }
  disconnect(): void {
    // no-op
  }
}

beforeAll(() => {
  vi.stubGlobal('ResizeObserver', ResizeObserverStub)
})

import type { DruckstationConfig } from '../settings/DruckstationBackend'
import { Products } from './Products'
import type { Produkt } from './Produkt'

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }))

function variante(overrides: Partial<Produkt['varianten'][number]> = {}) {
  return {
    id: 1,
    name: 'Standard',
    preisCents: 350,
    status: 'active' as const,
    createdAt: '2026-07-01T10:00:00Z',
    updatedAt: '2026-07-01T10:00:00Z',
    ...overrides,
  }
}

function produkt(overrides: Partial<Produkt>): Produkt {
  return {
    id: 1,
    name: 'Produkt',
    kategorie: 'essen',
    steuersatz: 'ermaessigt',
    status: 'active',
    varianten: [variante()],
    createdAt: '2026-07-01T10:00:00Z',
    updatedAt: '2026-07-01T10:00:00Z',
    ...overrides,
  }
}

const druckstationen: DruckstationConfig[] = []

function backend() {
  return {
    aktiviereVariante: vi.fn().mockResolvedValue(undefined),
    deaktiviereVariante: vi.fn().mockResolvedValue(undefined),
    createVariante: vi.fn().mockResolvedValue(1),
    updateVariante: vi.fn().mockResolvedValue(undefined),
    deleteVariante: vi.fn().mockResolvedValue(undefined),
    verschiebeProdukt: vi.fn().mockResolvedValue(undefined),
    verschiebeVariante: vi.fn().mockResolvedValue(undefined),
    sortiereVariantenAlphabetisch: vi.fn().mockResolvedValue(undefined),
  }
}

function renderProducts(products: Produkt[], be = backend()) {
  render(
    <TooltipProvider>
      <Products
        loading={false}
        backend={be}
        products={products}
        druckstationen={druckstationen}
        onEdit={vi.fn()}
        onDelete={vi.fn().mockResolvedValue(undefined)}
        onMoved={vi.fn()}
        onVariantCreated={vi.fn()}
        onVariantUpdated={vi.fn()}
        onVariantStatusChange={vi.fn()}
        onVariantDeleted={vi.fn()}
      />
    </TooltipProvider>,
  )
  return be
}

afterEach(cleanup)

describe('Products', () => {
  it('groups products into Essen and Getränke sections in order', () => {
    renderProducts([
      produkt({ id: 1, name: 'Cola', kategorie: 'getraenk' }),
      produkt({ id: 2, name: 'Pommes', kategorie: 'essen' }),
    ])

    const headings = screen.getAllByRole('heading', { level: 2 })
    expect(headings.map((h) => h.textContent)).toEqual([
      expect.stringContaining('Essen'),
      expect.stringContaining('Getränke'),
    ])
    expect(screen.getByText('Pommes')).toBeInTheDocument()
    expect(screen.getByText('Cola')).toBeInTheDocument()
  })

  it('deactivates a variant via the chip switch without a dialog', async () => {
    const user = userEvent.setup()
    const be = renderProducts([
      produkt({ id: 1, name: 'Bier', varianten: [variante({ id: 7 })] }),
    ])

    const toggle = screen.getByRole('switch', {
      name: /Standard.*deaktivieren/i,
    })
    await user.click(toggle)

    expect(be.deaktiviereVariante).toHaveBeenCalledWith(7)
  })

  it('offers an active Löschen entry in the actions menu', async () => {
    const user = userEvent.setup()
    renderProducts([produkt({ id: 1, name: 'Brezel' })])

    await user.click(screen.getByRole('button', { name: 'Weitere Aktionen' }))

    const loeschen = screen.getByRole('menuitem', { name: /Löschen/ })
    expect(loeschen).not.toHaveAttribute('data-disabled')
  })

  it('moves a produkt down and disables the arrows at the ends of a kategorie', async () => {
    const user = userEvent.setup()
    const be = renderProducts([
      produkt({ id: 1, name: 'Pommes' }),
      produkt({ id: 2, name: 'Brezel' }),
    ])

    // Das erste Produkt kann nicht höher, das letzte nicht tiefer.
    expect(
      screen.getByRole('button', { name: /Pommes.*nach oben/ }),
    ).toBeDisabled()
    expect(
      screen.getByRole('button', { name: /Brezel.*nach unten/ }),
    ).toBeDisabled()

    await user.click(screen.getByRole('button', { name: /Pommes.*nach unten/ }))

    expect(be.verschiebeProdukt).toHaveBeenCalledWith(1, 'runter')
  })

  it('moves a variante within its produkt', async () => {
    const user = userEvent.setup()
    const be = renderProducts([
      produkt({
        id: 1,
        name: 'Bier',
        varianten: [
          variante({ id: 7, name: 'Klein' }),
          variante({ id: 8, name: 'Groß' }),
        ],
      }),
    ])

    expect(
      screen.getByRole('button', { name: /Klein.*nach vorne/ }),
    ).toBeDisabled()

    await user.click(screen.getByRole('button', { name: /Groß.*nach vorne/ }))

    expect(be.verschiebeVariante).toHaveBeenCalledWith(8, 'hoch')
  })
})

describe('Products - alphabetische Variantensortierung', () => {
  it('sorts the varianten of a produkt from the actions menu', async () => {
    const user = userEvent.setup()
    const be = renderProducts([
      produkt({
        id: 4,
        name: 'Weine',
        varianten: [
          variante({ id: 7, name: 'Rotwein' }),
          variante({ id: 8, name: 'Federweisser' }),
        ],
      }),
    ])

    await user.click(screen.getByRole('button', { name: 'Weitere Aktionen' }))
    await user.click(
      screen.getByRole('menuitem', { name: /alphabetisch sortieren/ }),
    )

    expect(be.sortiereVariantenAlphabetisch).toHaveBeenCalledWith(4)
  })

  it('disables the sort entry for a produkt with a single variante', async () => {
    const user = userEvent.setup()
    renderProducts([
      produkt({ id: 5, name: 'Brezel', varianten: [variante({ id: 9 })] }),
    ])

    await user.click(screen.getByRole('button', { name: 'Weitere Aktionen' }))

    expect(
      screen.getByRole('menuitem', { name: /alphabetisch sortieren/ }),
    ).toHaveAttribute('data-disabled')
  })
})
