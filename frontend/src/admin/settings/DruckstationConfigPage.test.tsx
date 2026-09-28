import {
  cleanup,
  screen,
  waitFor,
  waitForElementToBeRemoved,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { toast } from 'sonner'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { FakeBackend } from '@/test/FakeBackend'
import { renderWithBackend } from '@/test/render'

import type {
  DruckstationConfig,
  FehlgeschlagenerDruckauftrag,
} from './DruckstationBackend'
import { DruckstationConfigPage } from './DruckstationConfigPage'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
}))

function makeAuftrag(
  id: number,
  bonArt = 'arbeitsbon',
): FehlgeschlagenerDruckauftrag {
  return {
    id,
    bonArt,
    zielIp: '192.168.1.51',
    referenz: `bestellung-aufgenommen:${String(id)}`,
    versuche: 6,
    letzterFehler: 'drucker nicht erreichbar',
    erstelltAm: new Date().toISOString(),
  }
}

function makeStation(
  overrides: Partial<DruckstationConfig>,
): DruckstationConfig {
  return {
    kategorie: 'essen',
    druckerIp: '192.168.1.50',
    bonmodus: 'pro_position',
    ...overrides,
  }
}

function backend({
  druckstationen = [],
  druckauftraege = [],
}: {
  druckstationen?: DruckstationConfig[]
  druckauftraege?: FehlgeschlagenerDruckauftrag[]
}): FakeBackend {
  return new FakeBackend()
    .respond('admin/get-druckstationen', { druckstationen })
    .respond('admin/get-fehlgeschlagene-druckauftraege', { druckauftraege })
    .respond('admin/druckauftraege-verwerfen', {
      verworfen: druckauftraege.length,
    })
    .respond('admin/update-druckstationen', {})
    .respond('admin/testbon-drucken', {})
}

async function renderPage(fake: FakeBackend) {
  const result = renderWithBackend(<DruckstationConfigPage />, fake)
  await waitFor(() => {
    expect(result.queryClient.isFetching()).toBe(0)
  })
  return result
}

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
})

describe('DruckstationConfigPage — Alarm-Karte', () => {
  it('zeigt bei mehreren fehlgeschlagenen Aufträgen die Alarm-Karte mit "Alle verwerfen" und löst das Sammel-Verwerfen aus', async () => {
    const user = userEvent.setup()
    const fake = backend({ druckauftraege: [makeAuftrag(1), makeAuftrag(2)] })
    await renderPage(fake)

    expect(
      screen.getByText(/2 Bons konnten nicht gedruckt werden/),
    ).toBeInTheDocument()

    const trigger = screen.getByRole('button', { name: 'Alle verwerfen' })
    await user.click(trigger)

    expect(
      screen.getByText('Alle fehlgeschlagenen Druckaufträge verwerfen?'),
    ).toBeInTheDocument()

    const confirmButtons = screen.getAllByRole('button', {
      name: 'Alle verwerfen',
    })
    await user.click(confirmButtons[confirmButtons.length - 1])

    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith('2 Aufträge verworfen.')
    })
    expect(fake.bodies('admin/druckauftraege-verwerfen')).toHaveLength(1)
  })

  it('beschreibt einen fehlgeschlagenen Kassenbeleg nicht als Küchenproblem und übersetzt den Fehlertext', async () => {
    await renderPage(
      backend({ druckauftraege: [makeAuftrag(1, 'kassenbeleg')] }),
    )

    expect(
      screen.getByText('1 Kassenbeleg konnte nicht gedruckt werden'),
    ).toBeInTheDocument()
    expect(screen.queryByText(/Küche/)).not.toBeInTheDocument()
    expect(screen.getByText(/Drucker nicht erreichbar/)).toBeInTheDocument()
  })

  it('nennt Arbeitsbon-Fehldrucke „Bon" und weist auf die fehlende Küche hin', async () => {
    await renderPage(
      backend({ druckauftraege: [makeAuftrag(1, 'arbeitsbon')] }),
    )

    expect(
      screen.getByText(
        '1 Bon konnte nicht gedruckt werden — die Küche hat ihn nicht!',
      ),
    ).toBeInTheDocument()
  })

  it('zeigt bei genau einem Auftrag keinen "Alle verwerfen"-Button, aber "Nochmal drucken" am Auftrag', async () => {
    await renderPage(backend({ druckauftraege: [makeAuftrag(1)] }))

    expect(
      screen.queryByRole('button', { name: 'Alle verwerfen' }),
    ).not.toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Nochmal drucken' }),
    ).toBeInTheDocument()
  })

  it('begrenzt die Fehl-Bon-Liste in der Höhe und macht sie scrollbar, ohne die Aktionen darunter zu verdrängen', async () => {
    const { container } = await renderPage(
      backend({
        druckauftraege: Array.from({ length: 20 }, (_, i) =>
          makeAuftrag(i + 1),
        ),
      }),
    )

    const scrollbereiche = container.querySelectorAll(
      '[class*="overflow-y-auto"]',
    )
    expect(scrollbereiche).toHaveLength(1)
    const liste = scrollbereiche[0]
    expect(liste.className).toMatch(/max-h-/)

    expect(
      screen.getAllByRole('button', { name: 'Nochmal drucken' }),
    ).toHaveLength(20)
    expect(
      screen.getByRole('button', { name: 'Alle verwerfen' }),
    ).toBeInTheDocument()
  })

  it('zeigt ohne fehlgeschlagene Aufträge keine Alarm-Karte', async () => {
    await renderPage(backend({ druckauftraege: [] }))

    expect(
      screen.queryByText(/konnten? nicht gedruckt werden/),
    ).not.toBeInTheDocument()
  })

  it('zeigt die Referenz fachlich und den Rohwert im title-Attribut', async () => {
    await renderPage(backend({ druckauftraege: [makeAuftrag(86)] }))

    const referenzZeile = screen.getByText(/Bestellung Nr\. 86/)
    expect(referenzZeile).toHaveAttribute('title', 'bestellung-aufgenommen:86')
  })
})

