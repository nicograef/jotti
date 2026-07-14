import { useRef, useState, type KeyboardEvent } from 'react'
import {
  Calculator,
  ChartColumn,
  Check,
  Printer,
  Receipt,
  ShoppingBag,
  Wallet,
  type LucideIcon,
} from 'lucide-react'

// Interaktiver Feature-Explorer (Handoff-Prototyp, PRD docs/prds/prd-website-redesign.md).
// Sechs Bereichs-Tiles mit je eigenem Spektral-Akzent; ein Bereich ist aktiv und
// füllt die sticky Detail-Karte. Umsetzung nach dem WAI-ARIA-Tabs-Pattern:
// role=tablist/tab/tabpanel, roving tabindex, Pfeiltasten + Home/End, automatische
// Aktivierung (Auswahl folgt dem Fokus). Statischer Sektionskopf (Eyebrow, H2,
// Intro) liegt in Features.astro; nur das interaktive Raster ist eine Island.
//
// Icon-Bedeutungen bewusst nach Handoff-Vorgabe (README): Bestellung = Beleg,
// Zahlung = Geldbörse (NICHT Kartenterminal), Direktverkauf = Einkaufstasche,
// Küche = Drucker, Kasse = Registrierkasse, Reporting = Balkendiagramm.
//
// Copy gegen docs/anforderungen.md geprüft: Im Reporting-Bereich entfällt die
// „Abrechnung pro Tisch" (R-03 per ADR 02 ersatzlos entfernt) zugunsten von
// „pro Servicekraft" (R-04); keine „In Entwicklung"-Markierungen.

interface Feature {
  Icon: LucideIcon
  accent: string
  title: string
  short: string
  detail: string
  points: [string, string, string]
}

const features: Feature[] = [
  {
    Icon: Receipt,
    accent: 'var(--sp-red)',
    title: 'Bestellungen',
    short: 'Pro Tisch aufnehmen',
    detail:
      'Bestellungen auf Tische buchen — mit Produkten, Varianten, Steuersätzen und Kommentaren. Umbuchen auf andere Tische inklusive.',
    points: [
      'Produkte & Varianten',
      'Steuersätze pro Position',
      'Kommentare & Umbuchung',
    ],
  },
  {
    Icon: Wallet,
    accent: 'var(--sp-orange)',
    title: 'Zahlung & Storno',
    short: 'Kassieren mit Rückgeld',
    detail:
      'Zahlungen kassieren, auch Teilzahlungen mit automatischer Rückgeldberechnung. Stornos nur für Admin & Serviceleitung — mit Pflichtkommentar.',
    points: [
      'Teilzahlung & Rückgeld',
      'Rollenbasierter Storno',
      'Pflichtkommentar',
    ],
  },
  {
    Icon: ShoppingBag,
    accent: 'var(--sp-green)',
    title: 'Direktverkauf',
    short: 'Theke ohne Tisch',
    detail:
      'An der Theke bestellen, kassieren und ausgeben in einem Schritt — ohne Tisch, mit Historie und Storno.',
    points: ['Ein-Schritt-Verkauf', 'Volle Historie', 'Storno möglich'],
  },
  {
    Icon: Printer,
    accent: 'var(--sp-teal)',
    title: 'Küche & Bon-Druck',
    short: 'Bons automatisch',
    detail:
      'Bestell- und Küchenbons gehen automatisch an die zugeordneten ESC/POS-Bondrucker — pro Kategorie konfigurierbar.',
    points: ['ESC/POS-Drucker', 'Pro Kategorie', 'Automatischer Versand'],
  },
  {
    Icon: Calculator,
    accent: 'var(--sp-blue)',
    title: 'Kasse & Abschluss',
    short: 'Sitzung bis Z-Bon',
    detail:
      'Kassensitzungen eröffnen und schließen, Anfangsbestand erfassen, Kassensturz mit Differenz und formaler Tagesabschluss (Z-Bon).',
    points: ['Kassensitzung', 'Kassensturz', 'Tagesabschluss (Z-Bon)'],
  },
  {
    Icon: ChartColumn,
    accent: 'var(--sp-violet)',
    title: 'Reporting & Export',
    short: 'Bis DSFinV-K',
    detail:
      'Tagesabrechnung nach Steuersatz und pro Servicekraft — plus maschinenlesbarer DSFinV-K-Export (v2.4).',
    points: ['Umsätze nach Steuersatz', 'Pro Servicekraft', 'DSFinV-K v2.4'],
  },
]

