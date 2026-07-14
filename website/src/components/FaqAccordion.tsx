import { useState } from 'react'
import { Plus } from 'lucide-react'

// FAQ-Accordion der Landing (Handoff-Prototyp docs/prds/design_handoff_jotti_website, #faq).
// Neun Items, single-open: das Öffnen eines Items schließt das zuvor offene.
// Disclosure-Pattern nach WAI-ARIA — jede Frage ist ein <button> mit
// aria-expanded und aria-controls, das Antwort-Panel eine per aria-labelledby
// benannte Region, im geschlossenen Zustand über das hidden-Attribut aus dem
// Accessibility-Baum genommen. Das Plus-Icon rotiert im offenen Zustand zu einem
// × (transform: rotate(45deg)). Statischer Sektionskopf liegt in Faq.astro; nur
// die Liste ist eine Island.
//
// Sieben Items aus dem Prototyp, ergänzt um zwei aus dem Plan (Phase 7):
//  - das Service-Item (bezahlte Unterstützung auf Anfrage) — Copy aus der
//    abgelösten #service-Sektion, als kostenpflichtig benannt (PRD).
//  - das Alternativen-Vergleichs-Item („Warum nicht Excel oder eine
//    Profi-Kasse?") — Kernaussage der abgelösten Vergleichstabelle.
// Die Installations-Antwort ist auf die reale Auslieferung abgestimmt (ZIP mit
// Starter + Docker Desktop, Leitfaden führt durch) — konsistent mit dem
// Download-Bereich, nicht mit dem im Prototyp versprochenen Doppelklick-Release.

interface FaqItem {
  q: string
  a: string
}

const faqs: FaqItem[] = [
  {
    q: 'Was kostet jotti?',
    a: 'Die Software selbst ist kostenlos — keine Lizenzgebühr, kein Cloud-Abo für jotti. Laufende Kosten entstehen nur für die gesetzlich vorgeschriebene Cloud-TSE von fiskaly und optional einen Server.',
  },
  {
    q: 'Warum nicht Excel oder eine Profi-Kasse?',
    a: 'Stift und Papier oder eine Excel-Tabelle bieten keine gemeinsame Tisch-Übersicht und erfüllen die gesetzlichen Kassenregeln (TSE, Belegausgabe) nicht. Eine Profi-Kasse kann das, ist aber für den täglichen Gastrobetrieb gemacht, kostet 30 bis 100 € im Monat plus Hardware und oft mit Vertragslaufzeit — für zwei, drei Feste im Jahr lohnt sich das selten. jotti ist kostenlos, läuft auf euren eigenen Smartphones und bringt die fiskalischen Bausteine ab Werk mit.',
  },
  {
    q: 'Ist jotti wirklich TSE-konform?',
    a: 'jotti bringt die fiskalischen Bausteine mit: eine BSI-zertifizierte Cloud-TSE von fiskaly, Belegausgabe nach § 146a AO, ein append-only Kassenjournal (GoBD) und den DSFinV-K-Export v2.4. Den konformen Betrieb — TSE-Vertrag, Kassenmeldung, Aufbewahrung — verantwortet der Verein als Betreiber.',
  },
  {
    q: 'Brauchen wir spezielle Hardware?',
    a: 'Nein. jotti läuft im Browser auf den Smartphones, Tablets und Rechnern, die ihr schon habt (BYOD). Für Küchenbons genügt ein handelsüblicher ESC/POS-Bondrucker (80 mm, Ethernet).',
  },
  {
    q: 'Können die Helfer:innen ihre eigenen Handys nutzen?',
    a: 'Genau dafür ist jotti gebaut. Servicekräfte öffnen die Adresse im Browser, melden sich per Einmalpasswort an und kassieren sofort — die App lässt sich als Icon auf den Startbildschirm legen (PWA), ganz ohne App Store.',
  },
  {
    q: 'Wie installieren wir jotti?',
    a: 'Ihr ladet das Release als ZIP von der GitHub-Releases-Seite herunter, entpackt es und startet den enthaltenen Starter — dafür braucht ihr Docker Desktop auf dem Rechner. Der Leitfaden für Vereine führt euch Schritt für Schritt durch die Einrichtung.',
  },
  {
    q: 'Gibt es Unterstützung beim Einrichten?',
    a: 'Ja. Die Software bleibt kostenlos — wer sich Einrichtung, Hosting oder Schulung der Helfer:innen nicht allein zutraut, kann diese Unterstützung auf Anfrage kostenpflichtig dazubuchen. Sie ist freiwillig und unabhängig von jotti; die Verantwortung für den Betrieb bleibt bei eurem Verein. Schreibt einfach kurz, was ihr braucht.',
  },
  {
    q: 'Für welche Vereine ist jotti gedacht?',
    a: 'Für eingetragene Vereine (e.V.), gemeinnützige Organisationen und NPOs mit temporären Veranstaltungen — Vereinsfeste, Sommerfeste, Weihnachtsmärkte, Konzerte. Nicht für Dauerbetrieb (Restaurants, Cafés) oder Kartenzahlung.',
  },
  {
    q: 'Ist der Quellcode offen?',
    a: 'Der Quellcode ist öffentlich einsehbar (source-available) — zum Lernen, Evaluieren und für Sicherheitsprüfungen. jotti ist kein Open-Source-Projekt im OSI-Sinn; jede Nutzung setzt eine kostenlose Nutzungsvereinbarung mit dem Autor voraus.',
  },
]

export default function FaqAccordion() {
  // Single-open: Index des offenen Items, -1 wenn alle geschlossen. Standard
  // offen ist das erste Item (Referenz 09-light/04-dark) — im SSR-HTML ist so
  // die erste Antwort auch ohne JavaScript sichtbar.
  const [open, setOpen] = useState(0)

  return (
    <div className="mt-10 flex flex-col gap-3">
      {faqs.map((item, index) => {
        const isOpen = index === open
        const buttonId = `faq-button-${index}`
        const panelId = `faq-panel-${index}`
        return (
          <div
            key={item.q}
            className="overflow-hidden rounded-[15px] border transition-colors"
            style={{
              borderColor: isOpen ? 'var(--primary)' : 'var(--border)',
              background: isOpen
                ? 'color-mix(in srgb, var(--primary) 4%, var(--card))'
                : 'var(--card)',
            }}
          >
            <h3 className="m-0">
              <button
                type="button"
                id={buttonId}
                aria-expanded={isOpen}
                aria-controls={panelId}
                onClick={() => setOpen(isOpen ? -1 : index)}
                className="flex w-full items-center justify-between gap-4 px-[22px] py-[19px] text-left focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-brand"
              >
                <span className="font-brand text-[17px] font-semibold">
                  {item.q}
                </span>
                <span
                  className="shrink-0 text-muted transition-transform duration-200"
                  style={{ transform: isOpen ? 'rotate(45deg)' : 'none' }}
                  aria-hidden="true"
                >
                  <Plus size={20} />
                </span>
              </button>
            </h3>
            <div
              id={panelId}
              role="region"
              aria-labelledby={buttonId}
              hidden={!isOpen}
              className="max-w-[64ch] px-[22px] pb-[21px] text-[15px] leading-[1.62] text-muted"
            >
              {item.a}
            </div>
          </div>
        )
      })}
    </div>
  )
}
