import { ChevronRight, Lamp, Search, TableIcon } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router'

import { EmptyState } from '@/components/common/EmptyState'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { useErstAufbau } from '@/hooks/use-erst-aufbau'
import { formatEuro } from '@/lib/utils'

import { EigeneUebersichtKarten } from './components/EigeneUebersicht'
import { MeinTischCard } from './components/MeinTischCard'
import { TischAuswahlDrawer } from './components/TischAuswahlDrawer'
import {
  useAktiveTischeMitFavoriten,
  useEigeneUebersicht,
  useMeineTischeState,
} from './table/hooks'
import type { AktiverTischMitFavorit, TischSession } from './table/Tisch'

const fussleisteFreiraum = 'pb-[calc(6rem+env(safe-area-inset-bottom,0px))]'

export function TableSelectionPage() {
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [suche, setSuche] = useState('')
  const { tische, isPending: tischeLoading } = useMeineTischeState()
  // Die Suche greift über alle aktiven Tische, nicht nur die favorisierten
  // „Meine Tische" — so findet der Nutzer auch einen nicht markierten Tisch und
  // öffnet ihn per Treffer direkt.
  const { tische: alleTische } = useAktiveTischeMitFavoriten()
  const { uebersicht, isPending: uebersichtLoading } = useEigeneUebersicht()

  const sucheGetrimmt = suche.trim()
  const sucheAktiv = sucheGetrimmt.length > 0

  const offeneTische = tische.filter(
    (state) => state.unbezahltePositionen.length > 0,
  )
  const erledigteTische = tische.filter(
    (state) => state.unbezahltePositionen.length === 0,
  )

  // Treffer der Hauptsuche über alle aktiven Tische, nach Name sortiert.
  const suchTreffer = useMemo(() => {
    if (!sucheAktiv) return []
    const q = sucheGetrimmt.toLowerCase()
    return alleTische
      .filter((t) => t.name.toLowerCase().includes(q))
      .sort((a, b) => a.name.localeCompare(b.name, 'de'))
  }, [alleTische, sucheAktiv, sucheGetrimmt])

  // Der Listen-Eintritt staffelt nur beim ersten Aufbau der Favoritenliste mit
  // Daten (nicht beim Skeleton, nicht bei späteren Refetches, nicht in der
  // Suchtrefferliste). Beide Gruppen teilen sich eine fortlaufende Staffelung:
  // „Erledigt" setzt hinter „Noch offen" fort.
  const erstAufbau = useErstAufbau(
    !tischeLoading && !sucheAktiv && tische.length > 0,
  )

  return (
    <div className={fussleisteFreiraum}>
      <EigeneUebersichtKarten
        uebersicht={uebersicht}
        loading={uebersichtLoading}
      />

      {alleTische.length > 0 && (
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

      {sucheAktiv ? (
        suchTreffer.length === 0 ? (
          <div className="py-8 text-center text-muted-foreground">
            <p>Kein aktiver Tisch passt zu „{sucheGetrimmt}“.</p>
          </div>
        ) : (
          <SuchTrefferListe treffer={suchTreffer} />
        )
      ) : tischeLoading ? (
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
      ) : (
        <div className="space-y-6">
          {offeneTische.length > 0 && (
            <TischGruppe
              titel="Noch offen"
              tische={offeneTische}
              eintrittAb={erstAufbau ? 0 : null}
            />
          )}
          {erledigteTische.length > 0 && (
            <TischGruppe
              titel="Erledigt"
              tische={erledigteTische}
              eintrittAb={erstAufbau ? offeneTische.length : null}
            />
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

// SuchTrefferListe zeigt die Treffer der Hauptsuche über alle aktiven Tische.
// Ein Tippen öffnet den Tisch direkt — auch einen nicht favorisierten.
function SuchTrefferListe({ treffer }: { treffer: AktiverTischMitFavorit[] }) {
  const navigate = useNavigate()
  return (
    <div className="flex flex-col gap-2">
      {treffer.map((tisch) => (
        <button
          key={tisch.id}
          type="button"
          onClick={() => {
            void navigate(`/service/tische/${tisch.id.toString()}`)
          }}
          className="flex w-full items-center gap-3 rounded-xl bg-card p-4 text-left ring-1 ring-foreground/10 transition-colors hover:bg-accent/50"
        >
          <div className="min-w-0 flex-1">
            <div className="text-base font-semibold">{tisch.name}</div>
            {tisch.istFavorit && (
              <div className="text-[13px] text-muted-foreground">
                Meine Tische
              </div>
            )}
          </div>
          <span
            className={
              tisch.saldoCents < 0
                ? 'text-sm font-medium text-destructive'
                : 'text-sm text-muted-foreground'
            }
          >
            {formatEuro(tisch.saldoCents)}
          </span>
          <ChevronRight className="size-5 shrink-0 text-muted-foreground" />
        </button>
      ))}
    </div>
  )
}

function TischGruppe({
  titel,
  tische,
  eintrittAb,
}: {
  titel: string
  tische: TischSession[]
  // Start-Index der Eintritts-Staffelung oder `null`, wenn nicht animiert
  // eingetreten werden soll.
  eintrittAb: number | null
}) {
  return (
    <section className="space-y-3">
      <h2 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
        {titel} · {tische.length}
      </h2>
      <div className="grid grid-cols-1 gap-3 lg:grid-cols-2 2xl:grid-cols-3">
        {tische.map((state, index) => (
          <MeinTischCard
            key={state.tischId}
            state={state}
            eintrittIndex={eintrittAb === null ? undefined : eintrittAb + index}
          />
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
