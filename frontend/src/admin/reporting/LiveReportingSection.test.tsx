import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { LiveReportingSection } from './LiveReportingSection'
import type { LiveReportingData } from './types'

vi.mock('react-router', () => ({
  NavLink: ({ children, to }: { children?: ReactNode; to: string }) => (
    <a href={to}>{children}</a>
  ),
}))

afterEach(() => {
  cleanup()
})

function liveData(
  servicekraefte: LiveReportingData['breakdowns']['servicekraefte'],
  stornierungenProServicekraft: LiveReportingData['breakdowns']['stornierungenProServicekraft'] = [],
  stornierungen: LiveReportingData['stornierungen'] = [],
  overrides: Partial<LiveReportingData> = {},
): LiveReportingData {
  return {
    kassensitzungNr: 1,
    bezeichnung: 'Sommerfest',
    datum: '2026-06-18',
    offeneTische: [],
    offeneSaldiCents: 0,
    summary: {
      gesamtUmsatzCents: 2400,
      gesamtBestellungenCents: 3600,
      gesamtStornierungenCents: 0,
      geldtransitCents: 0,
      anzahlBestellungen: 5,
      anzahlStornierungen: 0,
      anzahlDirektverkaeufe: 2,
      direktverkaufUmsatzCents: 800,
    },
    breakdowns: { servicekraefte, stornierungenProServicekraft },
    stornierungen,
    produktStatistik: [],
    ...overrides,
  }
}

const noopRefresh = () => {
  // no-op
}

