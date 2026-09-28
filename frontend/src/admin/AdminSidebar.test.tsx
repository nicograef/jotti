import { cleanup, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it } from 'vitest'

import type { TSESignaturQueue } from '@/admin/tse/TSEBackend'
import { ThemeProvider } from '@/components/theme-provider'
import { SidebarProvider } from '@/components/ui/sidebar'
import { FakeBackend } from '@/test/FakeBackend'
import { renderWithBackend } from '@/test/render'

import { AdminSidebar } from './AdminSidebar'
import type { AktiveKassensitzung } from './kasse/KasseBackend'
import type { FehlgeschlagenerDruckauftrag } from './settings/DruckstationBackend'

// Storage key of the ThemeProvider; a stored theme wins over the system theme.
const THEME_STORAGE_KEY = 'vite-ui-theme'

const ruhigeQueue: TSESignaturQueue = {
  offeneAuftraege: 0,
  fehlgeschlageneAuftraege: 0,
  letzterFehler: '',
  rueckstandSekunden: 0,
  signaturenProMinute: 0,
  signierdauerP95Sekunden: 0,
}

function druckauftrag(id: number): FehlgeschlagenerDruckauftrag {
  return {
    id,
    bonArt: 'arbeitsbon',
    zielIp: '192.168.1.51',
    referenz: `bestellung-aufgenommen:${String(id)}`,
    versuche: 6,
    letzterFehler: 'drucker nicht erreichbar',
    erstelltAm: '2026-07-12T14:05:00+02:00',
  }
}

function backend({
  kassensitzung = null,
  druckauftraege = [],
  istKonfiguriert = true,
  queue = ruhigeQueue,
}: {
  kassensitzung?: AktiveKassensitzung | null
  druckauftraege?: FehlgeschlagenerDruckauftrag[]
  istKonfiguriert?: boolean
  queue?: TSESignaturQueue
} = {}): FakeBackend {
  return new FakeBackend()
    .respond('health', { version: 'v1.0.0' })
    .respond('admin/get-aktive-kassensitzung', kassensitzung)
    .respond('admin/get-fehlgeschlagene-druckauftraege', { druckauftraege })
    .respond('admin/get-tse-status', { umgebung: 'LIVE', istKonfiguriert })
    .respond('admin/get-tse-signatur-queue', queue)
}

function renderSidebar(fake: FakeBackend) {
  return renderWithBackend(
    <MemoryRouter>
      <ThemeProvider>
        <SidebarProvider>
          <AdminSidebar />
        </SidebarProvider>
      </ThemeProvider>
    </MemoryRouter>,
    fake,
  )
}

// Before its queries answer, the sidebar shows the closed-register and
// no-warning defaults; absence assertions only mean something afterwards.
async function renderGeladeneSidebar(fake: FakeBackend = backend()) {
  const result = renderSidebar(fake)
  await waitFor(() => {
    expect(result.queryClient.isFetching()).toBe(0)
  })
  return result
}

const aktiveSitzung: AktiveKassensitzung = {
  zNr: 1,
  datum: '2026-07-12',
  bezeichnung: 'Sommerfest Tag 2',
  status: 'offen',
  eroeffnetAm: '2026-07-12T14:05:00+02:00',
}

// Aus derselben Quelle abgeleitet, damit die Assertion zeitzonenunabhängig bleibt.
const erwarteteUhrzeit = new Date(aktiveSitzung.eroeffnetAm).toLocaleTimeString(
  'de-DE',
  { hour: '2-digit', minute: '2-digit' },
)

afterEach(() => {
  cleanup()
  localStorage.removeItem(THEME_STORAGE_KEY)
  document.documentElement.classList.remove('light', 'dark')
})

