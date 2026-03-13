import { Field } from '@/components/ui/field'
import { Textarea } from '@/components/ui/textarea'

interface KommentarFieldProps {
  onChange: (value: string) => void
}

export function KommentarField({ onChange }: KommentarFieldProps) {
  return (
    <Field>
      <Textarea
        className="resize-none"
        placeholder="Kommentar (optional)"
        rows={3}
        maxLength={100}
        onChange={(e) => {
          onChange(e.target.value)
        }}
        spellCheck={false}
      />
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
