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

import { Bestellung } from './components/table/Bestellung'
import { Lieferung } from './components/table/Lieferung'
import { TischHistorie } from './components/table/TischHistorie'
import { Zahlung } from './components/table/Zahlung'
import { useActiveProducts } from './product/hooks'
import { useTischHistorie, useTischState } from './table/hooks'
import { TischBackend } from './table/TischBackend'

const tischBackend = new TischBackend(BackendSingleton)

export function TablePage() {
  const { tableId } = useParams<{ tableId: string }>()
  const isMobile = useIsMobile()
  const {
    state,
    loading: stateLoading,
    reload: reloadState,
  } = useTischState(Number(tableId))
  const { loading: productsLoading, products } = useActiveProducts()
  const {
    loading: historieLoading,
    historie,
    reload: reloadHistorie,
  } = useTischHistorie(Number(tableId))

  const tisch = { id: state.tischId, name: state.tischName }

  const offenePositionen = state.ungeliefertePositionen.reduce(
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
              <Badge>Alles geliefert!</Badge>
            )}
          </ItemTitle>
        </ItemContent>
        <ItemContent>
          <ItemDescription className="text-2xl">
            {stateLoading ? '?' : formatCents(state.saldoCents)} €
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
              Bezahlen
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
                  <Lieferung
                    backend={tischBackend}
                    tisch={tisch}
                    positionen={state.ungeliefertePositionen}
                    loading={stateLoading}
                    onProdukteGeliefert={() => {
                      reloadState()
                      reloadHistorie()
                    }}
                  />
                </Card>
              )}
              <Bestellung
                backend={tischBackend}
                tisch={tisch}
                products={products}
                productsLoading={productsLoading}
                onBestellungAufgegeben={() => {
                  reloadState()
                  reloadHistorie()
                }}
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
              loading={stateLoading}
              onZahlungRegistriert={() => {
                reloadState()
                reloadHistorie()
              }}
              onProdukteStorniert={() => {
                reloadState()
                reloadHistorie()
              }}
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
              onProdukteStorniert={() => {
                reloadState()
                reloadHistorie()
              }}
            />
          )}
        </TabsContent>
      </Tabs>
    </>
  )
}
