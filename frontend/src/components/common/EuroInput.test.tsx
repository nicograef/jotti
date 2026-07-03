import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, describe, expect, it } from 'vitest'

import { EuroInput } from './EuroInput'

function Harness({ initial = '' }: { initial?: string }) {
  const [value, setValue] = useState(initial)
  return <EuroInput value={value} onValueChange={setValue} />
}

afterEach(() => {
  cleanup()
})

describe('EuroInput', () => {
  it('lässt nur Ziffern und ein Komma zu', async () => {
    const user = userEvent.setup()
    render(<Harness />)
    const input = screen.getByPlaceholderText('0,00')

    await user.type(input, 'a1b2,c5')

    expect(input).toHaveValue('12,5')
  })

  it('normalisiert beim Verlassen auf zwei Nachkommastellen', async () => {
    const user = userEvent.setup()
    render(<Harness />)
    const input = screen.getByPlaceholderText('0,00')

    await user.type(input, '4,5')
    await user.tab()

    expect(input).toHaveValue('4,50')
  })

  it('behandelt einen Punkt als Dezimaltrenner (4.5 → 4,50, nicht 45,00)', async () => {
    const user = userEvent.setup()
    render(<Harness />)
    const input = screen.getByPlaceholderText('0,00')

    await user.type(input, '4.5')
    await user.tab()

    expect(input).toHaveValue('4,50')
  })

  it('verwirft weitere Trennzeichen nach dem ersten', async () => {
    const user = userEvent.setup()
    render(<Harness />)
    const input = screen.getByPlaceholderText('0,00')

    await user.type(input, '1,2,3')
    await user.tab()

    // Der zweite Trenner wird beim Tippen sichtbar verworfen: 1,2 dann 3 → 1,23 €.
    expect(input).toHaveValue('1,23')
  })

  it('leert das Feld beim Verlassen, wenn kein gültiger Betrag eingegeben wurde', async () => {
    const user = userEvent.setup()
    render(<Harness />)
    const input = screen.getByPlaceholderText('0,00')

    await user.type(input, 'abc')
    await user.tab()

    expect(input).toHaveValue('')
  })
})
