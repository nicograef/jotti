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
import { useTischSaldo, useTischUngeliefert } from './table/hooks'
import { useTisch } from './table/hooks'
import { TischBackend } from './table/TischBackend'

const tischBackend = new TischBackend(BackendSingleton)

export function TablePage() {
  const { tableId } = useParams<{ tableId: string }>()
  const { loading: tischLoading, tisch } = useTisch(Number(tableId))
  const {
    saldoCents,
    loading: saldoLoading,
    reload: reloadSaldo,
  } = useTischSaldo(Number(tableId))
  const {
    positionen: ungeliefertePositionen,
    loading: ungeliefertePositionenLoading,
    reload: reloadUngeliefertePositionen,
  } = useTischUngeliefert(Number(tableId))

  const offenePositionen = ungeliefertePositionen.reduce(
    (sum, position) => sum + position.menge,
    0,
  )

  return (
    <>
      <Item>
        <ItemContent>
          <ItemTitle className="text-2xl">
            {tischLoading ? 'Tisch ??' : tisch?.name}{' '}
            {!ungeliefertePositionenLoading && offenePositionen > 0 && (
              <Badge variant="destructive">{offenePositionen} offen</Badge>
            )}
            {!ungeliefertePositionenLoading && offenePositionen === 0 && (
              <Badge>Alles geliefert!</Badge>
            )}
          </ItemTitle>
        </ItemContent>
        <ItemContent>
          <ItemDescription className="text-2xl">
            {saldoLoading ? '?' : formatCents(saldoCents)} €
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
                    onProdukteGeliefert={() => {
                      reloadUngeliefertePositionen()
                    }}
                  />
                </Card>
              )}
              <Bestellung
                backend={tischBackend}
                tisch={tisch}
                onBestellungAufgegeben={() => {
                  reloadSaldo()
                  reloadUngeliefertePositionen()
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
              onZahlungRegistriert={() => {
                reloadSaldo()
              }}
              onProdukteStorniert={() => {
                reloadSaldo()
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
