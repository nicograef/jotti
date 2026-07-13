import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { SidebarProvider } from '@/components/ui/sidebar'

import { AdminSidebar } from './AdminSidebar'
import type { OffeneKassensitzung } from './kasse/KasseBackend'

const themeState = vi.hoisted<{
  isDark: boolean
  setTheme: ReturnType<typeof vi.fn>
}>(() => ({ isDark: false, setTheme: vi.fn() }))

vi.mock('@/components/theme-provider', () => ({
  useTheme: () => ({
    isDark: themeState.isDark,
    setTheme: themeState.setTheme,
  }),
}))

// jsdom kennt window.matchMedia nicht (von SidebarProvider via useIsMobile benötigt).
vi.mock('@/hooks/use-mobile', () => ({
  useIsMobile: () => false,
}))

const versionState = vi.hoisted<{ version: string | undefined }>(() => ({
  version: undefined,
}))

vi.mock('./hooks', () => ({
  useVersion: () => versionState.version,
}))

const kasseState = vi.hoisted<{
  kassensitzung: OffeneKassensitzung | null
}>(() => ({
  kassensitzung: null,
}))

vi.mock('./kasse/hooks', () => ({
  useOffeneKassensitzung: () => ({ kassensitzung: kasseState.kassensitzung }),
}))

const druckState = vi.hoisted<{ anzahl: number }>(() => ({ anzahl: 0 }))

vi.mock('@/admin/settings/hooks', () => ({
  useFehlgeschlageneDruckauftraege: () => ({
    druckauftraege: Array.from({ length: druckState.anzahl }),
  }),
}))

const tseState = vi.hoisted<{
  istKonfiguriert: boolean
  rueckstandSekunden: number
  fehlgeschlageneAuftraege: number
}>(() => ({
  istKonfiguriert: true,
  rueckstandSekunden: 0,
  fehlgeschlageneAuftraege: 0,
}))

vi.mock('@/admin/tse/hooks', () => ({
  RUECKSTAND_WARN_SEKUNDEN: 60,
  useTSEStatus: () => ({
    tseStatus: { istKonfiguriert: tseState.istKonfiguriert },
    isPending: false,
  }),
  useTSESignaturQueue: () => ({
    queue: {
      rueckstandSekunden: tseState.rueckstandSekunden,
      fehlgeschlageneAuftraege: tseState.fehlgeschlageneAuftraege,
    },
  }),
}))

function renderSidebar() {
  render(
    <MemoryRouter>
      <SidebarProvider>
        <AdminSidebar />
      </SidebarProvider>
    </MemoryRouter>,
  )
}

const offeneSitzung: OffeneKassensitzung = {
  zNr: 1,
  datum: '2026-07-12',
  bezeichnung: 'Sommerfest Tag 2',
  status: 'offen',
  eroeffnetAm: '2026-07-12T14:05:00+02:00',
}

// Erwartete Uhrzeit im Chip ("seit HH:MM") — aus derselben Quelle abgeleitet,
// damit die Assertion unabhängig von der Test-Zeitzone bleibt.
const erwarteteUhrzeit = new Date(offeneSitzung.eroeffnetAm).toLocaleTimeString(
  'de-DE',
  { hour: '2-digit', minute: '2-digit' },
)

beforeEach(() => {
  versionState.version = undefined
  kasseState.kassensitzung = null
  druckState.anzahl = 0
  tseState.istKonfiguriert = true
  tseState.rueckstandSekunden = 0
  tseState.fehlgeschlageneAuftraege = 0
  themeState.isDark = false
  themeState.setTheme = vi.fn()
})

afterEach(() => {
  cleanup()
})

