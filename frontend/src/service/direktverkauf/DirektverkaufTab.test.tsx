import { cleanup, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ComponentProps } from 'react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useErfolgsPop } from '@/hooks/use-erfolgs-pop'
import { useVorgang } from '@/hooks/use-vorgang'
import { BackendError } from '@/lib/Backend'
import type { Produkt } from '@/lib/produktSchemas'
import { FakeBackend } from '@/test/FakeBackend'
import { testProdukt } from '@/test/fixtures'
import { renderWithBackend, setViewportWidth } from '@/test/render'

import { ErfolgsPop } from '../ErfolgsPop'
import { ServiceDock } from '../ServiceDock'
import { DirektverkaufTab } from './DirektverkaufTab'

// Verkauf und Pfand-Vorgang liegen auf der Seite; der Harness übernimmt ihre Rolle.
function DirektverkaufTabHarness(
  props: Omit<
    ComponentProps<typeof DirektverkaufTab>,
    'vorgang' | 'pfandVorgang'
  >,
) {
  return (
    <DirektverkaufTab
      {...props}
      vorgang={useVorgang<number>()}
      pfandVorgang={useVorgang<number>()}
    />
  )
}

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

// Standardmäßig Handy-Layout (Dock-Aktionsbutton plus Bottom-Sheet); ein Test
// unten schaltet auf Desktop, um die Verdrahtung der festen Spalte zu prüfen.
// Das container-neutrale Verhalten der Spalte deckt DirektverkaufAbschluss.test.
beforeEach(() => {
  setViewportWidth(375)
})

afterEach(() => {
  cleanup()
  setViewportWidth(1024)
})

function backend(): FakeBackend {
  return new FakeBackend().respond('service/direktverkauf-taetigen', {})
}

const pfandProdukt: Produkt = {
  ...testProdukt,
  id: 9,
  name: 'Pfand',
  kategorie: 'sonstiges',
  steuersatz: 'regel',
  pfand: true,
  varianten: [
    { ...testProdukt.varianten[0], id: 19, name: 'Glas', preisCents: 200 },
  ],
}

function renderDirektverkauf() {
  const fake = backend()
  // Der Aktionsbutton rendert per Portal in den ServiceDock; deshalb wird der
  // DirektverkaufTab hier in ein Dock eingebettet (leiste bleibt leer).
  renderWithBackend(
    <ServiceDock leiste={null}>
      <DirektverkaufTabHarness
        pfandProdukte={[]}
        products={[testProdukt]}
        productsLoading={false}
        onErfolg={vi.fn()}
      />
    </ServiceDock>,
    fake,
  )
  return { fake }
}

