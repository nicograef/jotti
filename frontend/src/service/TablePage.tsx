import { useParams } from 'react-router'

import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import {
  Item,
  ItemContent,
  ItemDescription,
  ItemTitle,
} from '@/components/ui/item'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useIsMobile } from '@/hooks/use-mobile'
import { AuthSingleton } from '@/lib/Auth'
import { BackendSingleton } from '@/lib/Backend'
import { formatCents } from '@/lib/utils'

import { Ausgabe } from './components/table/Ausgabe'
import { Bestellung } from './components/table/Bestellung'
import { TischHistorie } from './components/table/TischHistorie'
import { Zahlung } from './components/table/Zahlung'
import { useAktiveProdukte } from './product/hooks'
import { useTischHistorie, useTischState } from './table/hooks'
import { TischBackend } from './table/TischBackend'

const tischBackend = new TischBackend(BackendSingleton)

export function TablePage() {
  const { tischId } = useParams<{ tischId: string }>()
  const isMobile = useIsMobile()
  const {
    state,
    isPending: stateLoading,
    refetch: reloadState,
  } = useTischState(Number(tischId))
  const { isPending, produkte } = useAktiveProdukte()
  const {
    isPending: historieLoading,
    historie,
    refetch: reloadHistorie,
  } = useTischHistorie(Number(tischId))

  const reload = () => {
    void reloadState()
    void reloadHistorie()
  }

  const tisch = {
    id: state.tischId,
    name: state.tischName,
    saldoCents: state.saldoCents,
  }

  const offenePositionen = state.ausstehendePositionen.reduce(
    (sum, position) => sum + position.menge,
    0,
  )
  const tabsLocked = stateLoading || historieLoading

  return (
    <>
      <Item>
        <ItemContent>
          <ItemTitle className="text-2xl">
            {stateLoading ? 'Tisch ??' : tisch.name}{' '}
            {!stateLoading && offenePositionen > 0 && (
              <Badge variant="destructive">{offenePositionen} offen</Badge>
            )}
            {!stateLoading && offenePositionen === 0 && (
              <Badge>Alles ausgegeben!</Badge>
            )}
          </ItemTitle>
        </ItemContent>
        <ItemContent>
          <ItemDescription className="text-2xl">
            {stateLoading ? (
              '?'
            ) : (
              <span className={state.saldoCents < 0 ? 'text-destructive' : ''}>
                {formatCents(state.saldoCents)} €
              </span>
            )}
            {!stateLoading && state.saldoCents < 0 && (
              <Badge variant="destructive" className="ml-2">
                Auszahlung ausstehend
              </Badge>
            )}
          </ItemDescription>
        </ItemContent>
      </Item>
      <Tabs defaultValue="order">
        <div
          className={
            isMobile
              ? 'w-full fixed bottom-[calc(1rem+env(safe-area-inset-bottom,0px))] left-0 z-50 flex flex-col items-center gap-2'
              : 'mb-4 flex flex-col items-start gap-2'
          }
        >
          {tabsLocked && (
            <p className="text-xs text-muted-foreground bg-background/90 px-3 py-1 rounded-md border">
              Lade Tischdaten. Tabs sind kurzzeitig deaktiviert.
            </p>
          )}
          <TabsList>
            <TabsTrigger value="order" className="p-4" disabled={tabsLocked}>
              Bestellen
            </TabsTrigger>
            <TabsTrigger value="payment" className="p-4" disabled={tabsLocked}>
              Kassieren
            </TabsTrigger>
            <TabsTrigger value="history" className="p-4" disabled={tabsLocked}>
              Historie
            </TabsTrigger>
          </TabsList>
        </div>
        <TabsContent value="order" className={isMobile ? 'pb-24' : ''}>
          {!stateLoading && (
            <>
              {offenePositionen > 0 && (
                <Card className="p-2 gap-0 mb-4">
                  <Ausgabe
                    backend={tischBackend}
                    tisch={tisch}
                    positionen={state.ausstehendePositionen}
                    loading={stateLoading}
                    onAusgabeBestaetigt={reload}
                  />
                </Card>
              )}
              <Bestellung
                backend={tischBackend}
                tisch={tisch}
                products={produkte}
                productsLoading={isPending}
                onBestellungAufgenommen={reload}
              />
            </>
          )}
        </TabsContent>
        <TabsContent value="payment" className={isMobile ? 'pb-24' : ''}>
          {!stateLoading && (
            <Zahlung
              backend={tischBackend}
              tisch={tisch}
              positionen={state.unbezahltePositionen}
              saldoCents={state.saldoCents}
              loading={stateLoading}
              onZahlungKassiert={reload}
              onAuszahlungGeleistet={reload}
            />
          )}
        </TabsContent>
        <TabsContent value="history" className={isMobile ? 'pb-24' : ''}>
          {!stateLoading && (
            <TischHistorie
              historie={historie}
              historieLoading={historieLoading}
              userId={AuthSingleton.userId}
              tisch={tisch}
              backend={tischBackend}
              onStornierungErteilt={reload}
              onBestellungUmgebucht={reload}
            />
          )}
        </TabsContent>
      </Tabs>
    </>
  )
}
