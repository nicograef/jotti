import type { ElementType } from 'react'

import { cn } from '@/lib/utils'

interface WortmarkeProps {
  as?: ElementType
  className?: string
}

// Der Verlauf folgt dem Theme, weil --spectral im Dark Mode überschrieben ist.
export function Wortmarke({
  as: Component = 'span',
  className,
}: WortmarkeProps) {
  return (
    <Component
      className={cn(
        'bg-[image:var(--spectral)] bg-clip-text font-heading font-bold text-transparent',
        className,
      )}
    >
      jotti
    </Component>
  )
}
