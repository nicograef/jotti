import { z } from 'zod'

export const ProductCategory = {
  FOOD: 'food',
  BEVERAGE: 'beverage',
  OTHER: 'other',
} as const
export type ProductCategory =
  (typeof ProductCategory)[keyof typeof ProductCategory]

/** Deutsche Labels für Produktkategorien */
export const ProductCategoryLabels: Record<ProductCategory, string> = {
  food: 'Essen',
  beverage: 'Getränke',
  other: 'Sonstiges',
}

/** Sortierreihenfolge der Kategorien in der UI */
export const ProductCategoryOrder: ProductCategory[] = [
  'food',
  'beverage',
  'other',
]

const ProductIdSchema = z.number().int().min(1)
const VariantIdSchema = z.number().int().min(1)
const NameSchema = z
  .string()
  .min(3, { message: 'Das sieht nicht nach einem echten Namen aus.' })
  .max(50, { message: 'Der Name ist zu lang.' })
const PriceCentsSchema = z
  .number()
  .int()
  .min(0, { message: 'Der Nettopreis muss positiv sein.' })
const CategorySchema = z.enum(['food', 'beverage', 'other'])

export const VariantSchema = z.object({
  id: VariantIdSchema,
  name: NameSchema,
  priceCents: PriceCentsSchema,
})
export type Variant = z.infer<typeof VariantSchema>

export const ProductSchema = z.object({
  id: ProductIdSchema,
  name: NameSchema,
  category: CategorySchema,
  variants: z.array(VariantSchema),
})
export type Product = z.infer<typeof ProductSchema>
