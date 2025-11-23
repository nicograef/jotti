import { z } from 'zod'

export const ProductCategory = {
  FOOD: 'food',
  BEVERAGE: 'beverage',
  OTHER: 'other',
} as const
export type ProductCategory =
  (typeof ProductCategory)[keyof typeof ProductCategory]

export const ProductStatus = {
  ACTIVE: 'active',
  INACTIVE: 'inactive',
  DELETED: 'deleted',
} as const
export type ProductStatus = (typeof ProductStatus)[keyof typeof ProductStatus]

const ProductIdSchema = z.number().int().min(1)
const NameSchema = z
  .string()
  .min(3, { message: 'Das sieht nicht nach einem echten Namen aus.' })
  .max(50, { message: 'Der Name ist zu lang.' })
const DescriptionSchema = z
  .string()
  .max(250, { message: 'Die Beschreibung ist zu lang.' })
const NetPriceSchema = z
  .number()
  .min(0, { message: 'Der Nettopreis muss positiv sein.' })
const CategorySchema = z.enum(ProductCategory)
const ProductStatusSchema = z.enum(ProductStatus)
const DateStringSchema = z.string().refine((date) => !isNaN(Date.parse(date)), {
  message: 'Ungültiges Datumsformat',
})

export const ProductSchema = z.object({
  id: ProductIdSchema,
  name: NameSchema,
  description: DescriptionSchema,
  netPrice: NetPriceSchema,
  category: CategorySchema,
  createdAt: DateStringSchema,
  status: ProductStatusSchema,
})
export type Product = z.infer<typeof ProductSchema>
