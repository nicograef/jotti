import {
  Item,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemTitle,
} from '@/components/ui/item'
import { useOrders } from '@/lib/order/hooks'

interface TableHistoryProps {
  tableId: number
}

export function TableHistory({ tableId }: TableHistoryProps) {
  const { loading, orders } = useOrders(tableId)

  if (loading) {
    return <div>Loading...</div>
  }

  return (
    <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
      {orders.map((order) => (
        <Item key={order.id} variant="outline">
          <ItemContent>
            <ItemTitle>Bestellung {order.id}</ItemTitle>
            <ItemDescription>
              <span className="font-bold">
                {(order.totalNetPriceCents / 100).toFixed(2)}&nbsp;€
              </span>
              &nbsp; &ndash; &nbsp;
              {order.placedAt}
            </ItemDescription>
          </ItemContent>
        </Item>
      ))}
    </ItemGroup>
  )
}
