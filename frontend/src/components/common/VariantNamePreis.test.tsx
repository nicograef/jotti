import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import { VariantNamePreis } from './VariantNamePreis'

afterEach(() => {
  cleanup()
})

describe('VariantNamePreis', () => {
  it('renders name and formatted price as two separate elements', () => {
    render(
      <div className="flex">
        <VariantNamePreis name="Normal" preisCents={350} />
      </div>,
    )

    const name = screen.getByText('Normal')
    const preis = screen.getByText(/3,50\s*€/)
    // Der Preis liegt nie im Namens-Knoten, wird also nicht mitgekürzt.
    expect(name).not.toBe(preis)
    expect(preis).not.toContainElement(name)
  })
})
