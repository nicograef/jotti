import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Stepper } from './Stepper'

afterEach(() => {
  cleanup()
})

describe('Stepper', () => {
  it('deaktiviert den Minus-Button bei Menge 0 eindeutig', async () => {
    const user = userEvent.setup()
    const onRemove = vi.fn()
    render(
      <Stepper
        menge={0}
        onAdd={vi.fn()}
        onRemove={onRemove}
        addLabel="hinzufügen"
        removeLabel="entfernen"
      />,
    )

    const minus = screen.getByRole('button', { name: 'entfernen' })
    // Regulär deaktiviert (nicht antippbar) statt geisterhaft-gestrichelt.
    expect(minus).toBeDisabled()
    // Die frühere „Ghost"-Darstellung (voll deckend + gestrichelt) ließ den
    // deaktivierten Button antippbar wirken — sie darf nicht zurückkehren.
    expect(minus.className).not.toContain('border-dashed')
    expect(minus.className).not.toContain('opacity-100')

    await user.click(minus)
    expect(onRemove).not.toHaveBeenCalled()
  })

  it('aktiviert den Minus-Button ab Menge 1', async () => {
    const user = userEvent.setup()
    const onRemove = vi.fn()
    render(
      <Stepper
        menge={1}
        onAdd={vi.fn()}
        onRemove={onRemove}
        addLabel="hinzufügen"
        removeLabel="entfernen"
      />,
    )

    const minus = screen.getByRole('button', { name: 'entfernen' })
    expect(minus).toBeEnabled()

    await user.click(minus)
    expect(onRemove).toHaveBeenCalledTimes(1)
  })

  it('zeigt Minus und Menge mit minusNurAbEins erst ab Menge 1', async () => {
    const user = userEvent.setup()
    const onRemove = vi.fn()
    const { rerender } = render(
      <Stepper
        menge={0}
        onAdd={vi.fn()}
        onRemove={onRemove}
        addLabel="hinzufügen"
        removeLabel="entfernen"
        minusNurAbEins
      />,
    )

    // Bei Menge 0 sind Minus und Mengenanzeige gar nicht im DOM: kein
    // dauerhaft deaktivierter Knopf je Zeile und nichts davon in der
    // Tab-Reihenfolge. Die Breite reserviert der Aufrufort (ProductList,
    // 132-px-Slot); in jsdom ist sie nicht messbar und wird e2e geprüft.
    expect(screen.queryByRole('button', { name: 'entfernen' })).toBeNull()
    expect(screen.queryByText('0')).toBeNull()
    expect(screen.getByRole('button', { name: 'hinzufügen' })).toBeEnabled()

    rerender(
      <Stepper
        menge={1}
        onAdd={vi.fn()}
        onRemove={onRemove}
        addLabel="hinzufügen"
        removeLabel="entfernen"
        minusNurAbEins
      />,
    )

    // Ab Menge 1 sind beide vorhanden und der Minus-Knopf ist bedienbar.
    const minus = screen.getByRole('button', { name: 'entfernen' })
    expect(minus).toBeEnabled()
    expect(screen.getByText('1')).toBeInTheDocument()

    await user.click(minus)
    expect(onRemove).toHaveBeenCalledTimes(1)
  })

  it('deckelt den Plus-Button über addDisabled', () => {
    render(
      <Stepper
        menge={2}
        onAdd={vi.fn()}
        onRemove={vi.fn()}
        addLabel="hinzufügen"
        removeLabel="entfernen"
        addDisabled
      />,
    )

    expect(screen.getByRole('button', { name: 'hinzufügen' })).toBeDisabled()
  })
})
