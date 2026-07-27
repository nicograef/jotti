import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { ReportingResults } from './ReportingResults'
import type { AbgeschlosseneSitzung, ReportingData } from './types'

const reportingResult: ReportingData = {
  kassensitzungNr: 1,
  metadaten: {
    eroeffnetAm: '2026-07-05T07:58:00Z',
    abgeschlossenAm: '2026-07-05T21:12:00Z',
    abgeschlossenVon: 'nico',
    kassensturzDifferenzCents: -150,
  },
  summary: {
    gesamtUmsatzCents: 12345,
    gesamtBestellungenCents: 6500,
    gesamtStornierungenCents: 300,
    geldtransitCents: 0,
    anzahlBestellungen: 8,
    anzahlStornierungen: 1,
    anzahlDirektverkaeufe: 7,
    direktverkaufUmsatzCents: 4500,
  },
  breakdowns: {
    abrechnungProServicekraft: [
      {
        userId: 5,
        userName: 'Bea',
        name: 'Bea B.',
        kassiertCents: 6789,
        anzahlZahlungen: 3,
        ruecknahmenCents: 300,
        anzahlStornierungen: 1,
        abzugebenCents: 6489,
      },
    ],
  },
  umsatzProSteuersatz: [
    {
      satz: 'regel',
      bruttoCents: 1190,
      nettoCents: 1000,
      steuerCents: 190,
    },
  ],
  stornierungen: [
    {
      zeitpunkt: '2026-07-05T12:00:00Z',
      quelle: 'tisch',
      barRueckgabe: true,
      tischId: 4,
      tischName: 'Tisch 4',
      akteur: { userId: 5, userName: 'Bea', name: 'Bea B.' },
      betroffene: [{ userId: 5, userName: 'Bea', name: 'Bea B.' }],
      betragCents: 300,
      kommentar: '',
      positionen: [],
    },
  ],
  produktStatistik: [
    {
      kategorie: 'essen',
      produktName: 'Pommes',
      ausgegebeneMenge: 9,
      umsatzCents: 2500,
      varianten: [
        {
          varianteId: 10,
          varianteName: 'groß',
          ausgegebeneMenge: 5,
          umsatzCents: 1500,
        },
        {
          varianteId: 11,
          varianteName: 'klein',
          ausgegebeneMenge: 4,
          umsatzCents: 1000,
        },
      ],
    },
    {
      kategorie: 'getraenk',
      produktName: 'Cola',
      ausgegebeneMenge: 3,
      umsatzCents: 900,
      varianten: [
        {
          varianteId: 20,
          varianteName: '0,5 l',
          ausgegebeneMenge: 3,
          umsatzCents: 900,
        },
      ],
    },
  ],
}

const sitzung: AbgeschlosseneSitzung = {
  zNr: 11,
  datum: '2026-07-05',
  bezeichnung: 'Sommerfest Tag 1',
  umsatzGesamtCents: 12345,
  abgeschlossenAm: '2026-07-05T21:12:00Z',
}

afterEach(() => {
  cleanup()
})

