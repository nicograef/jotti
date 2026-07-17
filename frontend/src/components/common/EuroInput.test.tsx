import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { EuroInput } from './EuroInput'

function Harness({ initial = '' }: { initial?: string }) {
  const [value, setValue] = useState(initial)
  return <EuroInput value={value} onValueChange={setValue} />
}

afterEach(() => {
  cleanup()
  vi.useRealTimers()
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

  it('kappt Nachkommastellen beim Tippen auf zwei', async () => {
    const user = userEvent.setup()
    render(<Harness />)
    const input = screen.getByPlaceholderText('0,00')

    await user.type(input, '4,505')
    await user.tab()

    expect(input).toHaveValue('4,50')
  })

  it('formatiert während einer Tipp-Pause nicht um (kein Debounce-Reformat)', () => {
    // Fake-Timer vor der Eingabe aktivieren, damit ein etwaiger Debounce-Timer
    // aus onChange unter der Fake-Uhr geplant würde und vom advanceTimersByTime
    // unten tatsächlich feuern könnte — sonst wäre der Test wirkungslos.
    // fireEvent (synchron, ohne eigene Timer) statt userEvent, das unter
    // Fake-Timern hängt.
    vi.useFakeTimers()
    render(<Harness />)
    const input = screen.getByPlaceholderText('0,00')

    fireEvent.change(input, { target: { value: '1' } })

    // Über eine Sekunde warten: früher hätte der Debounce hier zu „1,00" umformatiert.
    act(() => {
      vi.advanceTimersByTime(1500)
    })
    expect(input).toHaveValue('1')

    fireEvent.change(input, { target: { value: '15' } })
    expect(input).toHaveValue('15')

    fireEvent.blur(input)
    expect(input).toHaveValue('15,00')
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
