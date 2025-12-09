import { z } from 'zod'

export const ProductCategory = {
  FOOD: 'food',
  BEVERAGE: 'beverage',
  OTHER: 'other',
} as const
export type ProductCategory =
  (typeof ProductCategory)[keyof typeof ProductCategory]

const ProductIdSchema = z.number().int().min(1)
const NameSchema = z
  .string()
  .min(3, { message: 'Das sieht nicht nach einem echten Namen aus.' })
  .max(50, { message: 'Der Name ist zu lang.' })
const DescriptionSchema = z
  .string()
  .max(250, { message: 'Die Beschreibung ist zu lang.' })
const NetPriceCentsSchema = z
  .number()
  .int()
  .min(0, { message: 'Der Nettopreis muss positiv sein.' })
const CategorySchema = z.enum(ProductCategory)

export const ProductSchema = z.object({
  id: ProductIdSchema,
  name: NameSchema,
  description: DescriptionSchema,
  netPriceCents: NetPriceCentsSchema,
  category: CategorySchema,
})
export type Product = z.infer<typeof ProductSchema>
