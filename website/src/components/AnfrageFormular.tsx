import { useId, useRef, useState } from 'react'
import type { SyntheticEvent } from 'react'
import { ArrowRight, Check, Copy } from 'lucide-react'

import type { AnfrageFehler, AnfrageFelder } from '../lib/anfrage-mailto'
import {
  artOptionen,
  buildAnfrageMail,
  buildMailtoUrl,
  hatFehler,
  validateAnfrage,
} from '../lib/anfrage-mailto'
import { installationUrl } from '../lib/links'

// AnfrageFormular-Island der Seite /fuer-vereine (data-vereine-Formular).
// Rendert die Formularfelder, validiert clientseitig über
// src/lib/anfrage-mailto.ts, öffnet bei gültigem Absenden den vorbefüllten
// mailto-Entwurf per JS-Navigation
// (window.location.href — bewusst kein natives <form action="mailto:">, das
// die Produktiv-CSP form-action 'self' blockt) und wechselt in einen ehrlichen
// Erfolgs-State: der Entwurf ist geöffnet und muss noch gesendet werden. Der
// Erfolgs-State zeigt zusätzlich Empfänger, Betreff und den vollen Mailtext
// (aus buildAnfrageMail) in einem readOnly-Textfeld mit Kopieren-Button — für
// Geräte ohne Mailprogramm oder wenn sich kein Entwurf öffnet.
//
// Fehler sind programmatisch verknüpft (aria-invalid + aria-describedby am
// Feld) und werden zusätzlich über eine assertive Live-Region angekündigt. Die
// Erfolgs-Animation nutzt die geteilte .demo-pop-Klasse aus landing.css, die
// unter prefers-reduced-motion neutralisiert ist.

// Gemeinsame Feld-Optik (Handoff): Höhe, Radius, Rahmen, Fokus-Ring.
const feldKlassen =
  'w-full rounded-[11px] border border-card-border bg-background px-3.5 text-[15px] text-foreground outline-none transition-colors focus:border-brand focus:ring-[3px] focus:ring-[color:var(--ring)]'

const leereFelder: AnfrageFelder = {
  verein: '',
  name: '',
  email: '',
  art: artOptionen[0],
  message: '',
}