describe('AdminSidebar', () => {
  it('zeigt die Version im Footer, sobald sie geladen ist', () => {
    versionState.version = 'v1.0.0'
    renderSidebar()

    expect(screen.getByText('jotti v1.0.0')).toBeInTheDocument()
  })

  it('zeigt keine Versionszeile, solange die Version nicht geladen ist', () => {
    renderSidebar()

    expect(screen.queryByText(/jotti v/)).not.toBeInTheDocument()
  })

  it('gliedert die Navigation nach dem Festablauf', () => {
    renderSidebar()

    expect(screen.getByText('Heute')).toBeInTheDocument()
    expect(screen.getByText('Vorbereitung')).toBeInTheDocument()
    expect(screen.getByText('Nach dem Fest')).toBeInTheDocument()
    expect(screen.getByText('Service')).toBeInTheDocument()
  })

  it('zeigt bei geschlossener Kasse den neutralen Kassentag-Chip', () => {
    renderSidebar()

    expect(screen.getByText('Kein Kassentag')).toBeInTheDocument()
    expect(screen.getByText('Kasse geschlossen')).toBeInTheDocument()
    expect(screen.queryByText('Kasse offen')).not.toBeInTheDocument()
  })

  it('zeigt bei offener Kasse Bezeichnung, Status und Eröffnungszeit im Chip', () => {
    kasseState.kassensitzung = offeneSitzung
    renderSidebar()

    expect(screen.getByText('Sommerfest Tag 2')).toBeInTheDocument()
    expect(
      screen.getByText(`Kasse offen · seit ${erwarteteUhrzeit}`),
    ).toBeInTheDocument()
    // Der Kassentag-Menüpunkt bekommt zusätzlich einen grünen Statuspunkt.
    expect(
      screen.getAllByRole('img', { name: 'Kasse offen' }).length,
    ).toBeGreaterThanOrEqual(1)
  })

  it('markiert Bondrucker bei fehlgeschlagenen Druckaufträgen', () => {
    druckState.anzahl = 2
    renderSidebar()

    expect(
      screen.getByRole('img', { name: 'Druckauftrag fehlgeschlagen' }),
    ).toBeInTheDocument()
  })

  it('markiert Finanzamt & TSE bei nicht konfigurierter TSE', () => {
    tseState.istKonfiguriert = false
    renderSidebar()

    expect(
      screen.getByRole('img', { name: 'TSE benötigt Aufmerksamkeit' }),
    ).toBeInTheDocument()
  })

  it('markiert Finanzamt & TSE bei Signatur-Rückstand über der Schwelle', () => {
    tseState.rueckstandSekunden = 120
    renderSidebar()

    expect(
      screen.getByRole('img', { name: 'TSE benötigt Aufmerksamkeit' }),
    ).toBeInTheDocument()
  })

  it('markiert Finanzamt & TSE nicht bei konfigurierter TSE ohne Rückstand', () => {
    renderSidebar()

    expect(
      screen.queryByRole('img', { name: 'TSE benötigt Aufmerksamkeit' }),
    ).not.toBeInTheDocument()
  })

  it('beschriftet den Theme-Umschalter stabil, unabhängig vom aktiven Design', () => {
    themeState.isDark = false
    renderSidebar()
    expect(
      screen.getByRole('button', { name: 'Design wechseln' }),
    ).toBeInTheDocument()
    // Kein aus isDark abgeleitetes „Helles/Dunkles Design" mehr.
    expect(screen.queryByText(/Helles Design|Dunkles Design/)).toBeNull()

    cleanup()

    themeState.isDark = true
    renderSidebar()
    expect(
      screen.getByRole('button', { name: 'Design wechseln' }),
    ).toBeInTheDocument()
    expect(screen.queryByText(/Helles Design|Dunkles Design/)).toBeNull()
  })

  it('schaltet auf das Gegenteil des aktuellen Designs', async () => {
    const user = userEvent.setup()
    themeState.isDark = false
    renderSidebar()

    await user.click(screen.getByRole('button', { name: 'Design wechseln' }))
    expect(themeState.setTheme).toHaveBeenCalledWith('dark')
  })
})
