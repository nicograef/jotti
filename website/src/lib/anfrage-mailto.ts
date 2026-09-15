import { betreiberEmail, githubUrl } from './links'

export interface AnfrageFelder {
  verein: string
  sitz: string
  name: string
  email: string
  art: string
  message: string
}

// Das erste Label ist der Vorgabewert des Selects (AnfrageFormular.tsx).
export const artOptionen = [
  'Eingetragener Verein (e.V.)',
  'Gemeinnützige Stiftung',
  'NGO / NPO',
  'Sonstige gemeinnützige Organisation',
] as const

// art hat als Select immer einen Wert, message ist optional — beide werden
// nicht validiert.
export type AnfrageFehler = Partial<
  Record<'verein' | 'sitz' | 'name' | 'email', string>
>

// Nur Format (etwas@etwas.tld) — Zustellbarkeit prüft erst das Mailprogramm.
const emailMuster = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

export function validateAnfrage(felder: AnfrageFelder): AnfrageFehler {
  const fehler: AnfrageFehler = {}

  if (!felder.verein.trim()) {
    fehler.verein =
      'Bitte gib den Namen eures Vereins oder eurer Organisation an.'
  }
  // LICENSE verlangt den Sitz als Teil der Identifikation.
  if (!felder.sitz.trim()) {
    fehler.sitz = 'Bitte gib den Sitz eurer Organisation an.'
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

// Unencodiert: Quelle für buildMailtoUrl und für die Anzeige zum Kopieren.
export interface AnfrageMail {
  empfaenger: string
  betreff: string
  text: string
}

// Vorlage aus TERMS.md: Die Nutzungsvereinbarung kommt allein durch diese
// Annahme-E-Mail zustande, deshalb der wörtliche Annahmesatz mit Fassungsbezug
// (7. September 2026) und der TERMS-URL.
export function buildAnfrageMail(felder: AnfrageFelder): AnfrageMail {
  const verein = felder.verein.trim()
  const betreff = `Nutzungsvereinbarung jotti — ${verein}`

  const termsUrl = `${githubUrl}/blob/main/TERMS.md`

  const zeilen = [
    'Hallo Herr Gräf,',
    '',
    `wir sind ${verein} und akzeptieren die Nutzungsbedingungen für jotti in der Fassung vom 7. September 2026 (${termsUrl}).`,
    '',
    `Rechtsform: ${felder.art.trim()}`,
    `Sitz: ${felder.sitz.trim()}`,
    `Ansprechperson: ${felder.name.trim()}, ${felder.email.trim()}`,
  ]

  const nachricht = felder.message.trim()
  if (nachricht) {
    zeilen.push('', 'Nachricht:', nachricht)
  }

  zeilen.push('', 'Mit freundlichen Grüßen', felder.name.trim(), verein)

  return { empfaenger: betreiberEmail, betreff, text: zeilen.join('\n') }
}

export function buildMailtoUrl(felder: AnfrageFelder): string {
  const { empfaenger, betreff, text } = buildAnfrageMail(felder)

  return `mailto:${empfaenger}?subject=${encodeURIComponent(
    betreff,
  )}&body=${encodeURIComponent(text)}`
}