export default function AnfrageFormular() {
  const [felder, setFelder] = useState<AnfrageFelder>(leereFelder)
  const [fehler, setFehler] = useState<AnfrageFehler>({})
  const [gesendet, setGesendet] = useState(false)
  // Text der Live-Region; bei fehlgeschlagenem Absenden angekündigt.
  const [ankuendigung, setAnkuendigung] = useState('')
  // Zustand des Kopieren-Buttons im Erfolgs-State.
  const [kopieren, setKopieren] = useState<'idle' | 'kopiert' | 'fehler'>(
    'idle',
  )

  const formRef = useRef<HTMLFormElement>(null)
  // Eindeutige Präfixe, damit mehrere Instanzen kollisionsfrei blieben und die
  // aria-describedby-Verweise stabil sind.
  const uid = useId()

  function setFeld<K extends keyof AnfrageFelder>(
    key: K,
    value: AnfrageFelder[K],
  ) {
    setFelder((current) => ({ ...current, [key]: value }))
    // Fehler des gerade bearbeiteten Felds sofort aufheben.
    if (key in fehler) {
      setFehler((current) => {
        const next = { ...current }
        delete next[key as keyof AnfrageFehler]
        return next
      })
    }
  }

  function onSubmit(event: SyntheticEvent<HTMLFormElement>) {
    event.preventDefault()
    const gefunden = validateAnfrage(felder)
    if (hatFehler(gefunden)) {
      setFehler(gefunden)
      setAnkuendigung(
        'Der E-Mail-Entwurf konnte nicht geöffnet werden. Bitte fülle die markierten Pflichtfelder aus.',
      )
      // Fokus auf das erste fehlerhafte Feld.
      const erstesFeld = (['verein', 'name', 'email'] as const).find(
        (key) => gefunden[key],
      )
      if (erstesFeld) {
        formRef.current
          ?.querySelector<HTMLElement>(`[name="${erstesFeld}"]`)
          ?.focus()
      }
      return
    }
    setAnkuendigung('')
    // JS-Navigation zum vorbefüllten Entwurf (kein form-action-Verstoß).
    window.location.href = buildMailtoUrl(felder)
    setKopieren('idle')
    setGesendet(true)
  }

  if (gesendet) {
    const mail = buildAnfrageMail(felder)

    // navigator.clipboard fehlt in unsicheren Kontexten und alten Browsern
    // und wirft dann synchron — deshalb try/catch statt .then/.catch.
    async function kopiereText() {
      try {
        await navigator.clipboard.writeText(mail.text)
        setKopieren('kopiert')
        window.setTimeout(() => setKopieren('idle'), 2000)
      } catch {
        setKopieren('fehler')
      }
    }

    return (
      <div
        role="status"
        aria-atomic="false"
        className="relative overflow-hidden rounded-[22px] border border-card-border bg-card p-8 text-center shadow-[var(--shadow)]"
      >
        <div
          className="absolute inset-x-0 top-0 h-[5px]"
          style={{ background: 'var(--spectral)' }}
          aria-hidden="true"
        ></div>
        <div className="flex flex-col items-center">
          <div
            className="demo-pop flex h-[88px] w-[88px] items-center justify-center rounded-full text-brand"
            style={{
              background: 'color-mix(in srgb, var(--primary) 16%, transparent)',
            }}
          >
            <Check size={46} strokeWidth={2.4} aria-hidden="true" />
          </div>
          <h2 className="font-brand mt-6 text-[26px] font-bold tracking-[-0.02em]">
            E-Mail-Entwurf geöffnet
          </h2>
          <p className="mt-3 max-w-[34em] text-[15.5px] leading-relaxed text-muted">
            Wir haben einen vorbefüllten E-Mail-Entwurf in deinem Mailprogramm
            geöffnet. Bitte prüfe ihn und <strong>sende ihn ab</strong> — erst
            mit dem Absenden ist die Nutzungsvereinbarung geschlossen.
          </p>
          <p className="mt-4 max-w-[34em] text-[14px] leading-relaxed text-muted">
            Es gibt keine Freigabe und keine Zugangsdaten. Mit dem Absenden
            könnt ihr{' '}
            <a
              href={installationUrl}
              className="font-semibold text-brand hover:underline"
            >
              installieren
            </a>
            . Antwortet der Autor, kann die E-Mail in eurem Spam-Ordner landen —
            dort lohnt sich ein Blick.
          </p>
          <div className="mt-5 w-full text-left">
            <p className="text-[13px] text-muted">
              Öffnet sich kein Entwurf oder hast du keine Mail-App? Kopiere die
              Angaben unten und verschicke die E-Mail selbst.
            </p>
            <dl className="mt-2.5 flex flex-col gap-1 text-[13px]">
              <div className="flex gap-1.5">
                <dt className="font-semibold">An:</dt>
                <dd className="select-all">{mail.empfaenger}</dd>
              </div>
              <div className="flex gap-1.5">
                <dt className="font-semibold">Betreff:</dt>
                <dd className="select-all">{mail.betreff}</dd>
              </div>
            </dl>
            <label className="mt-2.5 block">
              <span className="mb-1 block text-[12px] font-semibold">Text</span>
              <textarea
                readOnly
                value={mail.text}
                rows={7}
                className="w-full resize-y rounded-[11px] border border-card-border bg-background px-3.5 py-2.5 text-[13px] leading-relaxed text-foreground outline-none focus:border-brand focus:ring-[3px] focus:ring-[color:var(--ring)]"
              />
            </label>
            <button
              type="button"
              onClick={kopiereText}
              className="btn btn-ghost mt-2.5"
            >
              {kopieren === 'kopiert' ? (
                <>
                  <Check size={16} aria-hidden="true" />
                  Kopiert
                </>
              ) : (
                <>
                  <Copy size={16} aria-hidden="true" />
                  Kopieren
                </>
              )}
            </button>
            {kopieren === 'fehler' && (
              <p className="mt-1.5 text-[13px] text-[var(--sp-red-text)]">
                Kopieren war nicht möglich. Bitte den Text im Feld markieren und
                manuell kopieren.
              </p>
            )}
          </div>
          <button
            type="button"
            onClick={() => {
              setKopieren('idle')
              setGesendet(false)
            }}
            className="btn btn-ghost mt-7"
          >
            Zurück zum Formular
          </button>
        </div>
      </div>
    )
  }

  const fehlerId = (feld: string) => `${uid}-${feld}-error`

  return (
    <form
      ref={formRef}
      onSubmit={onSubmit}
      noValidate
      className="relative overflow-hidden rounded-[22px] border border-card-border bg-card p-[30px] shadow-[var(--shadow)]"
    >
      <div
        className="absolute inset-x-0 top-0 h-[5px]"
        style={{ background: 'var(--spectral)' }}
        aria-hidden="true"
      ></div>

      {/* Assertive Live-Region: kündigt fehlgeschlagenes Absenden an. */}
      <div role="alert" className="sr-only">
        {ankuendigung}
      </div>

      <h2 className="font-brand text-[22px] font-bold tracking-[-0.02em]">
        Nutzungsvereinbarung abschließen
      </h2>
      <p className="mt-1.5 mb-[22px] text-[14px] text-muted">
        Eine einzige E-Mail schließt die Vereinbarung ab — danach könnt ihr
        direkt loslegen.
      </p>

      <div className="flex flex-col gap-[15px]">
        <label className="block">
          <span className="mb-1.5 block text-[13px] font-semibold">
            Verein / Organisation
          </span>
          <input
            name="verein"
            type="text"
            value={felder.verein}
            onChange={(event) => setFeld('verein', event.target.value)}
            placeholder="TSV Musterhausen e.V."
            required
            aria-invalid={fehler.verein ? true : undefined}
            aria-describedby={fehler.verein ? fehlerId('verein') : undefined}
            className={`${feldKlassen} h-[46px]`}
          />
          {fehler.verein && (
            <p
              id={fehlerId('verein')}
              className="mt-1.5 text-[13px] text-[var(--sp-red-text)]"
            >
              {fehler.verein}
            </p>
          )}
        </label>

        <label className="block">
          <span className="mb-1.5 block text-[13px] font-semibold">
            Ansprechpartner:in
          </span>
          <input
            name="name"
            type="text"
            value={felder.name}
            onChange={(event) => setFeld('name', event.target.value)}
            placeholder="Vor- und Nachname"
            required
            aria-invalid={fehler.name ? true : undefined}
            aria-describedby={fehler.name ? fehlerId('name') : undefined}
            className={`${feldKlassen} h-[46px]`}
          />
          {fehler.name && (
            <p
              id={fehlerId('name')}
              className="mt-1.5 text-[13px] text-[var(--sp-red-text)]"
            >
              {fehler.name}
            </p>
          )}
        </label>

        <label className="block">
          <span className="mb-1.5 block text-[13px] font-semibold">E-Mail</span>
          <input
            name="email"
            type="email"
            value={felder.email}
            onChange={(event) => setFeld('email', event.target.value)}
            placeholder="vorstand@musterhausen.de"
            required
            aria-invalid={fehler.email ? true : undefined}
            aria-describedby={fehler.email ? fehlerId('email') : undefined}
            className={`${feldKlassen} h-[46px]`}
          />
          {fehler.email && (
            <p
              id={fehlerId('email')}
              className="mt-1.5 text-[13px] text-[var(--sp-red-text)]"
            >
              {fehler.email}
            </p>
          )}
        </label>

        <label className="block">
          <span className="mb-1.5 block text-[13px] font-semibold">
            Rechtsform
          </span>
          <select
            name="art"
            value={felder.art}
            onChange={(event) => setFeld('art', event.target.value)}
            className={`${feldKlassen} h-[46px]`}
          >
            {artOptionen.map((option) => (
              <option key={option} value={option}>
                {option}
              </option>
            ))}
          </select>
        </label>

        <label className="block">
          <span className="mb-1.5 block text-[13px] font-semibold">
            Nachricht <span className="font-normal text-muted">(optional)</span>
          </span>
          <textarea
            name="message"
            rows={3}
            value={felder.message}
            onChange={(event) => setFeld('message', event.target.value)}
            placeholder="Für welche Veranstaltung möchtet ihr jotti nutzen?"
            className={`${feldKlassen} resize-y py-[11px]`}
          />
        </label>
      </div>

      <button type="submit" className="btn btn-primary mt-[22px] w-full">
        Vereinbarung abschließen
        <ArrowRight size={17} aria-hidden="true" />
      </button>
      <p className="mt-3.5 text-center text-[12px] leading-[1.5] text-muted">
        Kostenlos für gemeinnützige Organisationen. Das Formular öffnet nur
        einen vorbefüllten E-Mail-Entwurf — gesendet wird er von dir.
      </p>
    </form>
  )
}
