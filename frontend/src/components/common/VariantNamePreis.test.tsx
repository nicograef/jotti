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
    // Name und Preis sind eigenständige Elemente — der Preis ist nie Teil des
    // Namens-Knotens und wird deshalb von der Namenslänge nicht mitgezogen.
    expect(name).not.toBe(preis)
    expect(preis).not.toContainElement(name)
  })

  it('keeps the price element intact and separate for a very long name', () => {
    render(
      <div className="flex">
        <VariantNamePreis
          name="Sehr langer Variantenname der die Zeile sprengen würde"
          preisCents={1250}
        />
      </div>,
    )

    // Der Preis-Knoten trägt genau den formatierten Betrag, unabhängig von der
    // Namenslänge (Layout-Kürzung passiert per CSS-truncate, in jsdom nicht
    // messbar — geprüft wird die stabile DOM-Struktur).
    const preis = screen.getByText(/12,50\s*€/)
    expect(preis).toBeInTheDocument()
    expect(preis.textContent).not.toContain('Variantenname')
  })
})
