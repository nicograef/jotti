import { useState } from 'react'

import { Field } from '@/components/ui/field'
import { Textarea } from '@/components/ui/textarea'

interface KommentarFieldProps {
  onChange: (value: string) => void
  required?: boolean
  invalid?: boolean
}

export function KommentarField({
  onChange,
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
