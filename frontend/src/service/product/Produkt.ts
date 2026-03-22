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

export const VarianteSchema = z.object({
  id: VariantIdSchema,
  name: NameSchema,
  preisCents: PreisCentsSchema,
})
export type Variante = z.infer<typeof VarianteSchema>

export const ProduktSchema = z.object({
  id: ProductIdSchema,
  name: NameSchema,
  kategorie: KategorieSchema,
  varianten: z.array(VarianteSchema),
})
export type Produkt = z.infer<typeof ProduktSchema>
