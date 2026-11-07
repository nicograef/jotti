import { useRef, useState } from 'react'

import { Field } from '@/components/ui/field'
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
          // maxLength zählt UTF-16-Einheiten, das Backend Codepoints: Emoji sprengen sonst das Limit.
          const vorher = value ?? letzterWert.current
          let neu = e.target.value
          if (zeichenAnzahl(neu) > KOMMENTAR_MAX_ZEICHEN) {
            const begrenzt = begrenzeEinfuegung(
              vorher,
              neu,
              e.target.selectionEnd,
              KOMMENTAR_MAX_ZEICHEN,
            )
            neu = begrenzt.text
            e.target.value = neu
            e.target.setSelectionRange(begrenzt.cursor, begrenzt.cursor)
          }
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