export default function FeatureExplorer() {
  const [active, setActive] = useState(0)
  const tabRefs = useRef<(HTMLButtonElement | null)[]>([])

  function focusTab(index: number) {
    const next = (index + features.length) % features.length
    setActive(next)
    tabRefs.current[next]?.focus()
  }

  // Pfeiltasten (beide Achsen, da 2×3-Raster), Home/End; Auswahl folgt dem Fokus.
  function onKeyDown(event: KeyboardEvent<HTMLButtonElement>, index: number) {
    switch (event.key) {
      case 'ArrowRight':
      case 'ArrowDown':
        event.preventDefault()
        focusTab(index + 1)
        break
      case 'ArrowLeft':
      case 'ArrowUp':
        event.preventDefault()
        focusTab(index - 1)
        break
      case 'Home':
        event.preventDefault()
        focusTab(0)
        break
      case 'End':
        event.preventDefault()
        focusTab(features.length - 1)
        break
    }
  }

  const activeFeature = features[active]

  return (
    <div className="mt-10 grid items-start gap-7 nav:grid-cols-[1.15fr_1fr]">
      <div
        role="tablist"
        aria-label="Funktionsbereiche"
        className="grid grid-cols-2 gap-3"
      >
        {features.map((feature, index) => {
          const isActive = index === active
          const { Icon, accent } = feature
          return (
            <button
              key={feature.title}
              ref={(element) => {
                tabRefs.current[index] = element
              }}
              type="button"
              role="tab"
              id={`feature-tab-${index}`}
              aria-selected={isActive}
              aria-controls="feature-panel"
              tabIndex={isActive ? 0 : -1}
              onClick={() => setActive(index)}
              onKeyDown={(event) => onKeyDown(event, index)}
              className="flex flex-col rounded-2xl border p-[18px] text-left transition-[transform,box-shadow,border-color] duration-200 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand"
              style={
                isActive
                  ? {
                      borderColor: accent,
                      boxShadow: `0 0 0 1px ${accent}, var(--shadow)`,
                      transform: 'translateY(-3px)',
                      background: `color-mix(in srgb, ${accent} 7%, var(--card))`,
                    }
                  : { borderColor: 'var(--border)', background: 'var(--card)' }
              }
            >
              <span
                className="flex h-[42px] w-[42px] items-center justify-center rounded-xl"
                style={{
                  background: `color-mix(in srgb, ${accent} 15%, transparent)`,
                  color: accent,
                }}
              >
                <Icon size={21} aria-hidden="true" />
              </span>
              <span className="font-brand mt-[13px] text-[16.5px] font-semibold">
                {feature.title}
              </span>
              <span className="mt-[3px] text-[13px] text-muted">
                {feature.short}
              </span>
            </button>
          )
        })}
      </div>

      <div
        id="feature-panel"
        role="tabpanel"
        aria-labelledby={`feature-tab-${active}`}
        tabIndex={0}
        className="relative overflow-hidden rounded-[20px] border border-card-border bg-card p-[30px] shadow-[var(--shadow)] nav:sticky nav:top-[92px]"
      >
        <div
          className="absolute inset-x-0 top-0 h-[5px]"
          style={{ background: 'var(--spectral)' }}
          aria-hidden="true"
        ></div>
        <span
          className="flex h-[58px] w-[58px] items-center justify-center rounded-2xl"
          style={{
            background: `color-mix(in srgb, ${activeFeature.accent} 15%, transparent)`,
            color: activeFeature.accent,
          }}
        >
          <activeFeature.Icon size={28} aria-hidden="true" />
        </span>
        <h3 className="font-brand mt-[18px] text-[25px] font-bold tracking-[-0.02em]">
          {activeFeature.title}
        </h3>
        <p className="mt-2.5 text-[15.5px] leading-relaxed text-muted">
          {activeFeature.detail}
        </p>
        <ul className="mt-[22px] flex flex-col gap-2.5">
          {activeFeature.points.map((point) => (
            <li
              key={point}
              className="flex items-center gap-2.5 text-[14.5px] font-medium"
            >
              <span
                className="flex h-[22px] w-[22px] shrink-0 items-center justify-center rounded-lg text-brand"
                style={{
                  background:
                    'color-mix(in srgb, var(--primary) 14%, transparent)',
                }}
              >
                <Check size={13} aria-hidden="true" />
              </span>
              {point}
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}
