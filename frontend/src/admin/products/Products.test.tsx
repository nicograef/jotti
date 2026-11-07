import { cleanup, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { TooltipProvider } from '@/components/ui/tooltip'
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
    pfandVarianteId: null,
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
    .respond('admin/deactivate-varianten', {})
    .respond('admin/verschiebe-produkt', {})
    .respond('admin/verschiebe-variante', {})
    .respond('admin/sortiere-varianten', {})
    .respond('admin/create-produkt', { id: 9 })
    .respond('admin/update-produkt', {})
    .respond('admin/create-variante', { id: 11 })
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

  it('deactivates all variants in one request and reloads the list', async () => {
    const user = userEvent.setup()
    const onChanged = vi.fn()
    const fake = backend()
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
    expect(fake.bodies('admin/deactivate-varianten')).toEqual([
      { produktId: 1 },
    ])
    expect(fake.bodies('admin/deactivate-variante')).toEqual([])
  })

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

describe('Pfand-Verknüpfung', () => {
  const bier = produkt({
    id: 1,
    name: 'Bier',
    kategorie: 'getraenk',
    steuersatz: 'regel',
    varianten: [variante({ id: 10, name: '0,5 l', preisCents: 400 })],
  })
  const glaspfand = produkt({
    id: 2,
    name: 'Glaspfand',
    kategorie: 'sonstiges',
    steuersatz: 'regel',
    pfand: true,
    varianten: [variante({ id: 20, name: 'Glas', preisCents: 200 })],
  })
  const tellerpfand = produkt({
    id: 3,
    name: 'Tellerpfand',
    kategorie: 'sonstiges',
    steuersatz: 'ermaessigt',
    pfand: true,
    varianten: [variante({ id: 30, name: 'Teller', preisCents: 300 })],
  })

  async function oeffneNeueVariante(produktName: string) {
    const user = userEvent.setup()
    const zeile = screen.getByText(produktName).parentElement
    if (zeile === null) throw new Error(`Zeile ${produktName} fehlt`)
    await user.click(within(zeile).getByRole('button', { name: 'Variante' }))
    return user
  }

  it('Pfand-Verknüpfung: ohne passende Pfand-Variante fehlt das Feld Pfand', async () => {
    renderProducts([bier, tellerpfand])

    await oeffneNeueVariante('Bier')

    expect(
      await screen.findByRole('heading', { name: 'Neue Variante anlegen' }),
    ).toBeInTheDocument()
    expect(screen.queryByRole('combobox', { name: 'Pfand' })).toBeNull()
  })

  it('Pfand-Verknüpfung: bietet nur das Pfand gleichen Steuersatzes an und sendet pfandVarianteId', async () => {
    const fake = renderProducts([bier, glaspfand, tellerpfand])

    const user = await oeffneNeueVariante('Bier')
    await user.type(await screen.findByLabelText('Name'), '0,3 l')
    await user.type(screen.getByLabelText('Preis'), '3')
    await user.click(screen.getByRole('combobox', { name: 'Pfand' }))

    const optionen = screen.getAllByRole('option').map((o) => o.textContent)
    expect(optionen).toEqual(['Kein Pfand', 'Glaspfand Glas · 2,00\u00A0€'])

    await user.click(screen.getByRole('option', { name: /Glaspfand Glas/ }))
    await user.click(screen.getByRole('button', { name: 'Variante anlegen' }))

    await waitFor(() => {
      expect(fake.bodies('admin/create-variante')).toEqual([
        { produktId: 1, name: '0,3 l', preisCents: 300, pfandVarianteId: 20 },
      ])
    })
  })

  it('Pfand-Verknüpfung: der Chip einer verknüpften Variante zeigt + Pfand', () => {
    renderProducts([
      produkt({
        ...bier,
        varianten: [variante({ id: 10, name: '0,5 l', pfandVarianteId: 20 })],
      }),
      glaspfand,
    ])

    expect(screen.getAllByText('+ Pfand')).toHaveLength(1)
  })
})

describe('Lösch-Dialoge: Verknüpfte Varianten verlieren ihr Pfand', () => {
  const glaspfand = produkt({
    id: 2,
    name: 'Glaspfand',
    kategorie: 'sonstiges',
    steuersatz: 'regel',
    pfand: true,
    varianten: [
      variante({ id: 20, name: 'Glas', preisCents: 200 }),
      variante({ id: 21, name: 'Becher', preisCents: 100 }),
    ],
  })
  const bierMit = (pfandVarianteId: number | null) =>
    produkt({
      id: 1,
      name: 'Bier',
      kategorie: 'getraenk',
      steuersatz: 'regel',
      varianten: [variante({ id: 10, name: '0,5 l', pfandVarianteId })],
    })
  const satz = 'Verknüpfte Varianten verlieren ihr Pfand.'

  it.each([
    ['nennt, dass verknüpfte Varianten ihr Pfand verlieren', 20, true],
    [
      'schweigt ohne Verknüpfung, dass Varianten ihr Pfand verlieren',
      null,
      false,
    ],
  ])('Produkt löschen %s', async (_titel, pfandVarianteId, erwartet) => {
    renderProducts([bierMit(pfandVarianteId), glaspfand])
    const user = userEvent.setup()
    const zeile = screen.getByText('Glaspfand').parentElement
    if (zeile === null) throw new Error('Zeile Glaspfand fehlt')

    await user.click(
      within(zeile).getByRole('button', { name: 'Weitere Aktionen' }),
    )
    await user.click(await screen.findByRole('menuitem', { name: 'Löschen…' }))

    const dialog = await screen.findByRole('alertdialog', {
      name: 'Produkt löschen?',
    })
    expect(dialog.textContent.includes(satz)).toBe(erwartet)
  })

  it.each([
    ['nennt, dass verknüpfte Varianten ihr Pfand verlieren', 20, true],
    [
      'schweigt bei Verknüpfung einer anderen Variante, dass Varianten ihr Pfand verlieren',
      21,
      false,
    ],
  ])('Variante löschen %s', async (_titel, pfandVarianteId, erwartet) => {
    renderProducts([bierMit(pfandVarianteId), glaspfand])
    const user = userEvent.setup()

    await user.click(
      screen.getByRole('button', { name: 'Variante „Glas" bearbeiten' }),
    )
    await user.click(
      await screen.findByRole('button', { name: 'Variante löschen' }),
    )

    const dialog = await screen.findByRole('alertdialog', {
      name: 'Variante löschen?',
    })
    expect(dialog.textContent.includes(satz)).toBe(erwartet)
  })
})
