import { cleanup, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { TooltipProvider } from '@/components/ui/tooltip'
import { BackendError } from '@/lib/Backend'
import type { Produkt } from '@/lib/produktSchemas'
import { FakeBackend } from '@/test/FakeBackend'
import { renderWithBackend } from '@/test/render'

import type { DruckstationConfig } from '../settings/DruckstationBackend'
import { EditProductDialog } from './EditProductDialog'
import { NewProductDialog } from './NewProductDialog'
import { Products } from './Products'

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }))

function variante(overrides: Partial<Produkt['varianten'][number]> = {}) {
  return {
    id: 1,
    name: 'Standard',
    preisCents: 350,
    status: 'active' as const,
    ...overrides,
  }
}

function produkt(overrides: Partial<Produkt>): Produkt {
  return {
    id: 1,
    name: 'Produkt',
    kategorie: 'essen',
    steuersatz: 'ermaessigt',
    pfand: false,
    status: 'active',
    varianten: [variante()],
    ...overrides,
  }
}

const druckstationen: DruckstationConfig[] = []

function backend(): FakeBackend {
  return new FakeBackend()
    .respond('admin/activate-variante', {})
    .respond('admin/deactivate-variante', {})
    .respond('admin/verschiebe-produkt', {})
    .respond('admin/verschiebe-variante', {})
    .respond('admin/sortiere-varianten', {})
    .respond('admin/create-produkt', { id: 9 })
    .respond('admin/update-produkt', {})
}

function renderProducts(
  products: Produkt[],
  fake = backend(),
  onChanged = vi.fn(),
) {
  renderWithBackend(
    <TooltipProvider>
      <Products
        loading={false}
        products={products}
        druckstationen={druckstationen}
        onEdit={vi.fn()}
        onDelete={vi.fn().mockResolvedValue(undefined)}
        onChanged={onChanged}
        onVariantCreated={vi.fn()}
        onVariantDeleted={vi.fn()}
      />
    </TooltipProvider>,
    fake,
  )
  return fake
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
    const fake = renderProducts([
      produkt({ id: 1, name: 'Bier', varianten: [variante({ id: 7 })] }),
    ])

    const toggle = screen.getByRole('switch', {
      name: /Standard.*deaktivieren/i,
    })
    await user.click(toggle)

    await waitFor(() => {
      expect(fake.bodies('admin/deactivate-variante')).toEqual([{ id: 7 }])
    })
  })

  it.each([
    {
      fall: 'all succeed',
      fehlschlag: false,
      aufrufe: [{ id: 7 }, { id: 8 }, { id: 9 }],
    },
    {
      fall: 'the second fails',
      fehlschlag: true,
      aufrufe: [{ id: 7 }, { id: 8 }],
    },
  ])(
    'deactivates all variants and reloads the list once when $fall',
    async ({ fehlschlag, aufrufe }) => {
      const user = userEvent.setup()
      const onChanged = vi.fn()
      const fake = backend()
      if (fehlschlag) {
        fake.respond(
          'admin/deactivate-variante',
          vi
            .fn()
            .mockReturnValueOnce({})
            .mockImplementationOnce(() => {
              throw new BackendError(503, 'netzwerkfehler')
            }),
        )
      }
      renderProducts(
        [
          produkt({
            id: 1,
            name: 'Bier',
            varianten: [
              variante({ id: 7, name: 'Klein' }),
              variante({ id: 8, name: 'Groß' }),
              variante({ id: 9, name: 'Maß' }),
            ],
          }),
        ],
        fake,
        onChanged,
      )

      await user.click(screen.getByRole('button', { name: 'Weitere Aktionen' }))
      await user.click(
        screen.getByRole('menuitem', { name: /Alle Varianten deaktivieren/ }),
      )

      await waitFor(() => {
        expect(onChanged).toHaveBeenCalledTimes(1)
      })
      expect(fake.bodies('admin/deactivate-variante')).toEqual(aufrufe)
    },
  )

  it('moves a produkt down and disables the arrows at the ends of a kategorie', async () => {
    const user = userEvent.setup()
    const fake = renderProducts([
      produkt({ id: 1, name: 'Pommes' }),
      produkt({ id: 2, name: 'Brezel' }),
    ])

    expect(
      screen.getByRole('button', { name: /Pommes.*nach oben/ }),
    ).toBeDisabled()
    expect(
      screen.getByRole('button', { name: /Brezel.*nach unten/ }),
    ).toBeDisabled()

    await user.click(screen.getByRole('button', { name: /Pommes.*nach unten/ }))

    await waitFor(() => {
      expect(fake.bodies('admin/verschiebe-produkt')).toEqual([
        { id: 1, richtung: 'runter' },
      ])
    })
  })

  it('hides the variant chevrons when a produkt has a single variante', () => {
    renderProducts([
      produkt({
        id: 1,
        name: 'Kuchen',
        varianten: [variante({ name: 'Stück' })],
      }),
    ])

    expect(
      screen.queryByRole('button', { name: /Stück.*nach vorne/ }),
    ).not.toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: /Stück.*nach hinten/ }),
    ).not.toBeInTheDocument()
  })

  it('moves a variante within its produkt', async () => {
    const user = userEvent.setup()
    const fake = renderProducts([
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

    await waitFor(() => {
      expect(fake.bodies('admin/verschiebe-variante')).toEqual([
        { id: 8, richtung: 'hoch' },
      ])
    })
  })
})

describe('Products - alphabetische Variantensortierung', () => {
  it('sorts the varianten of a produkt from the actions menu', async () => {
    const user = userEvent.setup()
    const fake = renderProducts([
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

    await waitFor(() => {
      expect(fake.bodies('admin/sortiere-varianten')).toEqual([
        { produktId: 4 },
      ])
    })
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

describe('Produkt-Formulare', () => {
  it.each([
    {
      dialog: 'NewProductDialog',
      submit: 'Produkt anlegen',
      endpoint: 'admin/create-produkt',
      open: 'Neues Produkt',
      renderDialog: (fake: FakeBackend) => {
        renderWithBackend(<NewProductDialog created={vi.fn()} />, fake)
      },
    },
    {
      dialog: 'EditProductDialog',
      submit: 'Speichern',
      endpoint: 'admin/update-produkt',
      open: null,
      renderDialog: (fake: FakeBackend) => {
        renderWithBackend(
          <EditProductDialog
            product={produkt({ id: 3, name: 'Pfand' })}
            updated={vi.fn()}
            close={vi.fn()}
          />,
          fake,
        )
      },
    },
  ])(
    'sends pfand when the Pfand switch is on ($dialog)',
    async ({ submit, endpoint, open, renderDialog }) => {
      const user = userEvent.setup()
      const fake = backend()
      renderDialog(fake)
      if (open) {
        await user.click(screen.getByRole('button', { name: open }))
        await user.type(screen.getByLabelText('Name'), 'Pfand')
      }

      await user.click(screen.getByRole('switch', { name: 'Pfand-Produkt' }))
      await user.click(screen.getByRole('button', { name: submit }))

      await waitFor(() => {
        expect(fake.bodies(endpoint)).toEqual([
          expect.objectContaining({ name: 'Pfand', pfand: true }),
        ])
      })
    },
  )

  it('marks a Pfand-Produkt with a badge', () => {
    renderProducts([
      produkt({ id: 1, name: 'Pfand', pfand: true }),
      produkt({ id: 2, name: 'Pommes' }),
    ])

    expect(
      screen.getAllByText('Pfand', { selector: '[data-slot=badge]' }),
    ).toHaveLength(1)
  })
})
