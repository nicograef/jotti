import { formatEuro } from '@/lib/utils'

// Must sit in a flex container: the name grows and truncates (min-w-0 flex-1 truncate), the price keeps its width (shrink-0).
// The Service order list does not use it, since its variant names wrap to stay distinguishable.
export function VariantNamePreis({
  name,
  preisCents,
}: {
  name: string
  preisCents: number
}) {
  return (
    <>
      <span className="min-w-0 flex-1 truncate font-medium">{name}</span>
      <span className="shrink-0 text-sm font-semibold tabular-nums">
        {formatEuro(preisCents)}
      </span>
    </>
  )
}
