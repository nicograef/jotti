import { useParams } from 'react-router'

import {
  Item,
  ItemContent,
  ItemDescription,
  ItemTitle,
} from '@/components/ui/item'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { AuthSingleton } from '@/lib/Auth'
import { BackendSingleton } from '@/lib/Backend'

import { Order } from './Order'
import { Payment } from './Payment'
import { useTableBalance } from './table/hooks'
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

  return (
    <>
      <Item>
        <ItemContent>
          <ItemTitle className="text-2xl">
            {tableLoading ? 'Tisch ??' : table?.name}
          </ItemTitle>
        </ItemContent>
        <ItemContent>
          <ItemDescription className="text-2xl">
            {balanceLoading ? '??' : (balanceCents / 100).toFixed(2)} €
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
            <Order
              backend={tableBackend}
              table={table}
              onOrderPlaced={() => {
                void reloadBalance()
              }}
            />
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
