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
