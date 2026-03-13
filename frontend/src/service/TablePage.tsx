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
import { AuthSingleton } from '@/lib/Auth'
import { BackendSingleton } from '@/lib/Backend'
import { formatCents } from '@/lib/utils'

import { Bestellung } from './components/table/Bestellung'
import { Lieferung } from './components/table/Lieferung'
import { TischHistorie } from './components/table/TischHistorie'
import { Zahlung } from './components/table/Zahlung'
import { useTischState } from './table/hooks'
import { useTisch } from './table/hooks'
import { TischBackend } from './table/TischBackend'

const tischBackend = new TischBackend(BackendSingleton)

export function TablePage() {
  const { tableId } = useParams<{ tableId: string }>()
  const { loading: tischLoading, tisch } = useTisch(Number(tableId))
  const {
    state,
    loading: stateLoading,
    reload: reloadState,
  } = useTischState(Number(tableId))

  const offenePositionen = state.ungeliefertePositionen.reduce(
    (sum, position) => sum + position.menge,
    0,
  )

  return (
    <>
      <Item>
        <ItemContent>
          <ItemTitle className="text-2xl">
            {tischLoading ? 'Tisch ??' : tisch?.name}{' '}
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
        <div className="w-full fixed bottom-4 left-0 z-50 flex justify-center">
          <TabsList>
            <TabsTrigger value="order" className="p-4">
              Bestellen
            </TabsTrigger>
            <TabsTrigger value="payment" className="p-4">
              Bezahlen
            </TabsTrigger>
            <TabsTrigger value="history" className="p-4">
              Historie
            </TabsTrigger>
          </TabsList>
        </div>
        <TabsContent value="order">
          {tisch && (
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
                    }}
                  />
                </Card>
              )}
              <Bestellung
                backend={tischBackend}
                tisch={tisch}
                onBestellungAufgegeben={() => {
                  reloadState()
                }}
              />
            </>
          )}
        </TabsContent>
        <TabsContent value="payment">
          {tisch && (
            <Zahlung
              backend={tischBackend}
              tisch={tisch}
              positionen={state.unbezahltePositionen}
              loading={stateLoading}
              onZahlungRegistriert={() => {
                reloadState()
              }}
              onProdukteStorniert={() => {
                reloadState()
              }}
            />
          )}
        </TabsContent>
        <TabsContent value="history">
          {tisch && (
            <TischHistorie tischId={tisch.id} userId={AuthSingleton.userId} />
          )}
        </TabsContent>
      </Tabs>
    </>
  )
}
