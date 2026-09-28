import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import { Button, buttonVariants } from './button'

afterEach(() => {
  cleanup()
})

describe('Button destructive-solid variant', () => {
  // jsdom computes no colours, so the axe E2E gate (admin-kontrast-axe.spec.ts) measures the real AA contrast.
  // This test only pins that the variant exists and wires the token fill and text classes.
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
