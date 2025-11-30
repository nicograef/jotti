import { Button } from '@/components/ui/button'
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from '@/components/ui/drawer'
import type { ProductPublic } from '@/lib/product/Product'
import type { TablePublic } from '@/lib/table/Table'

type OrderProduct = ProductPublic & { amount: number }

interface OrderDrawerProps {
  table: TablePublic
  products: ProductPublic[]
  productsAmounts: Record<number, number>
  onSubmit: () => void
}

export function OrderDrawer({
  table,
  products,
  productsAmounts,
  onSubmit,
}: OrderDrawerProps) {
  const orderedProducts = orderProducts(products, productsAmounts)
  const totalPrice = calculateTotalPrice(orderedProducts)

  return (
    <Drawer>
      <DrawerTrigger asChild>
        <div className="text-center">
          <Button className="cursor-pointer hover:shadow-sm w-full lg:w-1/2">
            Bestellung überprüfen
          </Button>
        </div>
      </DrawerTrigger>
      <DrawerContent>
        <div className="mx-auto w-full max-w-sm">
          <DrawerHeader>
            <DrawerTitle>Bestellung für {table.name}</DrawerTitle>
            <DrawerDescription>
              Überprüfe deine Bestellung vor dem Absenden.
            </DrawerDescription>
          </DrawerHeader>
          <div className="my-4 space-y-2">
            {orderedProducts.map((product) => {
              return (
                <div
                  key={product.id}
                  className="flex justify-between border-b pb-2"
                >
                  <div>
                    {product.amount} x {product.name}
                  </div>
                  <div>
                    €{' '}
                    {((product.netPriceCents / 100) * product.amount).toFixed(
                      2,
                    )}
                  </div>
                </div>
              )
            })}
            <div className="flex justify-between font-bold pt-2">
              <div>Gesamt</div>
              <div>€ {(totalPrice / 100).toFixed(2)}</div>
            </div>
          </div>
          <DrawerFooter>
            <Button
              onClick={() => {
                onSubmit()
              }}
            >
              Jetzt Bestellen
            </Button>
            <DrawerClose asChild>
              <Button variant="outline">Abbrechen</Button>
            </DrawerClose>
          </DrawerFooter>
        </div>
      </DrawerContent>
    </Drawer>
  )
}

function orderProducts(
  products: ProductPublic[],
  productsAmounts: Record<number, number>,
): OrderProduct[] {
  return products
    .map((product) => ({
      ...product,
      amount: productsAmounts[product.id] || 0,
    }))
    .filter((product) => product.amount > 0)
}

function calculateTotalPrice(orderProducts: OrderProduct[]): number {
  return orderProducts.reduce(
    (total, product) => total + product.netPriceCents * product.amount,
    0,
  )
}
