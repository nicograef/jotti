import { ChevronDown, ChevronRight, Minus, Package, Plus } from 'lucide-react'
import { useState } from 'react'

import { EmptyState } from '@/components/common/EmptyState'
import { Button } from '@/components/ui/button'
import {
  Item,
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemTitle,
} from '@/components/ui/item'
import { Skeleton } from '@/components/ui/skeleton'
import { formatCents } from '@/lib/utils'

import {
  KategorieLabels,
  KategorieOrder,
  type Produkt,
  type Variante,
} from '../../product/Product'

interface ProductListComponentProps {
  products: Produkt[]
  variantMengen: Record<number, number>
  onAdd: (variantId: number) => void
  onRemove: (variantId: number) => void
}

export function ProductList(props: ProductListComponentProps) {
  const [expandedProducts, setExpandedProducts] = useState<Set<number>>(
    () => new Set(),
  )

  const toggleExpanded = (productId: number) => {
    setExpandedProducts((prev) => {
      const next = new Set(prev)
      if (next.has(productId)) {
        next.delete(productId)
      } else {
        next.add(productId)
      }
      return next
    })
  }

  const getProductTotal = (varianten: Variante[]) => {
    return varianten.reduce(
      (sum, v) => sum + (props.variantMengen[v.id] || 0),
      0,
    )
  }

  if (props.products.length === 0) {
    return (
      <EmptyState
        icon={Package}
        title="Keine Produkte verfügbar"
        description="Bitte im Admin-Bereich mindestens eine aktive Variante anlegen."
      />
    )
  }

  return (
    <div className="my-4 space-y-6">
      {KategorieOrder.map((category) => {
        const categoryProducts = props.products.filter(
          (p) => p.kategorie === category,
        )
        if (categoryProducts.length === 0) return null

        return (
          <div key={category}>
            <h2 className="text-lg font-semibold mb-2">
              {KategorieLabels[category]}
            </h2>
            <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3">
              {categoryProducts.map((product) => {
                const isExpanded = expandedProducts.has(product.id)
                const productTotal = getProductTotal(product.varianten)

                return (
                  <div key={product.id} className="space-y-1">
                    <Item
                      variant="outline"
                      className="cursor-pointer"
                      onClick={() => {
                        toggleExpanded(product.id)
                      }}
                    >
                      <ItemContent>
                        <ItemTitle className="flex items-center gap-2">
                          {isExpanded ? (
                            <ChevronDown className="h-4 w-4" />
                          ) : (
                            <ChevronRight className="h-4 w-4" />
                          )}
                          {product.name}
                        </ItemTitle>
                        <ItemDescription>
                          {product.varianten.length} Variante
                          {product.varianten.length !== 1 ? 'n' : ''}
                        </ItemDescription>
                      </ItemContent>
                      {productTotal > 0 && (
                        <ItemActions>
                          <span className="text-sm font-medium bg-primary text-primary-foreground rounded-full px-2 py-1">
                            {productTotal}
                          </span>
                        </ItemActions>
                      )}
                    </Item>
                    {isExpanded && (
                      <div className="ml-4 space-y-1">
                        {product.varianten.map((variant) => (
                          <VariantItem
                            key={variant.id}
                            variant={variant}
                            menge={props.variantMengen[variant.id] || 0}
                            onAdd={() => {
                              props.onAdd(variant.id)
                            }}
                            onRemove={() => {
                              props.onRemove(variant.id)
                            }}
                          />
                        ))}
                      </div>
                    )}
                  </div>
                )
              })}
            </ItemGroup>
          </div>
        )
      })}
    </div>
  )
}

function VariantItem({
  variant,
  menge,
  onAdd,
  onRemove,
}: {
  variant: Variante
  menge: number
  onAdd: () => void
  onRemove: () => void
}) {
  return (
    <Item variant="outline" className="bg-muted/30">
      <ItemContent>
        <ItemTitle className="text-sm">{variant.name}</ItemTitle>
        <ItemDescription>
          <span className="font-bold">
            {formatCents(variant.preisCents)}&nbsp;€
          </span>
        </ItemDescription>
      </ItemContent>
      <ItemActions>
        <Button
          size="icon"
          variant="outline"
          className="rounded-full min-h-12 min-w-12"
          aria-label="Variante entfernen"
          onClick={(e) => {
            e.stopPropagation()
            onRemove()
          }}
        >
          <Minus />
        </Button>
        <span className="text-lg mx-2">{menge}</span>
        <Button
          size="icon"
          variant="outline"
          className="rounded-full min-h-12 min-w-12"
          aria-label="Variante hinzufügen"
          onClick={(e) => {
            e.stopPropagation()
            onAdd()
          }}
        >
          <Plus />
        </Button>
      </ItemActions>
    </Item>
  )
}

export function ProductListSkeleton() {
  return (
    <div className="my-4 space-y-6">
      {KategorieOrder.map((category) => (
        <div key={category}>
          <Skeleton className="h-5 w-24 mb-2" />
          <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3">
            {Array.from({ length: 2 }).map((_, index) => (
              <Item
                key={`skeleton-${category}-${index.toString()}`}
                variant="outline"
              >
                <ItemContent>
                  <Skeleton className="h-4 w-24" />
                </ItemContent>
                <ItemActions>
                  <Plus />
                </ItemActions>
              </Item>
            ))}
          </ItemGroup>
        </div>
      ))}
    </div>
  )
}
