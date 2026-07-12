import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Produkt } from '../../product/Produkt'
import { ServiceDock } from '../ServiceDock'
import { Direktverkauf } from './Direktverkauf'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

afterEach(() => {
  cleanup()
})

const testProdukt: Produkt = {
  id: 1,
  name: 'Bratwurst',
  kategorie: 'essen',
  status: 'active',
  varianten: [
    {
      id: 1,
      name: 'Normal',
      preisCents: 350,
      status: 'active',
      createdAt: '2025-01-01T00:00:00Z',
      updatedAt: '2025-01-01T00:00:00Z',
    },
  ],
  createdAt: '2025-01-01T00:00:00Z',
  updatedAt: '2025-01-01T00:00:00Z',
}

function renderDirektverkauf() {
  const direktverkaufTaetigen = vi.fn().mockResolvedValue(undefined)
  // Der Aktionsbutton rendert per Portal in den ServiceDock; deshalb wird der
  // Direktverkauf hier in ein Dock eingebettet (leiste bleibt leer).
  render(
    <ServiceDock leiste={null}>
      <Direktverkauf
        backend={{ direktverkaufTaetigen }}
        products={[testProdukt]}
        productsLoading={false}
      />
    </ServiceDock>,
  )
  return { direktverkaufTaetigen }
}

describe('Direktverkauf', () => {
  it('zeigt die flache Produktliste und den deaktivierten Kassieren-Button ohne Auswahl', () => {
    renderDirektverkauf()

    expect(screen.getByText('Bratwurst')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Kassieren/ })).toBeDisabled()
    expect(screen.queryAllByRole('tab')).toHaveLength(0)
  })

  it('schließt einen Verkauf über den Drawer mit genau einem Backend-Call ab und setzt zurück', async () => {
    const user = userEvent.setup()
    const { direktverkaufTaetigen } = renderDirektverkauf()

    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )

    const kassierenButton = screen.getByRole('button', { name: /Kassieren/ })
    expect(kassierenButton).toBeEnabled()
    expect(kassierenButton).toHaveTextContent('3,50')

    await user.click(kassierenButton)
    const dialog = await screen.findByRole('dialog')
    await user.click(
      screen.getByRole('button', { name: 'Verkauf abschließen' }),
    )

    await waitFor(() => {
      expect(direktverkaufTaetigen).toHaveBeenCalledTimes(1)
    })
    expect(direktverkaufTaetigen).toHaveBeenCalledWith(
      expect.objectContaining({
        // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
        verkaufId: expect.stringMatching(
          /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/,
        ),
        positionen: [{ produktId: 1, varianteId: 1, menge: 1 }],
        kommentar: '',
      }),
    )

    // Nach Erfolg schließt der Drawer und die Auswahl ist zurückgesetzt.
    await waitFor(() => {
      expect(dialog).not.toBeInTheDocument()
    })
    expect(screen.getByRole('button', { name: /Kassieren/ })).toBeDisabled()
  })
})
