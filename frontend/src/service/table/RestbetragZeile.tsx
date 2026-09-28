import { formatEuro } from '@/lib/utils'

export function RestbetragZeile({ cents }: { cents: number }) {
  return (
    <div className="flex items-center justify-between gap-3 text-[13px] text-muted-foreground">
      <span>Nach dieser Zahlung noch offen</span>
      <span className="font-semibold tabular-nums text-foreground">
        {formatEuro(cents)}
      </span>
    </div>
  )
}
