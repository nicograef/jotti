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

import { Delivery } from './Delivery'
import { Order } from './Order'
import { Payment } from './Payment'
import { useTableBalance, useTableUndeliveredProducts } from './table/hooks'
import { useTable } from './table/hooks'
import { TableBackend } from './table/TableBackend'
import { TableHistory } from './TableHistory'

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
    products: undeliveredProducts,
    loading: undeliveredProductsLoading,
    reload: reloadUndeliveredProducts,
  } = useTableUndeliveredProducts(Number(tableId))

  const openProducts = undeliveredProducts.reduce(
    (sum, product) => sum + product.quantity,
    0,
  )

  return (
    <>
      <Item>
        <ItemContent>
          <ItemTitle className="text-2xl">
            {tableLoading ? 'Tisch ??' : table?.name}{' '}
            {!undeliveredProductsLoading && openProducts > 0 && (
              <Badge variant="destructive">{openProducts} offen</Badge>
            )}
            {!undeliveredProductsLoading && openProducts === 0 && (
              <Badge>Alles geliefert!</Badge>
            )}
          </ItemTitle>
        </ItemContent>
        <ItemContent>
          <ItemDescription className="text-2xl">
            {balanceLoading ? '?' : (balanceCents / 100).toFixed(2)} €
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
              {openProducts > 0 && (
                <Card className="p-2 gap-0 mb-4">
                  <Delivery
                    backend={tableBackend}
                    table={table}
                    onProductsDelivered={() => {
                      void reloadUndeliveredProducts()
                    }}
                  />
                </Card>
              )}
              <Order
                backend={tableBackend}
                table={table}
                onOrderPlaced={() => {
                  void reloadBalance()
                  void reloadUndeliveredProducts()
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
                void reloadBalance()
              }}
              onProductsCanceled={() => {
                void reloadBalance()
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
