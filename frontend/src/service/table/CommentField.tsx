import { useState } from 'react'

import { Field } from '@/components/ui/field'
import { Textarea } from '@/components/ui/textarea'
import { kuerzeAufUtf8Bytes } from '@/lib/utf8'

import { KOMMENTAR_MAX_BYTES } from '../schemas'

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
          // maxLength zählt UTF-16-Einheiten, das Backend Bytes: Umlaute sprengen sonst das Limit.
          const gekuerzt = kuerzeAufUtf8Bytes(
            e.target.value,
            KOMMENTAR_MAX_BYTES,
          )
          if (gekuerzt !== e.target.value) {
            e.target.value = gekuerzt
          }
          onChange(gekuerzt)
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
