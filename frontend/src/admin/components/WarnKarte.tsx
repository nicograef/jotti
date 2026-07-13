import { TriangleAlert } from 'lucide-react'
import type { ReactNode } from 'react'

import { cn } from '@/lib/utils'

// Warnkarte für problematische Zustände (Design-Handoff-Token: Rahmen
// destructive/40, Fläche destructive/4, abgerundet, mit Warn-Icon). Ab Phase
// 3/9 verwendet. Ein optionaler Titel steht fett über dem Fließtext.
export function WarnKarte({
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
        'flex gap-3 rounded-lg border border-destructive/40 bg-destructive/4 p-4 text-sm',
        className,
      )}
    >
      <TriangleAlert
        className="mt-0.5 size-4 shrink-0 text-destructive"
        aria-hidden
      />
      <div className="min-w-0 space-y-1 leading-relaxed">
        {title !== undefined && (
          <p className="font-medium text-destructive">{title}</p>
        )}
        <div>{children}</div>
      </div>
    </div>
  )
}
