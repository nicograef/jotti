import { z } from 'zod'

export const Kategorie = {
  ESSEN: 'essen',
  GETRAENK: 'getraenk',
  SONSTIGES: 'sonstiges',
} as const
export type Kategorie = (typeof Kategorie)[keyof typeof Kategorie]

export const VarianteStatus = {
  ACTIVE: 'active',
  INACTIVE: 'inactive',
} as const
export type VarianteStatus =
  (typeof VarianteStatus)[keyof typeof VarianteStatus]

export const ProduktIdSchema = z.number().int().min(1)
export const VarianteIdSchema = z.number().int().min(1)

const NameSchema = z
  .string()
  .min(3, { message: 'Das sieht nicht nach einem echten Namen aus.' })
  .max(100, { message: 'Der Name ist zu lang.' })
const PreisCentsSchema = z
  .number()
  .int()
  .min(0, { message: 'Preis muss mindestens 0 Cent sein.' })
const KategorieSchema = z.enum(['essen', 'getraenk', 'sonstiges'])
const VarianteStatusSchema = z.enum(['active', 'inactive'])
const DateStringSchema = z.string().refine((date) => !isNaN(Date.parse(date)), {
  message: 'Ungültiges Datumsformat',
})

export const VarianteSchema = z.object({
  id: VarianteIdSchema,
  name: NameSchema,
  preisCents: PreisCentsSchema,
  status: VarianteStatusSchema,
  createdAt: DateStringSchema,
})
export type Variante = z.infer<typeof VarianteSchema>

export const ProduktSchema = z.object({
  id: ProduktIdSchema,
  name: NameSchema,
  kategorie: KategorieSchema,
  varianten: z.array(VarianteSchema),
  createdAt: DateStringSchema,
})
export type Produkt = z.infer<typeof ProduktSchema>
