import { useState } from 'react'

import { Field } from '@/components/ui/field'
import { Textarea } from '@/components/ui/textarea'

interface KommentarFieldProps {
  onChange: (value: string) => void
  // Kontrolliert das Feld, wenn gesetzt. Die dauerhaft sichtbare
  // Abschluss-Spalte braucht das, damit der angezeigte Text nie vom
  // gesendeten kommentar-State abweicht (sonst würde ein remountetes,
  // leeres Feld einen alten State verdecken). Die kurzlebigen
  // Korrektur-Drawer bleiben unkontrolliert (sie remounten bei jedem Öffnen).
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
        maxLength={100}
        value={value}
        onChange={(e) => {
          setTouched(true)
          onChange(e.target.value)
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
        maxLength={100}
        spellCheck={false}
      />
    </Field>
  )
}
