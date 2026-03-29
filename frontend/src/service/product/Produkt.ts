import { z } from 'zod'

export const Kategorie = {
  ESSEN: 'essen',
  GETRAENK: 'getraenk',
  SONSTIGES: 'sonstiges',
} as const
export type Kategorie = (typeof Kategorie)[keyof typeof Kategorie]

/** Deutsche Labels für Produktkategorien */
export const KategorieLabels: Record<Kategorie, string> = {
  essen: 'Essen',
  getraenk: 'Getränke',
  sonstiges: 'Sonstiges',
}

/** Sortierreihenfolge der Kategorien in der UI */
export const KategorieOrder: Kategorie[] = ['essen', 'getraenk', 'sonstiges']

const ProductIdSchema = z.number().int().min(1)
const VariantIdSchema = z.number().int().min(1)
const NameSchema = z
  .string()
  .min(3, { message: 'Das sieht nicht nach einem echten Namen aus.' })
  .max(50, { message: 'Der Name ist zu lang.' })
const PreisCentsSchema = z
  .number()
  .int()
  .min(0, { message: 'Der Nettopreis muss positiv sein.' })
const KategorieSchema = z.enum(['essen', 'getraenk', 'sonstiges'])

const VarianteStatusSchema = z.enum(['active', 'inactive'])
const DateStringSchema = z.string().refine((date) => !isNaN(Date.parse(date)), {
  message: 'Ungültiges Datumsformat',
})

export const VarianteSchema = z.object({
  id: VariantIdSchema,
  name: NameSchema,
  preisCents: PreisCentsSchema,
  status: VarianteStatusSchema,
  createdAt: DateStringSchema,
  updatedAt: DateStringSchema,
})
export type Variante = z.infer<typeof VarianteSchema>

const ProduktStatusSchema = z.enum(['active', 'inactive'])

export const ProduktSchema = z.object({
  id: ProductIdSchema,
  name: NameSchema,
  kategorie: KategorieSchema,
  status: ProduktStatusSchema,
  varianten: z.array(VarianteSchema),
  createdAt: DateStringSchema,
  updatedAt: DateStringSchema,
})
export type Produkt = z.infer<typeof ProduktSchema>
