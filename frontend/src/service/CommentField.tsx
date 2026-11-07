import { useRef, useState } from 'react'

import { Field } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { kuerzeAufZeichen, zeichenAnzahl } from '@/lib/zeichen'

import { KOMMENTAR_MAX_ZEICHEN } from './schemas'

// Kürzt nur den neu eingefügten Teil auf das Zeichen-Budget; der bestehende Text bleibt.
// cursor ist das Ende der Einfügung in nachher, wie setSelectionRange in UTF-16-Einheiten.
function begrenzeEinfuegung(
  vorher: string,
  nachher: string,
  cursor: number,
  maxZeichen: number,
): { text: string; cursor: number } {
  let suffix = nachher.length - cursor
  if (!vorher.endsWith(nachher.slice(cursor))) {
    suffix = 0
    while (
      suffix < vorher.length &&
      suffix < nachher.length &&
      vorher[vorher.length - 1 - suffix] ===
        nachher[nachher.length - 1 - suffix]
    ) {
      suffix++
    }
  }
  let prefix = 0
  while (
    prefix < vorher.length - suffix &&
    prefix < nachher.length - suffix &&
    vorher[prefix] === nachher[prefix]
  ) {
    prefix++
  }
  const davor = nachher.slice(0, prefix)
  const danach = nachher.slice(nachher.length - suffix)
  const budget = maxZeichen - zeichenAnzahl(davor) - zeichenAnzahl(danach)
  const eingefuegt = kuerzeAufZeichen(
    nachher.slice(prefix, nachher.length - suffix),
    Math.max(budget, 0),
  )
  return {
    text: davor + eingefuegt + danach,
    cursor: prefix + eingefuegt.length,
  }
}

// Kappt eine Änderung auf KOMMENTAR_MAX_ZEICHEN und liefert den neuen Text.
// maxLength zählt UTF-16-Einheiten und kappte Emoji schon bei 50; das Backend zählt Codepoints.
function kappeAenderung(
  feld: HTMLInputElement | HTMLTextAreaElement,
  vorher: string,
): string {
  const neu = feld.value
  if (zeichenAnzahl(neu) <= KOMMENTAR_MAX_ZEICHEN) return neu
  const begrenzt = begrenzeEinfuegung(
    vorher,
    neu,
    feld.selectionEnd ?? neu.length,
    KOMMENTAR_MAX_ZEICHEN,
  )
  feld.value = begrenzt.text
  feld.setSelectionRange(begrenzt.cursor, begrenzt.cursor)
  return begrenzt.text
}

interface KommentarFieldProps {
  onChange: (value: string) => void
  // Kontrolliert das Feld, wenn gesetzt: die dauerhaft sichtbare Abschluss-Spalte
  // braucht das, sonst verdeckt ein remountetes leeres Feld einen alten State.
  // Die kurzlebigen Korrektur-Drawer bleiben unkontrolliert.
  value?: string
  required?: boolean
  invalid?: boolean
}

export function KommentarField({
  onChange,
  value,
  required = false,
  invalid = false,
}: KommentarFieldProps) {
  const [touched, setTouched] = useState(false)
  const letzterWert = useRef('')

  return (
    <Field>
      <Textarea
        className="resize-none"
        placeholder={
          required ? 'Kommentar (erforderlich)' : 'Kommentar (optional)'
        }
        rows={3}
        value={value}
        onChange={(e) => {
          setTouched(true)
          const vorher = value ?? letzterWert.current
          const neu = kappeAenderung(e.target, vorher)
          letzterWert.current = neu
          if (neu !== vorher) {
            onChange(neu)
          }
        }}
        onBlur={() => {
          setTouched(true)
        }}
        spellCheck={false}
      />
      {required && (
        <p
          className={`text-sm mt-1 ${
            touched && invalid ? 'text-destructive' : 'text-muted-foreground'
          }`}
        >
          Kommentar ist erforderlich (mind. 3 Zeichen).
        </p>
      )}
    </Field>
  )
}

export function Kommentar({ value }: { value: string }) {
  return (
    <Field>
      <Textarea
        readOnly
        value={value}
        className="resize-none focus-visible:ring-0 focus-visible:border-input"
        rows={3}
        spellCheck={false}
      />
    </Field>
  )
}

// Einzeiliges Feld für den Kommentar einer Position, mit der Kappung des
// KommentarField. Unkontrolliert: der Abschluss speichert den Rohtext, die
// Belegzeile trägt ihn getrimmt.
export function PositionsKommentarFeld({
  defaultValue,
  onChange,
  autoFocus = false,
}: {
  defaultValue: string
  onChange: (value: string) => void
  autoFocus?: boolean
}) {
  const letzterWert = useRef(defaultValue)

  return (
    <Input
      placeholder="Kommentar zur Position"
      aria-label="Kommentar zur Position"
      defaultValue={defaultValue}
      autoFocus={autoFocus}
      onChange={(e) => {
        const neu = kappeAenderung(e.target, letzterWert.current)
        if (neu !== letzterWert.current) {
          letzterWert.current = neu
          onChange(neu)
        }
      }}
      spellCheck={false}
    />
  )
}
