import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import { Button, buttonVariants } from './button'

afterEach(() => {
  cleanup()
})

describe('Button destructive-solid variant', () => {
  // Der solide destructive-Button trägt seinen AA-Kontrast über das Token
  // --destructive-solid-foreground (Light: Weiß, Dark: red-950 auf der
  // aufgehellten Fläche). jsdom kennt keine berechneten Farben — der reale
  // Kontrast wird im axe-E2E-Gate (admin-kontrast-axe.spec.ts) gemessen; hier
  // wird nur verankert, dass der Variant existiert und die Token-Flächen- und
  // -Textklassen verdrahtet.
  it('applies the solid destructive surface and the foreground token class', () => {
    const classes = buttonVariants({ variant: 'destructive-solid' })
    expect(classes).toContain('bg-destructive')
    expect(classes).toContain('text-destructive-solid-foreground')
  })

  it('renders through the Button component', () => {
    render(<Button variant="destructive-solid">Löschen</Button>)
    const button = screen.getByRole('button', { name: 'Löschen' })
    expect(button).toHaveAttribute('data-variant', 'destructive-solid')
    expect(button.className).toContain('bg-destructive')
    expect(button.className).toContain('text-destructive-solid-foreground')
  })
})
