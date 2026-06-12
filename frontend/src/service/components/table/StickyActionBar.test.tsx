import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { StickyActionBar } from './StickyActionBar'

afterEach(() => {
  cleanup()
})

describe('StickyActionBar', () => {
  it('zeigt Positionsanzahl und Summe der Auswahl', () => {
    render(<StickyActionBar label="Kassieren" anzahl={3} summeCents={1250} />)

    const bar = screen.getByRole('button', { name: /Kassieren/ })
    expect(bar).toHaveTextContent('3')
    expect(bar).toHaveTextContent('12,50')
  })

  it('ist ohne gewählte Position deaktiviert', () => {
    render(
      <StickyActionBar label="Kassieren" anzahl={0} summeCents={0} disabled />,
    )

    expect(screen.getByRole('button', { name: /Kassieren/ })).toBeDisabled()
  })

  it('löst beim Klick den Trigger aus (öffnet den Drawer)', async () => {
    const user = userEvent.setup()
    const onClick = vi.fn()
    render(
      <StickyActionBar
        label="Kassieren"
        anzahl={1}
        summeCents={500}
        onClick={onClick}
      />,
    )

    await user.click(screen.getByRole('button', { name: /Kassieren/ }))

    expect(onClick).toHaveBeenCalledTimes(1)
  })
})
