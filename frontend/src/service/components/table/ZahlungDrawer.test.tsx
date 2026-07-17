import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import { ServiceDock } from '../ServiceDock'
import { ZahlungDrawer } from './ZahlungDrawer'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

afterEach(() => {
  cleanup()
})

const tisch: Tisch = { id: 1, name: 'Stammtisch', saldoCents: 0 }

const position: Position = {
  positionId: '00000000-0000-0000-0000-000000000001',
  varianteId: 1,
  produktName: 'Bratwurst',
  varianteName: 'Normal',
  kategorie: 'essen',
  steuersatz: 'regel',
  einzelpreisCents: 350,
  menge: 2,
  bestellerUserId: 1,
  bestellerName: 'Tester',
}

function renderDrawer(
  zahlungKassieren: () => Promise<void> = vi.fn().mockResolvedValue(undefined),
) {
  const zahlungKassiert = vi.fn()
  // Der Aktionsbutton rendert per Portal in den ServiceDock; deshalb wird der
  // Drawer hier in ein Dock eingebettet (leiste bleibt leer).
  render(
    <ServiceDock leiste={null}>
      <ZahlungDrawer
        backend={{ zahlungKassieren }}
        tisch={tisch}
        unbezahltePositionen={[position]}
        mengen={{ [position.positionId]: 1 }}
        restNachZahlungCents={350}
        zahlungKassiert={zahlungKassiert}
      />
    </ServiceDock>,
  )
  return { zahlungKassiert }
}

async function openDrawer(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole('button', { name: /Kassieren/ }))
  return await screen.findByRole('dialog')
}

describe('ZahlungDrawer', () => {
  it('öffnet über den Trigger und schließt über Abbrechen', async () => {
    const user = userEvent.setup()
    renderDrawer()

    const dialog = await openDrawer(user)
    expect(dialog).toBeVisible()
    expect(
      document.querySelector('[data-slot="drawer-overlay"]'),
    ).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Abbrechen' }))
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })
  })

  it('schließt ohne laufenden Submit über Escape', async () => {
    const user = userEvent.setup()
    renderDrawer()

    await openDrawer(user)
    await user.keyboard('{Escape}')

    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })
  })

  it('scrollt nur im DrawerBody; der Kassieren-Button liegt im Footer außerhalb', async () => {
    const user = userEvent.setup()
    renderDrawer()

    const dialog = await openDrawer(user)
    const body = dialog.querySelector('[data-slot="drawer-body"]')
    expect(body).not.toBeNull()

    // Einziger Scrollbereich: kein weiteres overflow-y-auto im Drawer.
    const scrollContainers = dialog.querySelectorAll(
      '[class*="overflow-y-auto"]',
    )
    expect(scrollContainers).toHaveLength(1)
    expect(scrollContainers[0]).toBe(body)

    // Beleg scrollt im Body, Submit und Abbrechen bleiben außerhalb sichtbar.
    expect(body).toContainElement(screen.getByText(/Bratwurst/))
    const submit = screen.getByRole('button', { name: 'Kassieren' })
    const abbrechen = screen.getByRole('button', { name: 'Abbrechen' })
    expect(body).not.toContainElement(submit)
    expect(body).not.toContainElement(abbrechen)
  })

  it('ignoriert Escape und Backdrop-Tap während des Submits und schließt nach Erfolg', async () => {
    const user = userEvent.setup()
    let resolveSubmit: () => void = () => undefined
    const { zahlungKassiert } = renderDrawer(
      () =>
        new Promise<void>((resolve) => {
          resolveSubmit = resolve
        }),
    )

    const dialog = await openDrawer(user)
    await user.click(screen.getByRole('button', { name: 'Kassieren' }))

    // Pending-Zustand: Drawer markiert, Spinner sichtbar, Buttons deaktiviert.
    expect(dialog).toHaveAttribute('data-pending')
    expect(screen.getByRole('status', { name: 'Loading' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Abbrechen' })).toBeDisabled()

    await user.keyboard('{Escape}')
    expect(screen.getByRole('dialog')).toBeInTheDocument()

    const overlay = document.querySelector('[data-slot="drawer-overlay"]')
    if (overlay === null) throw new Error('Overlay nicht gefunden')
    await user.pointer({
      keys: '[MouseLeft]',
      target: overlay,
    })
    expect(screen.getByRole('dialog')).toBeInTheDocument()

    resolveSubmit()
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })
    expect(zahlungKassiert).toHaveBeenCalled()
  })

  it('zeigt den Tischnamen als Überschrift, das Erhalten-Feld und die Aufrunden-Chips', async () => {
    const user = userEvent.setup()
    renderDrawer()

    await openDrawer(user)

    expect(
      screen.getByRole('heading', { name: tisch.name }),
    ).toBeInTheDocument()

    // Erhalten bleibt ein Feld; der Zielbetrag wird über Chips gesetzt und das
    // freie Feld erscheint erst hinter „Anderer …".
    expect(screen.getByLabelText('Erhalten')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /genau/ })).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Anderer …' }),
    ).toBeInTheDocument()
    expect(
      screen.queryByLabelText('Zahlbetrag inkl. Trinkgeld'),
    ).not.toBeInTheDocument()
  })

  it('hebt das Rückgeld als größten Betrag im Sheet hervor', async () => {
    const user = userEvent.setup()
    renderDrawer()

    await openDrawer(user)
    await user.type(screen.getByLabelText('Erhalten'), '5,00')

    const rueckgeld = screen.getByText('Rückgeld').nextElementSibling
    expect(rueckgeld).toHaveClass('text-xl', 'font-bold', 'tabular-nums')
  })

  it('bleibt im Fehlerfall offen und wieder bedienbar', async () => {
    const { toast } = await import('sonner')
    const user = userEvent.setup()
    renderDrawer(vi.fn().mockRejectedValue(new Error('kaputt')))

    const dialog = await openDrawer(user)
    await user.click(screen.getByRole('button', { name: 'Kassieren' }))

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalled()
    })
    expect(screen.getByRole('dialog')).toBeInTheDocument()
    expect(dialog).not.toHaveAttribute('data-pending')
    expect(screen.getByRole('button', { name: 'Kassieren' })).toBeEnabled()
    expect(screen.getByRole('button', { name: 'Abbrechen' })).toBeEnabled()
  })
})
