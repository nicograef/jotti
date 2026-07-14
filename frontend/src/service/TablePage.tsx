import { useState } from 'react'
import { useParams } from 'react-router'

import { LadefehlerAlert } from '@/components/common/LadefehlerAlert'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useErstAufbau } from '@/hooks/use-erst-aufbau'
import { BackendSingleton } from '@/lib/Backend'
import { formatCents } from '@/lib/utils'

import { ServiceDock } from './components/ServiceDock'
import { Bestellung } from './components/table/Bestellung'
import { TischHistorie } from './components/table/TischHistorie'
import { Zahlung } from './components/table/Zahlung'
import { useAktiveProdukte } from './product/hooks'
import { useTischHistorie, useTischState } from './table/hooks'
import { TischBackend } from './table/TischBackend'

const tischBackend = new TischBackend(BackendSingleton)

// Das Status-Badge poppt bei jedem Wertwechsel (Motion-Inventar „Statuswechsel",
// 350 ms), aber nicht beim ersten Aufbau. Der key-Wechsel auf den Zählwert
// remountet StatusBadgeInhalt, das seine Pop-Entscheidung beim Mount erfasst:
// Beim ersten Aufbau ist `animieren` noch false; nach dem ersten Aufbau mountet
// jeder neue Wert mit true und poppt.
function TischStatusBadge({ anzahlUnbezahlt }: { anzahlUnbezahlt: number }) {
  const erstAufbau = useErstAufbau(true)
  return (
    <StatusBadgeInhalt
      key={anzahlUnbezahlt}
      anzahlUnbezahlt={anzahlUnbezahlt}
      animieren={!erstAufbau}
    />
  )
}

function StatusBadgeInhalt({
  anzahlUnbezahlt,
  animieren,
}: {
  anzahlUnbezahlt: number
  animieren: boolean
}) {
  const [poppen] = useState(animieren)
  const popKlasse = poppen ? 'animate-pop' : undefined

  return anzahlUnbezahlt > 0 ? (
    <Badge variant="destructive" className={popKlasse}>
      {anzahlUnbezahlt} unbezahlt
    </Badge>
  ) : (
    <Badge className={popKlasse}>Alles bezahlt</Badge>
  )
}

// Unterer Freiraum des Tab-Inhalts in Dock-Höhe, damit die letzte Listenzeile
// über dem fixierten ServiceDock endet und antippbar bleibt. Das Dock ist der
// Aktionsbutton (3.5rem) plus TabsList (2.5rem) plus Innenabstände; der Freiraum
// wächst mit env(safe-area-inset-bottom) mit.
const dockFreiraum = 'pb-[calc(9rem+env(safe-area-inset-bottom,0px))]'

export function TablePage() {
  const { tischId } = useParams<{ tischId: string }>()
  const {
    state,
    isPending: stateLoading,
    isError: stateError,
    refetch: reloadState,
  } = useTischState(Number(tischId))
  const { isPending, produkte } = useAktiveProdukte()
  const {
    isPending: historieLoading,
    isError: historieError,
    historie,
    refetch: reloadHistorie,
  } = useTischHistorie(Number(tischId))

  const reload = () => {
    void reloadState()
    void reloadHistorie()
  }

  // Expliziter Fehlerzustand statt der Leer-Defaults (Saldo 0,00 €) — sonst
  // wirkt der Tisch bei Netzabbruch abgerechnet.
  if (stateError || historieError) {
    return (
      <LadefehlerAlert
        titel="Tischdaten konnten nicht geladen werden"
        onErneutVersuchen={reload}
      />
    )
  }

  const tisch = {
    id: state.tischId,
    name: state.tischName,
    saldoCents: state.saldoCents,
  }

  const tabsLocked = stateLoading || historieLoading
  const anzahlUnbezahlt = state.unbezahltePositionen.length

  return (
    <>
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="font-heading text-[22px] font-semibold leading-tight">
            {stateLoading ? 'Tisch ??' : tisch.name}
          </h1>
          {!stateLoading && (
            <div className="mt-1.5 flex flex-wrap items-center gap-2">
              <TischStatusBadge anzahlUnbezahlt={anzahlUnbezahlt} />
              <span className="text-sm text-muted-foreground">
                {state.fuerMichErledigt
                  ? 'Für dich erledigt'
                  : 'Für dich noch offen'}
              </span>
            </div>
          )}
        </div>
        <div className="text-right">
          <div className="text-[11px] font-medium uppercase tracking-[0.04em] text-muted-foreground">
            Offen
          </div>
          <div
            data-slot="tisch-saldo"
            className="text-xl font-bold tabular-nums"
          >
            {stateLoading ? '?' : <>{formatCents(state.saldoCents)}&nbsp;€</>}
          </div>
        </div>
      </div>
      <Tabs defaultValue="order" className="mt-4">
        <ServiceDock
          leiste={
            <>
              {tabsLocked && (
                <p className="rounded-md border bg-background/90 px-3 py-1 text-center text-xs text-muted-foreground">
                  Lade Tischdaten. Tabs sind kurzzeitig deaktiviert.
                </p>
              )}
              <TabsList className="h-10 w-full">
                <TabsTrigger
                  value="order"
                  className="flex-1"
                  disabled={tabsLocked}
                >
                  Bestellen
                </TabsTrigger>
                <TabsTrigger
                  value="payment"
                  className="flex-1"
                  disabled={tabsLocked}
                >
                  Kassieren
                </TabsTrigger>
                <TabsTrigger
                  value="history"
                  className="flex-1"
                  disabled={tabsLocked}
                >
                  Historie
                </TabsTrigger>
              </TabsList>
            </>
          }
        >
          <TabsContent value="order" className={dockFreiraum}>
            {!stateLoading && (
              <Bestellung
                backend={tischBackend}
                tisch={tisch}
                products={produkte}
                productsLoading={isPending}
                onBestellungAufgenommen={reload}
              />
            )}
          </TabsContent>
          <TabsContent value="payment" className={dockFreiraum}>
            {!stateLoading && (
              <Zahlung
                backend={tischBackend}
                tisch={tisch}
                positionen={state.unbezahltePositionen}
                onZahlungKassiert={reload}
              />
            )}
          </TabsContent>
          <TabsContent value="history" className={dockFreiraum}>
            {!stateLoading && (
              <TischHistorie
                historie={historie}
                historieLoading={historieLoading}
                tisch={tisch}
                backend={tischBackend}
                onStornierungErteilt={reload}
                onBestellungUmgebucht={reload}
              />
            )}
          </TabsContent>
        </ServiceDock>
      </Tabs>
    </>
  )
}
