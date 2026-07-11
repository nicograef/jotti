import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Produkt } from '../../product/Produkt'
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
  render(
    <Direktverkauf
      backend={{ direktverkaufTaetigen }}
      products={[testProdukt]}
      productsLoading={false}
    />,
  )
  return { direktverkaufTaetigen }
}

describe('Direktverkauf', () => {
  it('renders the combined surface (products + confirm button, no tabs)', () => {
    renderDirektverkauf()

    expect(screen.getByText('Bratwurst')).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: /Verkauf abschließen/ }),
    ).toBeInTheDocument()
    expect(screen.queryAllByRole('tab')).toHaveLength(0)
  })

  it('completes a sale with exactly one backend call and resets the input', async () => {
    const user = userEvent.setup()
    const { direktverkaufTaetigen } = renderDirektverkauf()

    const confirmButton = screen.getByRole('button', {
      name: /Verkauf abschließen/,
    })
    expect(confirmButton).toBeDisabled()

    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )

    expect(confirmButton).toBeEnabled()
    await user.click(confirmButton)

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

    // After success the selection resets, disabling the confirm button again.
    await waitFor(() => {
      expect(
        screen.getByRole('button', { name: /Verkauf abschließen/ }),
      ).toBeDisabled()
    })
  })
})
