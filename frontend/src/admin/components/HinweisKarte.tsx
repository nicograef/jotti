import { Info } from 'lucide-react'
import type { ReactNode } from 'react'

export function HinweisKarte({ children }: { children: ReactNode }) {
  return (
    <div className="flex gap-3 rounded-lg border bg-sidebar p-4 text-sm">
      <Info
        className="mt-0.5 size-4 shrink-0 text-muted-foreground"
        aria-hidden
      />
      <div className="leading-relaxed text-muted-foreground">{children}</div>
    </div>
  )
}
