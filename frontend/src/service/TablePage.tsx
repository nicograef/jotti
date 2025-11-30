import { useParams } from 'react-router'

import {
  Item,
  ItemContent,
  ItemDescription,
  ItemTitle,
} from '@/components/ui/item'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { BackendSingleton } from '@/lib/Backend'
import { OrderBackend } from '@/lib/order/OrderBackend'
import { useActiveProducts } from '@/lib/product/hooks'
import { useTable } from '@/lib/table/hooks'
import type { TablePublic } from '@/lib/table/Table'

import { Order } from './Order'

const orderBackend = new OrderBackend(BackendSingleton)

export function TablePage() {
  const { tableId } = useParams<{ tableId: string }>()
  const { loading: productsLoading, products } = useActiveProducts()
  const { loading, table } = useTable(Number(tableId))

  return (
    <>
      {loading || !table ? (
        <TableHeaderSkeleton />
      ) : (
        <TableHeader table={table} />
      )}
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
              backend={orderBackend}
              table={table}
              loading={productsLoading}
              products={products}
            />
          )}
        </TabsContent>
        <TabsContent value="payment">Change your password here.</TabsContent>
        <TabsContent value="history">Change your password here.</TabsContent>
      </Tabs>
    </>
  )
}

function TableHeader({ table }: { table: TablePublic }) {
  return (
    <Item>
      <ItemContent>
        <ItemTitle className="text-2xl">{table.name}</ItemTitle>
      </ItemContent>
      <ItemContent>
        <ItemDescription className="text-2xl">24.00 €</ItemDescription>
      </ItemContent>
    </Item>
  )
}

function TableHeaderSkeleton() {
  return (
    <Item>
      <ItemContent>
        <ItemTitle className="text-2xl">Lade...</ItemTitle>
      </ItemContent>
      <ItemContent>
        <ItemDescription className="text-2xl">Lade...</ItemDescription>
      </ItemContent>
    </Item>
  )
}
