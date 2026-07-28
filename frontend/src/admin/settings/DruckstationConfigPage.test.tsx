import {
  cleanup,
  render,
  screen,
  waitForElementToBeRemoved,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { toast } from 'sonner'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type {
  DruckstationConfig,
  FehlgeschlagenerDruckauftrag,
} from './DruckstationBackend'
import { DruckstationConfigPage } from './DruckstationConfigPage'

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
}))

const { alleVerwerfen, updateDruckstation, testbonDrucken } = vi.hoisted(
  () => ({
    alleVerwerfen: vi.fn<() => Promise<number>>().mockResolvedValue(2),
    updateDruckstation: vi
      .fn<() => Promise<void>>()
      .mockResolvedValue(undefined),
    testbonDrucken: vi.fn<() => Promise<void>>().mockResolvedValue(undefined),
  }),
)

const druckstationenState = vi.hoisted(() => ({
  druckstationen: [] as DruckstationConfig[],
  isLoadingError: false,
}))

const fehlgeschlageneState = vi.hoisted(() => ({
  druckauftraege: [] as FehlgeschlagenerDruckauftrag[],
}))

vi.mock('./hooks', () => ({
  useDruckstationen: () => ({
    druckstationen: druckstationenState.druckstationen,
    isPending: false,
    isLoadingError: druckstationenState.isLoadingError,
    updateDruckstation,
    testbonDrucken,
  }),
  useFehlgeschlageneDruckauftraege: () => ({
    druckauftraege: fehlgeschlageneState.druckauftraege,
    isPending: false,
    error: null,
    erneutVersuchen: vi.fn(),
    verwerfen: vi.fn(),
    alleVerwerfen,
  }),
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

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
  fehlgeschlageneState.druckauftraege = []
  druckstationenState.druckstationen = []
  druckstationenState.isLoadingError = false
})

describe('DruckstationConfigPage — Alarm-Karte', () => {
  it('zeigt bei mehreren fehlgeschlagenen Aufträgen die Alarm-Karte mit "Alle verwerfen" und löst das Sammel-Verwerfen aus', async () => {
    fehlgeschlageneState.druckauftraege = [makeAuftrag(1), makeAuftrag(2)]
    const user = userEvent.setup()
    render(<DruckstationConfigPage />)

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

    expect(alleVerwerfen).toHaveBeenCalled()
    expect(toast.success).toHaveBeenCalledWith('2 Aufträge verworfen.')
  })

  it('beschreibt einen fehlgeschlagenen Kassenbeleg nicht als Küchenproblem und übersetzt den Fehlertext', () => {
    fehlgeschlageneState.druckauftraege = [makeAuftrag(1, 'kassenbeleg')]
    render(<DruckstationConfigPage />)

    expect(
      screen.getByText('1 Kassenbeleg konnte nicht gedruckt werden'),
    ).toBeInTheDocument()
    expect(screen.queryByText(/Küche/)).not.toBeInTheDocument()
    // Roher Relay-Jargon („drucker nicht erreichbar") wird laienverständlich.
    expect(screen.getByText(/Drucker nicht erreichbar/)).toBeInTheDocument()
  })

  it('nennt Arbeitsbon-Fehldrucke „Bon" und weist auf die fehlende Küche hin', () => {
    fehlgeschlageneState.druckauftraege = [makeAuftrag(1, 'arbeitsbon')]
    render(<DruckstationConfigPage />)

    expect(
      screen.getByText(
        '1 Bon konnte nicht gedruckt werden — die Küche hat ihn nicht!',
      ),
    ).toBeInTheDocument()
  })

  it('zeigt bei genau einem Auftrag keinen "Alle verwerfen"-Button, aber "Nochmal drucken" am Auftrag', () => {
    fehlgeschlageneState.druckauftraege = [makeAuftrag(1)]
    render(<DruckstationConfigPage />)

    expect(
      screen.queryByRole('button', { name: 'Alle verwerfen' }),
    ).not.toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Nochmal drucken' }),
    ).toBeInTheDocument()
  })

  it('begrenzt die Fehl-Bon-Liste in der Höhe und macht sie scrollbar, ohne die Aktionen darunter zu verdrängen', () => {
    fehlgeschlageneState.druckauftraege = Array.from({ length: 20 }, (_, i) =>
      makeAuftrag(i + 1),
    )
    const { container } = render(<DruckstationConfigPage />)

    // Genau ein Scrollbereich (die Liste selbst); höhenbegrenzt via max-h.
    const scrollbereiche = container.querySelectorAll(
      '[class*="overflow-y-auto"]',
    )
    expect(scrollbereiche).toHaveLength(1)
    const liste = scrollbereiche[0]
    expect(liste.className).toMatch(/max-h-/)

    // Alle Zeilen bleiben im DOM (nur visuell gekappt), und die
    // Sammel-Aktion darunter ist weiterhin erreichbar.
    expect(
      screen.getAllByRole('button', { name: 'Nochmal drucken' }),
    ).toHaveLength(20)
    expect(
      screen.getByRole('button', { name: 'Alle verwerfen' }),
    ).toBeInTheDocument()
  })

  it('zeigt ohne fehlgeschlagene Aufträge keine Alarm-Karte', () => {
    fehlgeschlageneState.druckauftraege = []
    render(<DruckstationConfigPage />)

    expect(
      screen.queryByText(/konnten? nicht gedruckt werden/),
    ).not.toBeInTheDocument()
  })

  it('zeigt die Referenz fachlich und den Rohwert im title-Attribut', () => {
    fehlgeschlageneState.druckauftraege = [makeAuftrag(86)]
    render(<DruckstationConfigPage />)

    const referenzZeile = screen.getByText(/Bestellung Nr\. 86/)
    expect(referenzZeile).toHaveAttribute('title', 'bestellung-aufgenommen:86')
  })
})

