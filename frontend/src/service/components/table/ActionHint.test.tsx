import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import { ActionHint } from './ActionHint'

afterEach(() => {
  cleanup()
})

describe('ActionHint', () => {
  it('rendert den Grund, wenn die Aktion gesperrt ist', () => {
    render(<ActionHint reason="Ziel-Tisch wählen" />)

    expect(screen.getByText('Ziel-Tisch wählen')).toBeVisible()
  })

  it('rendert nichts, sobald keine Bedingung mehr fehlt', () => {
    const { container } = render(<ActionHint reason={null} />)

    expect(container).toBeEmptyDOMElement()
  })
})
