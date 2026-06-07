import { Lamp, ShoppingCart, TableIcon } from 'lucide-react'
import { useState } from 'react'
import { Link } from 'react-router'

import { EmptyState } from '@/components/common/EmptyState'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'

import { EigeneUebersichtKarten } from './components/EigeneUebersicht'
import { MeinTischCard } from './components/MeinTischCard'
import { TischAuswahlDrawer } from './components/TischAuswahlDrawer'
import { useEigeneUebersicht, useMeineTischeState } from './table/hooks'

export function TableSelectionPage() {
  const [drawerOpen, setDrawerOpen] = useState(false)
  const { tische, isPending: tischeLoading } = useMeineTischeState()
  const { uebersicht, isPending: uebersichtLoading } = useEigeneUebersicht()

  return (
    <div className="py-2">
      <Button asChild size="lg" className="w-full mb-4">
        <Link to="/service/direktverkauf">
          <ShoppingCart />
          Direktverkauf
        </Link>
      </Button>

      <EigeneUebersichtKarten
        uebersicht={uebersicht}
        loading={uebersichtLoading}
      />

      {tischeLoading ? (
        <TischListSkeleton />
      ) : tische.length === 0 ? (
        <EmptyState
          icon={Lamp}
          title="Keine Tische markiert"
          description="Du hast noch keine Tische markiert. Wähle Tische aus, um sie hier zu sehen."
          action={
            <Button
              variant="outline"
              onClick={() => {
                setDrawerOpen(true)
              }}
            >
              Tische auswählen
            </Button>
          }
        />
      ) : (
        <div className="grid gap-3 lg:grid-cols-2 2xl:grid-cols-3 mb-4">
          {tische.map((state) => (
            <MeinTischCard key={state.tischId} state={state} />
          ))}
        </div>
      )}

      <div className="mt-2 flex justify-center">
        <Button
          variant="outline"
          className="w-full max-w-xs"
          onClick={() => {
            setDrawerOpen(true)
          }}
        >
          <TableIcon />
          Alle Tische
        </Button>
      </div>

      <TischAuswahlDrawer open={drawerOpen} onOpenChange={setDrawerOpen} />
    </div>
  )
}

function TischListSkeleton() {
  return (
    <div className="grid gap-3 lg:grid-cols-2 2xl:grid-cols-3 mb-4">
      {Array.from({ length: 4 }).map((_, index) => (
        <div
          key={`skeleton-${index.toString()}`}
          className="rounded-lg border p-4"
        >
          <div className="flex justify-between mb-3">
            <Skeleton className="h-5 w-24" />
            <Skeleton className="h-5 w-16" />
          </div>
          <Skeleton className="h-4 w-20" />
        </div>
      ))}
    </div>
  )
}
