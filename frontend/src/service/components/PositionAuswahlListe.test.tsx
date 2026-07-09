import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import {
  type AuswahlPosition,
  PositionAuswahlListe,
} from './PositionAuswahlListe'

afterEach(() => {
  cleanup()
})

const positionen: AuswahlPosition[] = [
  { id: 'a', name: 'Cola 0,5l', einzelpreisCents: 500, maxMenge: 3 },
  { id: 'b', name: 'Pommes', einzelpreisCents: 250, maxMenge: 1 },
]

describe('PositionAuswahlListe', () => {
  it('rendert Name, Einzelpreis und Maximalmenge je Position', () => {
    render(
      <PositionAuswahlListe
        positionen={positionen}
        mengen={{}}
        onAdd={vi.fn()}
        onRemove={vi.fn()}
      />,
    )

    expect(screen.getByText('Cola 0,5l')).toBeInTheDocument()
    expect(screen.getByText(/5,00.*·.*3.*Stück/)).toBeInTheDocument()
    expect(screen.getByText('Pommes')).toBeInTheDocument()
  })

  it('zeigt die ausgewählte Menge controlled aus mengen an', () => {
    render(
      <PositionAuswahlListe
        positionen={positionen}
        mengen={{ a: 2 }}
        onAdd={vi.fn()}
        onRemove={vi.fn()}
      />,
    )

    // Cola steht auf 2, Pommes auf 0.
    expect(screen.getByText('2')).toBeInTheDocument()
    expect(screen.getByText('0')).toBeInTheDocument()
  })

  it('meldet Plus mit Position-ID an onAdd', async () => {
    const user = userEvent.setup()
    const onAdd = vi.fn()
    render(
      <PositionAuswahlListe
        positionen={positionen}
        mengen={{}}
        onAdd={onAdd}
        onRemove={vi.fn()}
      />,
    )

    await user.click(
      screen.getByRole('button', { name: 'Cola 0,5l hinzufügen' }),
    )

    expect(onAdd).toHaveBeenCalledTimes(1)
    expect(onAdd).toHaveBeenCalledWith('a')
  })

  it('meldet Minus mit Position-ID an onRemove', async () => {
    const user = userEvent.setup()
    const onRemove = vi.fn()
    render(
      <PositionAuswahlListe
        positionen={positionen}
        mengen={{ a: 1 }}
        onAdd={vi.fn()}
        onRemove={onRemove}
      />,
    )

    await user.click(
      screen.getByRole('button', { name: 'Cola 0,5l verringern' }),
    )

    expect(onRemove).toHaveBeenCalledTimes(1)
    expect(onRemove).toHaveBeenCalledWith('a')
  })

  it('vergibt aria-Labels für alle Stepper-Buttons', () => {
    render(
      <PositionAuswahlListe
        positionen={positionen}
        mengen={{}}
        onAdd={vi.fn()}
        onRemove={vi.fn()}
      />,
    )

    expect(
      screen.getByRole('button', { name: 'Pommes hinzufügen' }),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Pommes verringern' }),
    ).toBeInTheDocument()
  })
})