describe('LiveReportingSection — Übersicht', () => {
  it('zeigt die Hero-Kennzahl „Kassierter Umsatz" und die vier Nebenkarten', () => {
    render(
      <LiveReportingSection
        liveData={liveData([])}
        loading={false}
        dataUpdatedAt={0}
        onRefresh={noopRefresh}
      />,
    )

    // Hero: kassierter Umsatz mit erklärender Unterzeile.
    expect(screen.getByText('Kassierter Umsatz')).toBeInTheDocument()
    expect(screen.getByText('24,00 €')).toBeInTheDocument()
    expect(
      screen.getByText('bereits bezahlt, Stornos abgezogen'),
    ).toBeInTheDocument()

    // Nebenkarten mit Handoff-Unterzeilen.
    expect(screen.getByText('Noch offen')).toBeInTheDocument()
    expect(screen.getByText('Bestellt gesamt')).toBeInTheDocument()
    expect(screen.getByText('bezahlt + offen zusammen')).toBeInTheDocument()
    expect(screen.getByText('Direktverkauf')).toBeInTheDocument()
    expect(screen.getByText('2 Verkäufe ohne Tisch')).toBeInTheDocument()
    expect(screen.getByText('Storniert')).toBeInTheDocument()
  })

  it('zeigt offene Arbeit einer Servicekraft als Euro-Betrag mit Tischanzahl', () => {
    render(
      <LiveReportingSection
        liveData={liveData([
          {
            userId: 7,
            userName: 'Anna',
            name: 'Anna A.',
            zahlungenCents: 1500,
            offenCents: 750,
            offeneTische: [
              { tischId: 3, tischName: 'Tisch 3' },
              { tischId: 4, tischName: 'Zelt A2' },
            ],
            erledigt: false,
          },
          {
            userId: 9,
            userName: 'Cleo',
            name: '',
            zahlungenCents: 900,
            offenCents: 0,
            offeneTische: [],
            erledigt: true,
          },
        ])}
        loading={false}
        dataUpdatedAt={0}
        onRefresh={noopRefresh}
      />,
    )

    // Keine Tabs mehr.
    expect(screen.queryByRole('tab')).not.toBeInTheDocument()

    // Servicekraft mit offener Arbeit: Euro-Betrag (5,00 + 2,50 = 7,50 €) und
    // die Tischnamen inline.
    expect(screen.getByText('Anna (Anna A.)')).toBeInTheDocument()
    expect(screen.getByText('7,50 €')).toBeInTheDocument()
    expect(screen.getByText(/Tisch 3, Zelt A2/)).toBeInTheDocument()

    // Fertige Servicekraft: Abrechnungs-Hinweis.
    expect(screen.getByText('Cleo')).toBeInTheDocument()
    expect(screen.getByText('Alles abgerechnet')).toBeInTheDocument()

    // Der kassierte Betrag bleibt sichtbar.
    expect(screen.getByText('15,00 €')).toBeInTheDocument()
  })

  it('kürzt die Liste offener Tische nach fünf Einträgen und blendet den Rest ein', async () => {
    const user = userEvent.setup()
    const offeneTische = Array.from({ length: 7 }, (_, i) => ({
      tischId: i + 1,
      tischName: `Tisch ${String(i + 1)}`,
      saldoCents: 100 * (i + 1),
    }))

    render(
      <LiveReportingSection
        liveData={liveData([], [], [], {
          offeneTische,
          offeneSaldiCents: 2800,
        })}
        loading={false}
        dataUpdatedAt={0}
        onRefresh={noopRefresh}
      />,
    )

    // Nur fünf Zeilen sichtbar, Tisch 6/7 verborgen.
    expect(screen.getByText('Tisch 5')).toBeInTheDocument()
    expect(screen.queryByText('Tisch 6')).not.toBeInTheDocument()

    // „Alle 7 anzeigen" blendet den Rest ein.
    await user.click(screen.getByRole('button', { name: 'Alle 7 anzeigen' }))
    expect(screen.getByText('Tisch 6')).toBeInTheDocument()
    expect(screen.getByText('Tisch 7')).toBeInTheDocument()
  })

  it('zeigt die Stornierungen eingeklappt und expandiert zur Detail-Liste', async () => {
    const user = userEvent.setup()

    render(
      <LiveReportingSection
        liveData={liveData(
          [],
          [
            {
              userId: 3,
              userName: 'felix',
              name: 'Felix W.',
              anzahlStornierungen: 1,
              stornierungenCents: 500,
            },
          ],
          [
            {
              zeitpunkt: '2026-06-18T12:00:00Z',
              quelle: 'tisch',
              barRueckgabe: true,
              tischId: 9,
              tischName: 'Tisch 9',
              // Die Serviceleitung hat stellvertretend für felix storniert.
              akteur: { userId: 1, userName: 'lena', name: 'Lena C.' },
              betroffene: [{ userId: 3, userName: 'felix', name: 'Felix W.' }],
              betragCents: 500,
              kommentar: 'Falsch gebucht',
              positionen: [],
            },
          ],
          {
            summary: {
              gesamtUmsatzCents: 2400,
              gesamtBestellungenCents: 3600,
              gesamtStornierungenCents: 500,
              geldtransitCents: 0,
              anzahlBestellungen: 5,
              anzahlStornierungen: 1,
              anzahlDirektverkaeufe: 2,
              direktverkaufUmsatzCents: 800,
            },
          },
        )}
        loading={false}
        dataUpdatedAt={0}
        onRefresh={noopRefresh}
      />,
    )

    // Zusammenfassung sichtbar, Detail (Kommentar) zunächst verborgen.
    expect(screen.getByText('1 Stornierung')).toBeInTheDocument()
    expect(screen.queryByText('Falsch gebucht')).not.toBeInTheDocument()

    // Aufklappen zeigt die bestehende Detail-Liste: zuerst die betroffene
    // Servicekraft, der stellvertretende Akteur nur als Zusatz.
    await user.click(screen.getByRole('button', { name: /Details/ }))
    expect(screen.getByText('Falsch gebucht')).toBeInTheDocument()
    expect(screen.getByText('Tisch 9 · felix (Felix W.)')).toBeInTheDocument()
    expect(
      screen.getByText(/storniert von lena \(Lena C\.\)/),
    ).toBeInTheDocument()
  })

  it('zeigt „Live · aktualisiert HH:MM" und löst den Aktualisieren-Button aus', async () => {
    const user = userEvent.setup()
    const onRefresh = vi.fn()
    const stand = new Date('2026-06-18T14:05:00Z').getTime()

    render(
      <LiveReportingSection
        liveData={liveData([])}
        loading={false}
        dataUpdatedAt={stand}
        onRefresh={onRefresh}
      />,
    )

    expect(
      screen.getByText(/^Live · aktualisiert \d{2}:\d{2}$/),
    ).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Aktualisieren' }))
    expect(onRefresh).toHaveBeenCalledTimes(1)
  })

  it('zeigt den Leerzustand ohne Kassensitzung mit Link zur Kasse', () => {
    render(
      <LiveReportingSection
        liveData={null}
        loading={false}
        dataUpdatedAt={0}
        onRefresh={noopRefresh}
      />,
    )

    expect(screen.getByText('Keine Kassensitzung geöffnet')).toBeInTheDocument()
    expect(
      screen.getByRole('link', { name: 'Zur Kassensitzungs-Seite' }),
    ).toHaveAttribute('href', '/admin/kasse')
  })
})

describe('LiveReportingSection — Verkäufe pro Produkt', () => {
  it('zeigt die Produkt-/Varianten-Statistik der offenen Sitzung', () => {
    render(
      <LiveReportingSection
        liveData={liveData([], [], [], {
          produktStatistik: [
            {
              kategorie: 'essen',
              produktName: 'Pommes',
              // Bestellt/ausgegeben, aber noch nicht kassiert: Menge > 0, Umsatz 0.
              ausgegebeneMenge: 6,
              umsatzCents: 0,
              varianten: [
                {
                  varianteId: 10,
                  varianteName: 'groß',
                  ausgegebeneMenge: 6,
                  umsatzCents: 0,
                },
              ],
            },
          ],
        })}
        loading={false}
        dataUpdatedAt={0}
        onRefresh={noopRefresh}
      />,
    )

    expect(screen.getByText('Verkäufe pro Produkt')).toBeInTheDocument()
    expect(screen.getByText('Essen')).toBeInTheDocument()
    expect(screen.getByText('Pommes groß')).toBeInTheDocument()
  })

  it('zeigt einen leeren Zustand ohne Verkäufe', () => {
    render(
      <LiveReportingSection
        liveData={liveData([])}
        loading={false}
        dataUpdatedAt={0}
        onRefresh={noopRefresh}
      />,
    )

    expect(
      screen.getByText('Keine Verkäufe in dieser Kassensitzung.'),
    ).toBeInTheDocument()
  })
})
