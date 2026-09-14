import type { ReactNode } from 'react'

import { HeaderGlow, type SpektralFarbe } from './HeaderGlow'

export function AdminPageHeader({
  titel,
  unterzeile,
  aktionen,
  glowFarben,
}: {
  titel: string
  unterzeile?: ReactNode
  aktionen?: ReactNode
  glowFarben?: readonly [SpektralFarbe, SpektralFarbe]
}) {
  return (
    <div className="relative isolate flex flex-wrap items-start justify-between gap-4">
      <HeaderGlow farben={glowFarben} />
      <div>
        <h1 className="font-heading text-2xl font-bold leading-8">{titel}</h1>
        {unterzeile !== undefined && (
          <p className="mt-1 text-sm text-muted-foreground">{unterzeile}</p>
        )}
      </div>
      {aktionen !== undefined && (
        <div className="flex shrink-0 items-center gap-2">{aktionen}</div>
      )}
    </div>
  )
}