describe('AdminSidebar', () => {
  it('zeigt die Version im Footer, sobald sie geladen ist', async () => {
    renderSidebar(backend())

    expect(await screen.findByText('jotti v1.0.0')).toBeInTheDocument()
  })

  it('zeigt keine Versionszeile, solange die Version nicht geladen ist', () => {
    // The health request never answers, so the version stays unloaded.
    renderSidebar(
      backend().respond('health', () => new Promise(() => undefined)),
    )

    expect(screen.queryByText(/jotti v/)).not.toBeInTheDocument()
  })

  it('gliedert die Navigation nach dem Festablauf', async () => {
    await renderGeladeneSidebar()

    expect(screen.getByText('Heute')).toBeInTheDocument()
    expect(screen.getByText('Vorbereitung')).toBeInTheDocument()
    expect(screen.getByText('Nach dem Fest')).toBeInTheDocument()
    expect(screen.getByText('Service')).toBeInTheDocument()
  })

  it('zeigt bei geschlossener Kasse den neutralen Kassentag-Chip', async () => {
    await renderGeladeneSidebar()

    expect(screen.getByText('Kein Kassentag')).toBeInTheDocument()
    expect(screen.getByText('Kasse geschlossen')).toBeInTheDocument()
    expect(screen.queryByText('Kasse offen')).not.toBeInTheDocument()
  })

  it('zeigt bei offener Kasse Bezeichnung, Status und Eröffnungszeit im Chip', async () => {
    await renderGeladeneSidebar(backend({ kassensitzung: aktiveSitzung }))

    expect(screen.getByText('Sommerfest Tag 2')).toBeInTheDocument()
    expect(
      screen.getByText(`Kasse offen · seit ${erwarteteUhrzeit}`),
    ).toBeInTheDocument()
    expect(
      screen.getAllByRole('img', { name: 'Kasse offen' }).length,
    ).toBeGreaterThanOrEqual(1)
  })

  it('zeigt im Barrierestatus den unterbrochenen Abschluss statt „Kasse offen"', async () => {
    await renderGeladeneSidebar(
      backend({
        kassensitzung: { ...aktiveSitzung, status: 'wird_abgeschlossen' },
      }),
    )

    expect(
      screen.getByText('Abschluss unterbrochen — erneut abschließen'),
    ).toBeInTheDocument()
    expect(screen.queryByText(/Kasse offen/)).not.toBeInTheDocument()
    // Menüpunkt und Kopf-Chip tragen denselben Statuspunkt.
    expect(
      screen.getAllByRole('img', { name: 'Abschluss unterbrochen' }).length,
    ).toBe(2)
  })

  it('markiert Bondrucker bei fehlgeschlagenen Druckaufträgen', async () => {
    await renderGeladeneSidebar(
      backend({ druckauftraege: [druckauftrag(1), druckauftrag(2)] }),
    )

    expect(
      screen.getByRole('img', { name: 'Druckauftrag fehlgeschlagen' }),
    ).toBeInTheDocument()
  })

  it('markiert Finanzamt & TSE bei nicht konfigurierter TSE', async () => {
    await renderGeladeneSidebar(backend({ istKonfiguriert: false }))

    expect(
      screen.getByRole('img', { name: 'TSE benötigt Aufmerksamkeit' }),
    ).toBeInTheDocument()
  })

  it('markiert Finanzamt & TSE bei Signatur-Rückstand über der Schwelle', async () => {
    await renderGeladeneSidebar(
      backend({ queue: { ...ruhigeQueue, rueckstandSekunden: 120 } }),
    )

    expect(
      screen.getByRole('img', { name: 'TSE benötigt Aufmerksamkeit' }),
    ).toBeInTheDocument()
  })

  it('markiert Finanzamt & TSE nicht bei konfigurierter TSE ohne Rückstand', async () => {
    await renderGeladeneSidebar()

    expect(
      screen.queryByRole('img', { name: 'TSE benötigt Aufmerksamkeit' }),
    ).not.toBeInTheDocument()
  })

  it('beschriftet den Theme-Umschalter stabil, unabhängig vom aktiven Design', async () => {
    await renderGeladeneSidebar()
    expect(document.documentElement).toHaveClass('light')
    expect(
      screen.getByRole('button', { name: 'Design wechseln' }),
    ).toBeInTheDocument()
    expect(screen.queryByText(/Helles Design|Dunkles Design/)).toBeNull()

    cleanup()

    localStorage.setItem(THEME_STORAGE_KEY, 'dark')
    await renderGeladeneSidebar()
    expect(document.documentElement).toHaveClass('dark')
    expect(
      screen.getByRole('button', { name: 'Design wechseln' }),
    ).toBeInTheDocument()
    expect(screen.queryByText(/Helles Design|Dunkles Design/)).toBeNull()
  })

  it('schaltet auf das Gegenteil des aktuellen Designs', async () => {
    const user = userEvent.setup()
    await renderGeladeneSidebar()
    expect(document.documentElement).toHaveClass('light')

    await user.click(screen.getByRole('button', { name: 'Design wechseln' }))
    expect(document.documentElement).toHaveClass('dark')
    expect(localStorage.getItem(THEME_STORAGE_KEY)).toBe('dark')
  })
})