describe('DirektverkaufTab', () => {
  it('zeigt die flache Produktliste und den deaktivierten Kassieren-Button ohne Auswahl', () => {
    renderDirektverkauf()

    expect(screen.getByText('Bratwurst')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Kassieren/ })).toBeDisabled()
    expect(screen.queryAllByRole('tab')).toHaveLength(0)
  })

  it('schließt einen Verkauf über den Drawer mit genau einem Backend-Call ab und setzt zurück', async () => {
    const user = userEvent.setup()
    const { fake } = renderDirektverkauf()

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
      expect(fake.bodies('service/direktverkauf-taetigen')).toHaveLength(1)
    })
    expect(fake.bodies('service/direktverkauf-taetigen')).toEqual([
      expect.objectContaining({
        // oxlint-disable-next-line typescript/no-unsafe-assignment
        verkaufId: expect.stringMatching(
          /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/,
        ),
        positionen: [{ produktId: 1, varianteId: 1, menge: 1 }],
        kommentar: '',
      }),
    ])

    await waitFor(() => {
      expect(dialog).not.toBeInTheDocument()
    })
    expect(screen.getByRole('button', { name: /Kassieren/ })).toBeDisabled()
  })

  // Ein verrechneter Verkauf, dessen Antwort verloren ging, weicht beim Wiederholen mit anderem
  // Pfand ab. Mit stehendem Korb und alter verkaufId würde jeder weitere Versuch abgewiesen.
  it('leert nach pfandrueckgabe_abweichend den Korb, der nächste Verkauf trägt eine neue verkaufId', async () => {
    setViewportWidth(1024)
    const user = userEvent.setup()
    const direktverkaufTaetigen = vi
      .fn()
      .mockImplementationOnce(() => {
        throw new BackendError(409, 'pfandrueckgabe_abweichend')
      })
      .mockReturnValue({})
    const fake = new FakeBackend().respond(
      'service/direktverkauf-taetigen',
      direktverkaufTaetigen,
    )
    renderWithBackend(
      <DirektverkaufTabHarness
        pfandProdukte={[pfandProdukt]}
        products={[testProdukt]}
        productsLoading={false}
        onErfolg={vi.fn()}
      />,
      fake,
    )

    const [bratwurst, glas] = screen.getAllByRole('button', {
      name: 'Variante hinzufügen',
    }) as [HTMLElement, HTMLElement]
    await user.click(bratwurst)
    await user.click(glas)
    await user.click(
      screen.getByRole('button', { name: 'Verkauf abschließen' }),
    )
    await waitFor(() => {
      expect(
        screen.getByRole('button', { name: 'Verkauf abschließen' }),
      ).toBeDisabled()
    })

    await user.click(
      screen.getAllByRole('button', { name: 'Variante hinzufügen' })[0],
    )
    await user.click(
      screen.getByRole('button', { name: 'Verkauf abschließen' }),
    )
    await waitFor(() => {
      expect(fake.bodies('service/direktverkauf-taetigen')).toHaveLength(2)
    })

    const [erster, zweiter] = fake.bodies('service/direktverkauf-taetigen') as [
      { verkaufId: string },
      { verkaufId: string; positionen: unknown; pfand?: unknown },
    ]
    expect(zweiter.verkaufId).not.toBe(erster.verkaufId)
    expect(zweiter.positionen).toEqual([
      { produktId: 1, varianteId: 1, menge: 1 },
    ])
    expect(zweiter.pfand).toBeUndefined()
  })

  it('rendert ab lg die feste Abschluss-Spalte statt Dock und Drawer', async () => {
    setViewportWidth(1024)
    const user = userEvent.setup()
    // Kein ServiceDock: die feste Spalte trägt den Aktionsbutton selbst.
    renderWithBackend(
      <DirektverkaufTabHarness
        pfandProdukte={[]}
        products={[testProdukt]}
        productsLoading={false}
        onErfolg={vi.fn()}
      />,
      backend(),
    )

    expect(screen.getByText('Bratwurst')).toBeInTheDocument()
    const button = screen.getByRole('button', { name: 'Verkauf abschließen' })
    expect(button).toBeDisabled()
    expect(
      screen.queryByRole('button', { name: /Kassieren/ }),
    ).not.toBeInTheDocument()

    // Auswahl links aktiviert den Button rechts (gemeinsamer Mengen-State).
    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    expect(
      screen.getByRole('button', { name: 'Verkauf abschließen' }),
    ).toBeEnabled()
  })
})

// Verdrahtet wie die Buchungsseiten: Der Erfolg öffnet den Pop, der Refetch
// (hier `reload`) läuft erst beim Schließen.
function DirektverkaufMitPop({ reload }: { reload: () => void }) {
  const { zeigeErfolg, pop } = useErfolgsPop(reload)
  return (
    <ServiceDock leiste={null}>
      <DirektverkaufTabHarness
        pfandProdukte={[]}
        products={[testProdukt]}
        productsLoading={false}
        onErfolg={zeigeErfolg}
      />
      <ErfolgsPop {...pop} />
    </ServiceDock>
  )
}

describe('Erfolgs-Pop im Buchungsflow', () => {
  it('öffnet den Pop ohne Erfolgs-Toast und lädt erst nach dem Schließen nach', async () => {
    const { toast } = await import('sonner')
    const user = userEvent.setup()
    const reload = vi.fn()
    const fake = new FakeBackend().respond('service/direktverkauf-taetigen', {})
    renderWithBackend(<DirektverkaufMitPop reload={reload} />, fake)

    await user.click(
      screen.getByRole('button', { name: 'Variante hinzufügen' }),
    )
    await user.click(screen.getByRole('button', { name: /Kassieren/ }))
    await screen.findByRole('dialog')
    await user.click(
      screen.getByRole('button', { name: 'Verkauf abschließen' }),
    )

    await screen.findByText('Verkauf abgeschlossen.')
    expect(reload).not.toHaveBeenCalled()
    expect(toast.success).not.toHaveBeenCalled()

    await waitFor(
      () => {
        expect(
          screen.queryByText('Verkauf abgeschlossen.'),
        ).not.toBeInTheDocument()
      },
      { timeout: 2000 },
    )
    expect(reload).toHaveBeenCalledTimes(1)
  })
})
