import { Field } from '@/components/ui/field'
import { Textarea } from '@/components/ui/textarea'

interface CommentFieldProps {
  onChange: (value: string) => void
}

export function CommentField({ onChange }: CommentFieldProps) {
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

export function Comment({ value }: { value: string }) {
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
