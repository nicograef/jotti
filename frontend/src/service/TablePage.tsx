import { useParams } from 'react-router'

import {
  Item,
  ItemContent,
  ItemDescription,
  ItemTitle,
} from '@/components/ui/item'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { BackendSingleton } from '@/lib/Backend'
import { useTableBalance } from '@/lib/order/hooks'
import { OrderBackend } from '@/lib/order/OrderBackend'
import { useTable } from '@/lib/table/hooks'

import { Order } from './Order'
import { Payment } from './Payment'
import { TableHistory } from './TableHistory'

const orderBackend = new OrderBackend(BackendSingleton)

export function TablePage() {
  const { tableId } = useParams<{ tableId: string }>()
  const { loading: tableLoading, table } = useTable(Number(tableId))
  const { balanceCents, loading: balanceLoading } = useTableBalance(
    Number(tableId),
  )

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
          {table && <Order backend={orderBackend} table={table} />}
        </TabsContent>
        <TabsContent value="payment">
          {table && <Payment backend={orderBackend} table={table} />}
        </TabsContent>
        <TabsContent value="history">
          {table && <TableHistory tableId={table.id} />}
        </TabsContent>
      </Tabs>
    </>
  )
}
