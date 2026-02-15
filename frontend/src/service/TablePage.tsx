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

import { Delivery } from './components/table/Delivery'
import { Order } from './components/table/Order'
import { Payment } from './components/table/Payment'
import { TableHistory } from './components/table/TableHistory'
import { useTableBalance, useTableUndeliveredVariants } from './table/hooks'
import { useTable } from './table/hooks'
import { TableBackend } from './table/TableBackend'

const tableBackend = new TableBackend(BackendSingleton)

export function TablePage() {
  const { tableId } = useParams<{ tableId: string }>()
  const { loading: tableLoading, table } = useTable(Number(tableId))
  const {
    balanceCents,
    loading: balanceLoading,
    reload: reloadBalance,
  } = useTableBalance(Number(tableId))
  const {
    variants: undeliveredVariants,
    loading: undeliveredVariantsLoading,
    reload: reloadUndeliveredVariants,
  } = useTableUndeliveredVariants(Number(tableId))

  const openVariants = undeliveredVariants.reduce(
    (sum, variant) => sum + variant.quantity,
    0,
  )

  return (
    <>
      <Item>
        <ItemContent>
          <ItemTitle className="text-2xl">
            {tableLoading ? 'Tisch ??' : table?.name}{' '}
            {!undeliveredVariantsLoading && openVariants > 0 && (
              <Badge variant="destructive">{openVariants} offen</Badge>
            )}
            {!undeliveredVariantsLoading && openVariants === 0 && (
              <Badge>Alles geliefert!</Badge>
            )}
          </ItemTitle>
        </ItemContent>
        <ItemContent>
          <ItemDescription className="text-2xl">
            {balanceLoading ? '?' : formatCents(balanceCents)} €
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
          {table && (
            <>
              {openVariants > 0 && (
                <Card className="p-2 gap-0 mb-4">
                  <Delivery
                    backend={tableBackend}
                    table={table}
                    onVariantsDelivered={() => {
                      reloadUndeliveredVariants()
                    }}
                  />
                </Card>
              )}
              <Order
                backend={tableBackend}
                table={table}
                onOrderPlaced={() => {
                  reloadBalance()
                  reloadUndeliveredVariants()
                }}
              />
            </>
          )}
        </TabsContent>
        <TabsContent value="payment">
          {table && (
            <Payment
              backend={tableBackend}
              table={table}
              onPaymentRegistered={() => {
                reloadBalance()
              }}
              onVariantsCanceled={() => {
                reloadBalance()
              }}
            />
          )}
        </TabsContent>
        <TabsContent value="history">
          {table && (
            <TableHistory tableId={table.id} userId={AuthSingleton.userId} />
          )}
        </TabsContent>
      </Tabs>
    </>
  )
}