describe('DruckstationConfigPage — Ladefehler', () => {
  // Nur das gescheiterte Erstladen (`isLoadingError`) ersetzt die Seite. Ein
  // gescheiterter Hintergrund-Refetch lässt sie stehen — siehe hooks.test.ts.
  it('zeigt bei gescheitertem Erstladen eine Meldung statt der Stationen', () => {
    druckstationenState.isLoadingError = true
    fehlgeschlageneState.druckauftraege = [makeAuftrag(1)]
    render(<DruckstationConfigPage />)

    expect(
      screen.getByText('Fehler beim Laden der Druckstationen.'),
    ).toBeInTheDocument()
    expect(screen.queryByLabelText('Drucker-IP')).not.toBeInTheDocument()
  })
})

describe('DruckstationConfigPage — Stationskarten', () => {
  it('löst den Testbon-Endpunkt für die Station aus und zeigt einen Erfolgs-Toast', async () => {
    druckstationenState.druckstationen = [
      makeStation({ kategorie: 'essen', druckerIp: '192.168.1.50' }),
    ]
    const user = userEvent.setup()
    render(<DruckstationConfigPage />)

    await user.click(screen.getByRole('button', { name: /Testbon/ }))

    expect(testbonDrucken).toHaveBeenCalledWith('essen')
    expect(toast.success).toHaveBeenCalledWith('Testbon an „Essen“ gesendet.')
  })

  it('speichert die Drucker-IP on-blur mit Erfolgs-Toast', async () => {
    druckstationenState.druckstationen = [
      makeStation({ kategorie: 'essen', druckerIp: '192.168.1.50' }),
    ]
    const user = userEvent.setup()
    render(<DruckstationConfigPage />)

    const input = screen.getByLabelText('Drucker-IP')
    await user.clear(input)
    await user.type(input, '192.168.1.99')
    await user.tab()

    expect(updateDruckstation).toHaveBeenCalledWith(
      expect.objectContaining({
        kategorie: 'essen',
        druckerIp: '192.168.1.99',
      }),
    )
    expect(toast.success).toHaveBeenCalledWith(
      'Drucker-IP für „Essen“ gespeichert.',
    )
  })

  it('zeigt nach erfolgreichem IP-Speichern eine Inline-Bestätigung, die nach ~2 Sekunden verschwindet', async () => {
    druckstationenState.druckstationen = [
      makeStation({ kategorie: 'essen', druckerIp: '192.168.1.50' }),
    ]
    const user = userEvent.setup()
    render(<DruckstationConfigPage />)

    const input = screen.getByLabelText('Drucker-IP')
    await user.clear(input)
    await user.type(input, '192.168.1.99')
    await user.tab()

    // Inline-Bestätigung am Feld zusätzlich zum Toast.
    expect(await screen.findByText('Gespeichert')).toBeInTheDocument()

    // Nach ~2 Sekunden verschwindet die Bestätigung wieder.
    await waitForElementToBeRemoved(() => screen.queryByText('Gespeichert'), {
      timeout: 3000,
    })
  })

  it('speichert eine unveränderte IP nicht und zeigt bei ungültiger IP einen Fehler ohne Inline-Bestätigung', async () => {
    druckstationenState.druckstationen = [
      makeStation({ kategorie: 'essen', druckerIp: '192.168.1.50' }),
    ]
    const user = userEvent.setup()
    render(<DruckstationConfigPage />)

    const input = screen.getByLabelText('Drucker-IP')
    await user.clear(input)
    await user.type(input, '999.1.1.1')
    await user.tab()

    expect(updateDruckstation).not.toHaveBeenCalled()
    expect(screen.getByText('Ungültige IPv4-Adresse')).toBeInTheDocument()
    expect(screen.queryByText('Gespeichert')).not.toBeInTheDocument()
  })

  it('fasst nicht konfigurierte Stationen als gestrichelte Karte mit "Drucker zuweisen" zusammen', async () => {
    druckstationenState.druckstationen = [
      makeStation({ kategorie: 'essen', druckerIp: '192.168.1.50' }),
      makeStation({ kategorie: 'sonstiges', druckerIp: '' }),
      makeStation({
        kategorie: 'abholbon',
        druckerIp: '',
        bonmodus: 'pro_bestellung',
      }),
    ]
    const user = userEvent.setup()
    render(<DruckstationConfigPage />)

    expect(
      screen.getByText(/Sonstiges & Abholbon — kein Drucker/),
    ).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Drucker zuweisen' }))

    // Nach dem Aufklappen sind die IP-Felder der bisher unkonfigurierten
    // Stationen editierbar.
    expect(screen.getAllByLabelText('Drucker-IP').length).toBe(3)
  })
})
