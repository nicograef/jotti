import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import {
  Drawer,
  DrawerContent,
  DrawerTitle,
  DrawerTrigger,
} from '@/components/ui/drawer'

import { ServiceDock } from '../ServiceDock'
import { DockActionButton } from './DockActionButton'

afterEach(() => {
  cleanup()
})

// Der Button rendert nur über den Portal-Slot von ServiceDock; deshalb wird er
// in allen Tests in ein Dock eingebettet (leiste bleibt leer).
function renderInDock(button: React.ReactNode) {
  return render(<ServiceDock leiste={null}>{button}</ServiceDock>)
}

describe('DockActionButton', () => {
  it('zeigt Positionsanzahl und Summe der Auswahl', () => {
    renderInDock(
      <DockActionButton label="Kassieren" anzahl={3} summeCents={1250} />,
    )

    const bar = screen.getByRole('button', { name: /Kassieren/ })
    expect(bar).toHaveTextContent('3')
    expect(bar).toHaveTextContent('12,50')
  })

  it('ist ohne gewählte Position deaktiviert', () => {
    renderInDock(
      <DockActionButton label="Kassieren" anzahl={0} summeCents={0} disabled />,
    )

    expect(screen.getByRole('button', { name: /Kassieren/ })).toBeDisabled()
  })

  it('löst beim Klick den übergebenen onClick aus', async () => {
    const user = userEvent.setup()
    const onClick = vi.fn()
    renderInDock(
      <DockActionButton
        label="Kassieren"
        anzahl={1}
        summeCents={500}
        onClick={onClick}
      />,
    )

    await user.click(screen.getByRole('button', { name: /Kassieren/ }))

    expect(onClick).toHaveBeenCalledTimes(1)
  })

  // Kernrisiko der Phase: Der DrawerTrigger klont den Button über die
  // Portal-Grenze hinweg. Der Klick auf den im Dock gerenderten Button muss den
  // Drawer öffnen (Radix-Context bleibt über das Portal erhalten).
  it('öffnet den Drawer über den Dock-Button (Portal + DrawerTrigger)', async () => {
    const user = userEvent.setup()
    render(
      <Drawer>
        <ServiceDock leiste={null}>
          <DrawerTrigger asChild>
            <DockActionButton label="Kassieren" anzahl={1} summeCents={500} />
          </DrawerTrigger>
        </ServiceDock>
        <DrawerContent>
          <DrawerTitle>Zahlung</DrawerTitle>
        </DrawerContent>
      </Drawer>,
    )

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /Kassieren/ }))

    expect(await screen.findByRole('dialog')).toBeInTheDocument()
    expect(screen.getByText('Zahlung')).toBeInTheDocument()
  })
})
