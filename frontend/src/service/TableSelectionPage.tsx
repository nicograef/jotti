import { Lamp, Search, TableIcon } from 'lucide-react'
import { useMemo, useState } from 'react'

import { EmptyState } from '@/components/common/EmptyState'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'

import { EigeneUebersichtKarten } from './components/EigeneUebersicht'
import { MeinTischCard } from './components/MeinTischCard'
import { TischAuswahlDrawer } from './components/TischAuswahlDrawer'
import { useEigeneUebersicht, useMeineTischeState } from './table/hooks'
import type { TischSession } from './table/Tisch'

const fussleisteFreiraum = 'pb-[calc(6rem+env(safe-area-inset-bottom,0px))]'

export function TableSelectionPage() {
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [suche, setSuche] = useState('')
  const { tische, isPending: tischeLoading } = useMeineTischeState()
  const { uebersicht, isPending: uebersichtLoading } = useEigeneUebersicht()

  const gefilterteTische = useMemo(
    () =>
      tische.filter((state) =>
        state.tischName.toLowerCase().includes(suche.trim().toLowerCase()),
      ),
    [tische, suche],
  )
  const offeneTische = gefilterteTische.filter(
    (state) => state.unbezahltePositionen.length > 0,
  )
  const erledigteTische = gefilterteTische.filter(
    (state) => state.unbezahltePositionen.length === 0,
  )

  return (
    <div className={fussleisteFreiraum}>
      <EigeneUebersichtKarten
        uebersicht={uebersicht}
        loading={uebersichtLoading}
      />

      {tische.length > 0 && (
        <div className="relative mb-4">
          <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            className="h-11 pl-9"
            placeholder="Tisch suchen — Name oder Nummer"
            value={suche}
            onChange={(e) => {
              setSuche(e.target.value)
            }}
          />
        </div>
      )}

      {tischeLoading ? (
        <TischListSkeleton />
      ) : tische.length === 0 ? (
        <EmptyState
          icon={Lamp}
          title="Keine Tische markiert"
          description="Du hast noch keine Tische markiert. Wähle Tische aus, um sie hier zu sehen. Für den Direktverkauf an der Theke wechselst du den Arbeitsmodus im Benutzermenü oben rechts."
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
      ) : gefilterteTische.length === 0 ? (
        <div className="py-8 text-center text-muted-foreground">
          <p>Kein markierter Tisch passt zu „{suche.trim()}“.</p>
          <Button
            variant="link"
            onClick={() => {
              setDrawerOpen(true)
            }}
          >
            In allen Tischen suchen
          </Button>
        </div>
      ) : (
        <div className="space-y-6">
          {offeneTische.length > 0 && (
            <TischGruppe titel="Noch offen" tische={offeneTische} />
          )}
          {erledigteTische.length > 0 && (
            <TischGruppe titel="Erledigt" tische={erledigteTische} />
          )}
        </div>
      )}

      <div className="fixed inset-x-0 bottom-0 z-40 border-t bg-background px-4 pt-3 pb-[calc(1rem+env(safe-area-inset-bottom,0px))]">
        <div className="mx-auto w-full max-w-md">
          <Button
            variant="outline"
            className="h-12 w-full"
            onClick={() => {
              setDrawerOpen(true)
            }}
          >
            <TableIcon />
            Alle Tische
          </Button>
        </div>
      </div>

      <TischAuswahlDrawer open={drawerOpen} onOpenChange={setDrawerOpen} />
    </div>
  )
}

function TischGruppe({
  titel,
  tische,
}: {
  titel: string
  tische: TischSession[]
}) {
  return (
    <section className="space-y-3">
      <h2 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
        {titel} · {tische.length}
      </h2>
      <div className="grid grid-cols-1 gap-3 lg:grid-cols-2 2xl:grid-cols-3">
        {tische.map((state) => (
          <MeinTischCard key={state.tischId} state={state} />
        ))}
      </div>
    </section>
  )
}

function TischListSkeleton() {
  return (
    <div className="grid grid-cols-1 gap-3 lg:grid-cols-2 2xl:grid-cols-3">
      {Array.from({ length: 4 }).map((_, index) => (
        <div
          key={`skeleton-${index.toString()}`}
          className="rounded-xl p-4 ring-1 ring-foreground/10"
        >
          <div className="mb-3 flex justify-between">
            <Skeleton className="h-5 w-24" />
            <Skeleton className="h-5 w-16" />
          </div>
          <Skeleton className="h-4 w-20" />
        </div>
      ))}
    </div>
  )
}