describe('DruckstationConfigPage — Stationskarten', () => {
  it('löst den Testbon-Endpunkt für die Station aus und zeigt einen Erfolgs-Toast', async () => {
    const user = userEvent.setup()
    const fake = backend({
      druckstationen: [
        makeStation({ kategorie: 'essen', druckerIp: '192.168.1.50' }),
      ],
    })
    await renderPage(fake)

    await user.click(screen.getByRole('button', { name: /Testbon/ }))

    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith('Testbon an „Essen“ gesendet.')
    })
    expect(fake.bodies('admin/testbon-drucken')).toEqual([
      { kategorie: 'essen' },
    ])
  })

  it('speichert die Drucker-IP on-blur mit Erfolgs-Toast', async () => {
    const user = userEvent.setup()
    const fake = backend({
      druckstationen: [
        makeStation({ kategorie: 'essen', druckerIp: '192.168.1.50' }),
      ],
    })
    await renderPage(fake)

    const input = screen.getByLabelText('Drucker-IP')
    await user.clear(input)
    await user.type(input, '192.168.1.99')
    await user.tab()

    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith(
        'Drucker-IP für „Essen“ gespeichert.',
      )
    })
    expect(fake.bodies('admin/update-druckstationen')).toEqual([
      expect.objectContaining({
        kategorie: 'essen',
        druckerIp: '192.168.1.99',
      }),
    ])
  })

  it('zeigt nach erfolgreichem IP-Speichern eine Inline-Bestätigung, die nach ~2 Sekunden verschwindet', async () => {
    const user = userEvent.setup()
    await renderPage(
      backend({
        druckstationen: [
          makeStation({ kategorie: 'essen', druckerIp: '192.168.1.50' }),
        ],
      }),
    )

    const input = screen.getByLabelText('Drucker-IP')
    await user.clear(input)
    await user.type(input, '192.168.1.99')
    await user.tab()

    expect(await screen.findByText('Gespeichert')).toBeInTheDocument()

    // Nach ~2 Sekunden verschwindet die Bestätigung wieder.
    await waitForElementToBeRemoved(() => screen.queryByText('Gespeichert'), {
      timeout: 3000,
    })
  })

  it('speichert eine unveränderte IP nicht und zeigt bei ungültiger IP einen Fehler ohne Inline-Bestätigung', async () => {
    const user = userEvent.setup()
    const fake = backend({
      druckstationen: [
        makeStation({ kategorie: 'essen', druckerIp: '192.168.1.50' }),
      ],
    })
    await renderPage(fake)

    const input = screen.getByLabelText('Drucker-IP')
    await user.clear(input)
    await user.type(input, '999.1.1.1')
    await user.tab()

    expect(fake.bodies('admin/update-druckstationen')).toHaveLength(0)
    expect(screen.getByText('Ungültige IPv4-Adresse')).toBeInTheDocument()
    expect(screen.queryByText('Gespeichert')).not.toBeInTheDocument()
  })

  it('bietet den Bonmodus „Pro Stück" nur an der Abholbon-Station an und speichert ihn', async () => {
    const user = userEvent.setup()
    const fake = backend({
      druckstationen: [
        makeStation({ kategorie: 'essen', druckerIp: '192.168.1.50' }),
        makeStation({
          kategorie: 'abholbon',
          druckerIp: '192.168.1.77',
          bonmodus: 'pro_bestellung',
        }),
      ],
    })
    await renderPage(fake)

    // Beide Karten zeigen die zwei Standard-Kacheln, „Pro Stück" nur der Abholbon.
    expect(
      screen.getAllByRole('button', { name: /Pro Position/ }),
    ).toHaveLength(2)
    const proStueck = screen.getAllByRole('button', { name: /Pro Stück/ })
    expect(proStueck).toHaveLength(1)

    await user.click(proStueck[0])

    await waitFor(() => {
      expect(fake.bodies('admin/update-druckstationen')).toEqual([
        {
          kategorie: 'abholbon',
          druckerIp: '192.168.1.77',
          bonmodus: 'pro_stueck',
        },
      ])
    })
  })

  it('fasst nicht konfigurierte Stationen als gestrichelte Karte mit "Drucker zuweisen" zusammen', async () => {
    const user = userEvent.setup()
    await renderPage(
      backend({
        druckstationen: [
          makeStation({ kategorie: 'essen', druckerIp: '192.168.1.50' }),
          makeStation({ kategorie: 'sonstiges', druckerIp: '' }),
          makeStation({
            kategorie: 'abholbon',
            druckerIp: '',
            bonmodus: 'pro_bestellung',
          }),
        ],
      }),
    )

    expect(
      screen.getByText(/Sonstiges & Abholbon — kein Drucker/),
    ).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Drucker zuweisen' }))

    // Nach dem Aufklappen sind die IP-Felder der unkonfigurierten Stationen editierbar.
    expect(screen.getAllByLabelText('Drucker-IP').length).toBe(3)
  })
})
