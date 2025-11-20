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

export const CreateProductRequestSchema = z.object({
  name: NameSchema,
  description: DescriptionSchema,
  netPrice: NetPriceSchema,
  category: CategorySchema,
})
const CreateProductResponseSchema = z.object({
  product: ProductSchema,
})

export const UpdateProductRequestSchema = ProductSchema.pick({
  id: true,
  name: true,
  description: true,
  netPrice: true,
  category: true,
})
const UpdateProductResponseSchema = z.object({
  product: ProductSchema,
})

const GetProductsResponseSchema = z.object({
  products: ProductSchema.array(),
})

interface Backend {
  post<TResponse>(
    endpoint: string,
    body: unknown,
    responseSchema?: z.ZodType<TResponse>,
  ): Promise<TResponse>
}

export class ProductBackend {
  private readonly backend: Backend

  constructor(backend: Backend) {
    this.backend = backend
  }

  public async createProduct(
    newProduct: z.infer<typeof CreateProductRequestSchema>,
  ): Promise<{ product: Product }> {
    const body = CreateProductRequestSchema.parse(newProduct)
    const { product } = await this.backend.post(
      'admin/create-product',
      body,
      CreateProductResponseSchema,
    )
    return { product }
  }

  public async updateProduct(
    updatedProduct: z.infer<typeof UpdateProductRequestSchema>,
  ): Promise<Product> {
    const body = UpdateProductRequestSchema.parse(updatedProduct)
    const { product } = await this.backend.post(
      'admin/update-product',
      body,
      UpdateProductResponseSchema,
    )
    return product
  }

  public async getProducts(): Promise<Product[]> {
    const { products } = await this.backend.post(
      'admin/get-products',
      {},
      GetProductsResponseSchema,
    )
    return products
  }

  public async activateProduct(id: number): Promise<void> {
    const body = z.object({ id: ProductIdSchema }).parse({ id })
    await this.backend.post('admin/activate-product', body)
  }

  public async deactivateProduct(id: number): Promise<void> {
    const body = z.object({ id: ProductIdSchema }).parse({ id })
    await this.backend.post('admin/deactivate-product', body)
  }
}