describe('ReportingResults', () => {
  it('zeigt den formalen Berichtskopf mit Nr., Bezeichnung und Metadaten', () => {
    render(
      <ReportingResults
        result={reportingResult}
        sitzung={sitzung}
        loading={false}
      />,
    )

    expect(
      screen.getByRole('heading', {
        name: 'Tagesbericht Nr. 11 — Sommerfest Tag 1',
      }),
    ).toBeInTheDocument()
    // Metadaten-Zeile: abschließender Benutzer und Kassensturz-Differenz.
    expect(screen.getByText(/von nico/)).toBeInTheDocument()
    expect(
      screen.getByText(/Kassensturz-Differenz -1,50 €/),
    ).toBeInTheDocument()
  })

  it('zeigt die vier Kennzahl-Kacheln', () => {
    render(
      <ReportingResults
        result={reportingResult}
        sitzung={sitzung}
        loading={false}
      />,
    )

    expect(screen.getByText('Kassierter Umsatz')).toBeInTheDocument()
    expect(screen.getByText('Bestellungen')).toBeInTheDocument()
    expect(screen.getByText('Direktverkauf')).toBeInTheDocument()
    expect(screen.getByText('Storniert')).toBeInTheDocument()
    expect(screen.getByText('45,00 €')).toBeInTheDocument()
  })

  it('zeigt Steuersatz-Tabelle, Servicekräfte und Stornierungen ohne Tabs untereinander', () => {
    render(
      <ReportingResults
        result={reportingResult}
        sitzung={sitzung}
        loading={false}
      />,
    )

    // Keine Tabs mehr: alle Abschnitte gleichzeitig sichtbar.
    expect(screen.queryByRole('tab')).not.toBeInTheDocument()
    expect(screen.getByText('Umsatz nach Steuersatz')).toBeInTheDocument()
    expect(screen.getByText('Regelsteuersatz (19 %)')).toBeInTheDocument()
    expect(screen.getByText('Abrechnung pro Servicekraft')).toBeInTheDocument()
    expect(screen.getByText('Bea (Bea B.)')).toBeInTheDocument()
  })

  it('zeigt pro Servicekraft „Abzugeben" als Hauptzahl mit Kassiert und Rücknahmen darunter', () => {
    render(
      <ReportingResults
        result={reportingResult}
        sitzung={sitzung}
        loading={false}
      />,
    )

    // Hauptzahl: Abzugeben (67,89 € kassiert − 3,00 € Rücknahmen).
    expect(screen.getByText('64,89 €')).toBeInTheDocument()
    // Nebenzeile: die Herleitung des Abzugs.
    expect(
      screen.getByText('Kassiert 67,89 € · Rücknahmen 3,00 €'),
    ).toBeInTheDocument()
    // Die Unterzeile grenzt den Direktverkauf aus.
    expect(
      screen.getByText(/Direktverkäufe sind nicht enthalten/),
    ).toBeInTheDocument()
  })

  it('setzt den Storno-Marker bei der betroffenen, nicht bei der stornierenden Servicekraft', () => {
    render(
      <ReportingResults
        result={{
          ...reportingResult,
          breakdowns: {
            abrechnungProServicekraft: [
              // Bea hat kassiert und die Rücknahme zugeordnet bekommen.
              {
                userId: 5,
                userName: 'Bea',
                name: 'Bea B.',
                kassiertCents: 6789,
                anzahlZahlungen: 3,
                ruecknahmenCents: 300,
                anzahlStornierungen: 1,
                abzugebenCents: 6489,
              },
              // Lena hat stellvertretend storniert — ohne eigene Zuordnung.
              {
                userId: 1,
                userName: 'lena',
                name: 'Lena C.',
                kassiertCents: 1000,
                anzahlZahlungen: 1,
                ruecknahmenCents: 0,
                anzahlStornierungen: 0,
                abzugebenCents: 1000,
              },
            ],
          },
        }}
        sitzung={sitzung}
        loading={false}
      />,
    )

    const marker = screen.getByText('1 Storno')
    expect(marker).toBeInTheDocument()
    // Der Marker steht in der Zeile von Bea, nicht in der von Lena.
    const zeile = marker.closest('div')
    expect(zeile).toHaveTextContent('Bea (Bea B.)')
    expect(zeile).not.toHaveTextContent('lena')
  })

  it('zeigt den Abschnitt „Verkäufe pro Produkt" mit Kategorien, Zwischensumme und Ein-Varianten-Zeile', () => {
    render(
      <ReportingResults
        result={reportingResult}
        sitzung={sitzung}
        loading={false}
      />,
    )

    expect(screen.getByText('Verkäufe pro Produkt')).toBeInTheDocument()
    // Kategorie-Überschriften.
    expect(screen.getByText('Essen')).toBeInTheDocument()
    expect(screen.getByText('Getränke')).toBeInTheDocument()
    // Mehr-Varianten-Produkt: Produktzeile plus zwei Variantenzeilen.
    expect(screen.getByText('Pommes')).toBeInTheDocument()
    expect(screen.getByText('groß')).toBeInTheDocument()
    expect(screen.getByText('klein')).toBeInTheDocument()
    // Ein-Varianten-Produkt zu einer Zeile zusammengefasst (produktName ==
    // varianteName erscheint genau einmal).
    expect(screen.getByText('Cola 0,5 l')).toBeInTheDocument()
    // Spaltenüberschriften und ein Umsatzwert.
    expect(screen.getByText('Ausgegeben')).toBeInTheDocument()
    expect(screen.getByText('25,00 €')).toBeInTheDocument()
  })

  it('zeigt einen leeren Zustand ohne Verkäufe', () => {
    render(
      <ReportingResults
        result={{ ...reportingResult, produktStatistik: [] }}
        sitzung={sitzung}
        loading={false}
      />,
    )

    expect(
      screen.getByText('Keine Verkäufe in dieser Kassensitzung.'),
    ).toBeInTheDocument()
  })

  it('nennt in der Storno-Zeile die betroffene Servicekraft ohne Akteurs-Zusatz, wenn sie selbst storniert hat', () => {
    render(
      <ReportingResults
        result={reportingResult}
        sitzung={sitzung}
        loading={false}
      />,
    )

    expect(screen.getByText('Tisch 4 · Bea (Bea B.)')).toBeInTheDocument()
    expect(screen.queryByText(/storniert von/)).not.toBeInTheDocument()
  })

  it('nennt bei stellvertretender Stornierung die betroffenen Servicekräfte und den Akteur als Zusatz', () => {
    render(
      <ReportingResults
        result={{
          ...reportingResult,
          stornierungen: [
            {
              ...reportingResult.stornierungen[0],
              akteur: { userId: 1, userName: 'lena', name: 'Lena C.' },
              betroffene: [
                { userId: 5, userName: 'Bea', name: 'Bea B.' },
                { userId: 6, userName: 'tom', name: '' },
              ],
            },
          ],
        }}
        sitzung={sitzung}
        loading={false}
      />,
    )

    expect(screen.getByText('Tisch 4 · Bea (Bea B.), tom')).toBeInTheDocument()
    expect(
      screen.getByText(/storniert von lena \(Lena C\.\)/),
    ).toBeInTheDocument()
  })

  it('druckt beim Klick auf Drucken über window.print', async () => {
    const printSpy = vi.spyOn(window, 'print').mockImplementation(vi.fn())
    const { default: userEvent } = await import('@testing-library/user-event')
    const user = userEvent.setup()
    render(
      <ReportingResults
        result={reportingResult}
        sitzung={sitzung}
        loading={false}
      />,
    )

    await user.click(screen.getByRole('button', { name: 'Drucken' }))
    expect(printSpy).toHaveBeenCalledOnce()
    printSpy.mockRestore()
  })
})
