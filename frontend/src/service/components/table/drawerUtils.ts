import type { LineItem } from '../../table/Order'

export function selectVariants(
  variants: LineItem[],
  selectedQuantity: Record<number, number>,
): LineItem[] {
  return variants
    .map((variant) => ({
      ...variant,
      quantity: selectedQuantity[variant.id] || 0,
    }))
    .filter((variant) => variant.quantity > 0)
}

export function calculateTotalPrice(variants: LineItem[]): number {
  return variants.reduce(
    (total, variant) => total + variant.priceCents * variant.quantity,
    0,
  )
}
