// UI-freies Logik-Modul des Anfrage-Formulars (/fuer-vereine).
// Baut aus den Feldwerten eine korrekt encodierte mailto-URL (Empfänger ist die
// Betreiber-Adresse aus links.ts, Betreff und Annahme-E-Mail-Body nach der
// TERMS.md-Vorlage) und validiert die Pflichtfelder. Kein DOM, keine React-Abhängigkeit —
// die AnfrageFormular-Island (src/components/AnfrageFormular.tsx) rendert die
// Felder, ruft dieses Modul auf und öffnet die URL per JS-Navigation (kein
// natives <form action="mailto:">, das die Produktiv-CSP form-action 'self'
// blockt).
//
// Feldnamen und Rechtsform-Labels stammen aus dem Handoff-Prototyp
// (PRD docs/prds/prd-website-redesign.md).

import { betreiberEmail, githubUrl } from './links'

export interface AnfrageFelder {
  verein: string
  name: string
  email: string
  art: string
  message: string
}

// Rechtsform-Auswahl mit den vollen Prototyp-Labels. Einzige Quelle für das
// Select der Island und den Body der mailto-URL, damit beide nicht auseinander
// laufen. Das erste Label ist der Vorgabewert des Selects.
export const artOptionen = [
  'Eingetragener Verein (e.V.)',
  'Gemeinnützige Stiftung',
  'NGO / NPO',
  'Sonstige gemeinnützige Organisation',
] as const

// Fehler je Pflichtfeld (verein, name, email); die Werte sind
// benutzer-sichtbare deutsche Meldungen. art hat als Select immer einen Wert,
// message ist optional — beide brauchen keine Validierung.
export type AnfrageFehler = Partial<
  Record<'verein' | 'name' | 'email', string>
>

// Einfacher Format-Check (etwas@etwas.tld); die eigentliche Zustellbarkeit
// prüft erst das Mailprogramm.
const emailMuster = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

export function validateAnfrage(felder: AnfrageFelder): AnfrageFehler {
  const fehler: AnfrageFehler = {}

  if (!felder.verein.trim()) {
    fehler.verein = 'Bitte gib den Namen eures Vereins oder eurer Organisation an.'
  }
  if (!felder.name.trim()) {
    fehler.name = 'Bitte gib eine:n Ansprechpartner:in an.'
  }

  const email = felder.email.trim()
  if (!email) {
    fehler.email = 'Bitte gib eure E-Mail-Adresse an.'
  } else if (!emailMuster.test(email)) {
    fehler.email = 'Bitte gib eine gültige E-Mail-Adresse ein.'
  }

  return fehler
}

export function hatFehler(fehler: AnfrageFehler): boolean {
  return Object.keys(fehler).length > 0
}

// Baut die mailto-URL: Empfänger als roher addr-spec im Pfad, Betreff und Body
// per encodeURIComponent (encodiert Umlaute, Zeilenumbrüche als %0A, Leerzeichen
// als %20 und Sonderzeichen wie & ? = +). Betreff und Body folgen der
// E-Mail-Vorlage aus TERMS.md: Die Nutzungsvereinbarung kommt durch diese eine
// Annahme-E-Mail zustande, deshalb enthält der Body den wörtlichen Annahmesatz
// mit Fassungsbezug (14. Juli 2026) und der TERMS-URL neben den Kontaktfeldern.
// Der optionale Nachrichten-Block entfällt, wenn keine Nachricht eingegeben wurde.
export function buildMailtoUrl(felder: AnfrageFelder): string {
  const verein = felder.verein.trim()
  const betreff = `Nutzungsvereinbarung jotti — ${verein}`

  const termsUrl = `${githubUrl}/blob/main/TERMS.md`

  const zeilen = [
    'Hallo Herr Gräf,',
    '',
    `wir sind ${verein} und akzeptieren die Nutzungsbedingungen für jotti in der Fassung vom 14. Juli 2026 (${termsUrl}).`,
    '',
    `Rechtsform: ${felder.art.trim()}`,
    `Ansprechperson: ${felder.name.trim()}, ${felder.email.trim()}`,
  ]

  const nachricht = felder.message.trim()
  if (nachricht) {
    zeilen.push('', 'Nachricht:', nachricht)
  }

  zeilen.push('', 'Mit freundlichen Grüßen', felder.name.trim(), verein)

  const body = zeilen.join('\n')

  return `mailto:${betreiberEmail}?subject=${encodeURIComponent(
    betreff,
  )}&body=${encodeURIComponent(body)}`
}
