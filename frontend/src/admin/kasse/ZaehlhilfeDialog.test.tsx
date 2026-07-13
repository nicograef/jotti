import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { ZaehlhilfeDialog } from './ZaehlhilfeDialog'

afterEach(() => {
  cleanup()
})

describe('ZaehlhilfeDialog', () => {
  it('summiert Stückzahlen je Nennwert live und übernimmt die Summe', async () => {
    const user = userEvent.setup()
    const onUebernehmen = vi.fn<(summeCents: number) => void>()
    const onOpenChange = vi.fn<(open: boolean) => void>()
    render(
      <ZaehlhilfeDialog
        open
        onOpenChange={onOpenChange}
        onUebernehmen={onUebernehmen}
      />,
    )

    // 2×50 € (10000) + 3×2 € (600) + 5×20 ct (100) = 10700 → 107,00 €.
    await user.type(screen.getByLabelText('50 €'), '2')
    await user.type(screen.getByLabelText('2 €'), '3')
    await user.type(screen.getByLabelText('20 ct'), '5')

    expect(screen.getByTestId('zaehlhilfe-summe')).toHaveTextContent('107,00 €')

    await user.click(screen.getByRole('button', { name: 'Übernehmen' }))

    expect(onUebernehmen).toHaveBeenCalledWith(10700)
    // Nach der Übernahme schließt der Dialog.
    expect(onOpenChange).toHaveBeenCalledWith(false)
  })

  it('zeigt ohne Eingaben die Summe 0,00 €', () => {
    render(
      <ZaehlhilfeDialog open onOpenChange={vi.fn()} onUebernehmen={vi.fn()} />,
    )

    expect(screen.getByTestId('zaehlhilfe-summe')).toHaveTextContent('0,00 €')
  })

  it('nutzt Text-Felder ohne native Number-Spinner', () => {
    render(
      <ZaehlhilfeDialog open onOpenChange={vi.fn()} onUebernehmen={vi.fn()} />,
    )

    // type="number" zeigt native Spinner-Pfeile; die Stückzahl-Felder sind
    // deshalb konsistente Text-Felder mit numerischer Tastatur (wie EuroInput).
    const feld = screen.getByLabelText('50 €')
    expect(feld).toHaveAttribute('type', 'text')
    expect(feld).toHaveAttribute('inputmode', 'numeric')
  })
})
