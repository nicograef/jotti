import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Receipt } from './Receipt'

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

describe('Receipt', () => {
  // Zwei Bestellrunden derselben Variante: gleicher Name, gleicher Preis.
  it('hält zwei gleichnamige Positionen zum selben Preis auseinander', () => {
    const fehler = vi
      .spyOn(console, 'error')
      .mockImplementation(() => undefined)

    render(
      <Receipt
        positionen={[
          { id: 'a', name: 'Bier 0,5', einzelpreisCents: 400, menge: 2 },
          { id: 'b', name: 'Bier 0,5', einzelpreisCents: 400, menge: 1 },
        ]}
      />,
    )

    expect(screen.getByText('2 x Bier 0,5')).toBeInTheDocument()
    expect(screen.getByText('1 x Bier 0,5')).toBeInTheDocument()
    expect(fehler).not.toHaveBeenCalled()
  })
})
