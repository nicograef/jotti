import { Input } from '@/components/ui/input'
import { cn, formatCents, parseCents } from '@/lib/utils'

/**
 * Keeps only digits and a single decimal separator so the field can never hold
 * an invalid amount. A typed `.` counts as decimal separator (manche
 * Tastatur-Layouts liefern trotz `inputMode="decimal"` einen Punkt statt des
 * Kommas — `4.5` muss 4,50 € ergeben, nicht 45,00 €); everything after the
 * first separator is treated as decimals, additional separators are dropped,
 * and decimals are capped at two (parseCents rejects over-precise input).
 */
const cleanInput = (input: string): string => {
  const normalized = input.replace(/\./g, ',').replace(/[^0-9,]/g, '')
  const [ganze, ...rest] = normalized.split(',')
  if (rest.length === 0) return ganze
  return ganze + ',' + rest.join('').slice(0, 2)
}

/** Normalises a raw Euro string to the canonical `12,50` form, or empty when there is no amount. */
const formatBlur = (raw: string): string => {
  const cents = parseCents(raw)
  return cents > 0 ? formatCents(cents) : ''
}

interface EuroInputProps {
  /** The raw, user-facing string (e.g. `12,50`). */
  value: string
  /** Receives the sanitised string on every change and the canonical string on blur. */
  onValueChange: (value: string) => void
  onBlur?: () => void
  id?: string
  placeholder?: string
  className?: string
  disabled?: boolean
  'aria-invalid'?: boolean
}

/**
 * The canonical money input: a `€` sign inside the field, the decimal
 * keypad on mobile, input sanitised to digits and a comma, and normalisation to
 * two decimals on blur. String in, string out — wrap it in a form field for
 * cents-based state.
 */
export function EuroInput({
  value,
  onValueChange,
  onBlur,
  id,
  placeholder = '0,00',
  className,
  disabled,
  'aria-invalid': ariaInvalid,
}: EuroInputProps) {
  return (
    <div className="relative">
      <span className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-2.5 text-sm text-muted-foreground">
        €
      </span>
      <Input
        id={id}
        type="text"
        inputMode="decimal"
        className={cn('pl-7', className)}
        autoComplete="off"
        spellCheck={false}
        placeholder={placeholder}
        disabled={disabled}
        aria-invalid={ariaInvalid}
        value={value}
        onChange={(e) => {
          onValueChange(cleanInput(e.target.value))
        }}
        onBlur={() => {
          onValueChange(formatBlur(value))
          onBlur?.()
        }}
      />
    </div>
  )
}
