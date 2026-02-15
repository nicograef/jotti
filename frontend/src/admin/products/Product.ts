import { z } from 'zod'

export const ProductCategory = {
  FOOD: 'food',
  BEVERAGE: 'beverage',
  OTHER: 'other',
} as const
export type ProductCategory =
  (typeof ProductCategory)[keyof typeof ProductCategory]

export const VariantStatus = {
  ACTIVE: 'active',
  INACTIVE: 'inactive',
} as const
export type VariantStatus = (typeof VariantStatus)[keyof typeof VariantStatus]

export const ProductIdSchema = z.number().int().min(1)
export const VariantIdSchema = z.number().int().min(1)

const NameSchema = z
  .string()
  .min(3, { message: 'Das sieht nicht nach einem echten Namen aus.' })
  .max(50, { message: 'Der Name ist zu lang.' })
const PriceCentsSchema = z
  .number()
  .int()
  .min(0, { message: 'Preis muss mindestens 0 Cent sein.' })
const CategorySchema = z.enum(['food', 'beverage', 'other'])
const VariantStatusSchema = z.enum(['active', 'inactive'])
const DateStringSchema = z.string().refine((date) => !isNaN(Date.parse(date)), {
  message: 'Ungültiges Datumsformat',
})

export const VariantSchema = z.object({
  id: VariantIdSchema,
  name: NameSchema,
  priceCents: PriceCentsSchema,
  status: VariantStatusSchema,
  createdAt: DateStringSchema,
})
export type Variant = z.infer<typeof VariantSchema>

export const ProductSchema = z.object({
  id: ProductIdSchema,
  name: NameSchema,
  category: CategorySchema,
  variants: z.array(VariantSchema),
  createdAt: DateStringSchema,
})
export type Product = z.infer<typeof ProductSchema>
