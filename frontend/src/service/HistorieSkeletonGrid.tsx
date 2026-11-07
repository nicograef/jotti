import { Skeleton } from '@/components/ui/skeleton'

function HistorieRowSkeleton() {
  return (
    <div className="flex items-center gap-3 rounded-md border px-3 py-3">
      <Skeleton className="size-10 shrink-0 rounded-full" />
      <div className="flex flex-1 flex-col gap-1">
        <Skeleton className="h-4 w-32" />
        <Skeleton className="h-3 w-48" />
      </div>
      <Skeleton className="h-4 w-16" />
    </div>
  )
}

// Platzhalter einer ladenden Historie im selben Raster wie ihre Zeilen.
export function HistorieSkeletonGrid() {
  return (
    <div className="my-4 grid grid-cols-1 gap-2 lg:grid-cols-2 2xl:grid-cols-3">
      {Array.from({ length: 6 }).map((_, index) => (
        // oxlint-disable-next-line react/no-array-index-key
        <HistorieRowSkeleton key={index} />
      ))}
    </div>
  )
}
