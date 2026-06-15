import { useEffect, useRef } from 'react'

import { Input } from '@/components/ui/input'
import { cn, formatCents, parseCents } from '@/lib/utils'

/** Keeps only digits and a single comma so the field can never hold invalid characters. */
const cleanInput = (input: string): string => {
  return input.replace(/[^0-9,]/g, '').replace(/,+/g, ',')
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
 * The canonical money input: a fixed `€` prefix, the decimal keypad on mobile,
 * input sanitised to digits and a comma, and normalisation to two decimals on
 * blur (and after a typing pause). String in, string out — wrap it in a form
 * field for cents-based state.
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
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Clear any pending normalisation when the input unmounts.
  useEffect(() => {
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current)
    }
  }, [])

  return (
    <div className="flex">
      <span className="inline-flex items-center px-3 rounded-l-md border border-r-0 border-input bg-background text-muted-foreground text-sm">
        €
      </span>
      <Input
        id={id}
        type="text"
        inputMode="decimal"
        className={cn('border-l-0', className)}
        autoComplete="off"
        spellCheck={false}
        placeholder={placeholder}
        disabled={disabled}
        aria-invalid={ariaInvalid}
        value={value}
        onChange={(e) => {
          const cleaned = cleanInput(e.target.value)
          onValueChange(cleaned)

          // Debounce the reformat so the cursor does not jump while typing.
          if (debounceRef.current) clearTimeout(debounceRef.current)
          debounceRef.current = setTimeout(() => {
            onValueChange(formatBlur(cleaned))
          }, 1000)
        }}
        onBlur={() => {
          if (debounceRef.current) clearTimeout(debounceRef.current)
          onValueChange(formatBlur(value))
          onBlur?.()
        }}
      />
    </div>
  )
}
