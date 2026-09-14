import { Info } from 'lucide-react'
import type { ReactNode } from 'react'

import { cn } from '@/lib/utils'

export function HinweisKarte({
  title,
  children,
  className,
}: {
  title?: string
  children: ReactNode
  className?: string
}) {
  return (
    <div
      className={cn(
        'flex gap-3 rounded-lg border bg-sidebar p-4 text-sm',
        className,
      )}
    >
      <Info
        className="mt-0.5 size-4 shrink-0 text-muted-foreground"
        aria-hidden
      />
      <div className="space-y-1 leading-relaxed">
        {title !== undefined && <p className="font-medium">{title}</p>}
        <div className="text-muted-foreground">{children}</div>
      </div>
    </div>
  )
}
