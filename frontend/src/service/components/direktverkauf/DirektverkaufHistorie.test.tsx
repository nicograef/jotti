import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { DirektverkaufHistorieEintrag } from '../../direktverkauf/Direktverkauf'
import { DirektverkaufHistorie } from './DirektverkaufHistorie'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

vi.mock('@/lib/Auth', () => ({
  AuthSingleton: { canCancel: true },
}))

// vaul's Drawer and Radix' ScrollArea depend on browser APIs unavailable in jsdom.
// Render their children inline so the storno form logic stays testable.
vi.mock('@/components/ui/drawer', () => {
  const Passthrough = ({ children }: { children?: ReactNode }) => children
  return {
    Drawer: Passthrough,
    DrawerContent: Passthrough,
    DrawerHeader: Passthrough,
    DrawerTitle: Passthrough,
    DrawerDescription: Passthrough,
    DrawerFooter: Passthrough,
    DrawerClose: Passthrough,
  }
})

vi.mock('@/components/ui/scroll-area', () => {
  const Passthrough = ({ children }: { children?: ReactNode }) => children
  return { ScrollArea: Passthrough }
})

afterEach(() => {
  cleanup()
})

const positionId = '22222222-2222-2222-2222-222222222222'

const verkauf: DirektverkaufHistorieEintrag = {
  verkaufId: '11111111-1111-1111-1111-111111111111',
  userName: 'Anna',
  getaetigtAm: '2026-06-08T10:00:00Z',
  positionen: [
    {
      positionId,
      varianteId: 1,
      produktName: 'Cola',
      varianteName: '0,5l',
      kategorie: 'getraenk',
      steuersatz: 'regel',
      einzelpreis: 500,
      menge: 2,
    },
  ],
  gesamtbetragCents: 1000,
  kommentar: '',
  offenePositionen: [
    {
      positionId,
      varianteId: 1,
      produktName: 'Cola',
      varianteName: '0,5l',
      kategorie: 'getraenk',
      steuersatz: 'regel',
      einzelpreis: 500,
      menge: 2,
    },
  ],
  gesamtStorniertCents: 0,
  stornierungen: [],
}

describe('DirektverkaufHistorie', () => {
  it('cancels selected positions with exactly one backend call', async () => {
    const user = userEvent.setup()
    const direktverkaufStornieren = vi.fn().mockResolvedValue(undefined)
    const kassenbelegDrucken = vi.fn().mockResolvedValue(undefined)
    const onStorniert = vi.fn()
    render(
      <DirektverkaufHistorie
        historie={[verkauf]}
        historieLoading={false}
        backend={{ direktverkaufStornieren, kassenbelegDrucken }}
        onStorniert={onStorniert}
      />,
    )

    await user.click(screen.getByRole('button', { name: 'Stornieren' }))
    await user.click(
      screen.getByRole('button', { name: 'Cola 0,5l hinzufügen' }),
    )
    await user.type(
      screen.getByPlaceholderText('Kommentar (erforderlich)'),
      'Rückgabe',
    )
    await user.click(
      screen.getByRole('button', { name: /Stornierung erteilen/ }),
    )

    await waitFor(() => {
      expect(direktverkaufStornieren).toHaveBeenCalledTimes(1)
    })
    expect(direktverkaufStornieren).toHaveBeenCalledWith({
      verkaufId: '11111111-1111-1111-1111-111111111111',
      positionen: [{ positionId, menge: 1 }],
      kommentar: 'Rückgabe',
    })
    expect(onStorniert).toHaveBeenCalled()
  })

  it('triggers kassenbeleg print for a sale with exactly one backend call', async () => {
    const user = userEvent.setup()
    const kassenbelegDrucken = vi.fn().mockResolvedValue(undefined)
    render(
      <DirektverkaufHistorie
        historie={[verkauf]}
        historieLoading={false}
        backend={{
          direktverkaufStornieren: vi.fn().mockResolvedValue(undefined),
          kassenbelegDrucken,
        }}
        onStorniert={vi.fn()}
      />,
    )

    await user.click(
      screen.getByRole('button', { name: 'Kassenbeleg drucken' }),
    )

    await waitFor(() => {
      expect(kassenbelegDrucken).toHaveBeenCalledTimes(1)
    })
    expect(kassenbelegDrucken).toHaveBeenCalledWith({
      verkaufId: '11111111-1111-1111-1111-111111111111',
    })
  })

  it('triggers stornobeleg print with the stornierungId of the cancellation', async () => {
    const user = userEvent.setup()
    const kassenbelegDrucken = vi.fn().mockResolvedValue(undefined)
    const stornierterVerkauf: DirektverkaufHistorieEintrag = {
      ...verkauf,
      offenePositionen: [],
      gesamtStorniertCents: 1000,
      stornierungen: [
        {
          stornierungId: '33333333-3333-3333-3333-333333333333',
          storniertAm: '2026-06-08T11:00:00Z',
          gesamtStornierungCents: 1000,
        },
      ],
    }
    render(
      <DirektverkaufHistorie
        historie={[stornierterVerkauf]}
        historieLoading={false}
        backend={{
          direktverkaufStornieren: vi.fn().mockResolvedValue(undefined),
          kassenbelegDrucken,
        }}
        onStorniert={vi.fn()}
      />,
    )

    await user.click(
      screen.getByRole('button', { name: 'Stornobeleg drucken' }),
    )

    await waitFor(() => {
      expect(kassenbelegDrucken).toHaveBeenCalledTimes(1)
    })
    expect(kassenbelegDrucken).toHaveBeenCalledWith({
      verkaufId: '11111111-1111-1111-1111-111111111111',
      stornierungId: '33333333-3333-3333-3333-333333333333',
    })
  })
})
