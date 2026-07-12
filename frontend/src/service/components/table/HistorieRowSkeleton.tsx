import { Skeleton } from '@/components/ui/skeleton'

// Geteiltes Lade-Skeleton für eine Historien-Zeile — genutzt von der Tisch- und
// der Direktverkauf-Historie.
export function HistorieRowSkeleton() {
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
