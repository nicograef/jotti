import { z } from 'zod'

export const Kategorie = {
  ESSEN: 'essen',
  GETRAENK: 'getraenk',
  SONSTIGES: 'sonstiges',
} as const
export type Kategorie = (typeof Kategorie)[keyof typeof Kategorie]

export const Steuersatz = {
  REGEL: 'regel',
  ERMAESSIGT: 'ermaessigt',
  BEFREIT: 'befreit',
  KOMBI: 'kombi',
} as const
export type Steuersatz = (typeof Steuersatz)[keyof typeof Steuersatz]

export const STEUERSATZ_LABEL: Record<Steuersatz, string> = {
  regel: 'Regelsteuersatz (19 %)',
  ermaessigt: 'Ermäßigter Steuersatz (7 %)',
  befreit: 'Steuerbefreit (0 %)',
  kombi: 'Kombi (70/30)',
}

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
  .max(99999, { message: 'Preis darf maximal 999,99 € betragen.' })
const KategorieSchema = z.enum(['essen', 'getraenk', 'sonstiges'])
const SteuersatzSchema = z.enum(['regel', 'ermaessigt', 'befreit', 'kombi'])
const VarianteStatusSchema = z.enum(['active', 'inactive'])
const DateStringSchema = z.string().refine((date) => !isNaN(Date.parse(date)), {
  message: 'Ungültiges Datumsformat',
})

export function defaultSteuersatzByKategorie(kategorie: Kategorie): Steuersatz {
  if (kategorie === Kategorie.ESSEN) {
    return Steuersatz.ERMAESSIGT
  }

  return Steuersatz.REGEL
}

export const VarianteSchema = z.object({
  id: VarianteIdSchema,
  name: NameSchema,
  preisCents: PreisCentsSchema,
  status: VarianteStatusSchema,
  createdAt: DateStringSchema,
  updatedAt: DateStringSchema,
})
export type Variante = z.infer<typeof VarianteSchema>

export const ProduktSchema = z.object({
  id: ProduktIdSchema,
  name: NameSchema,
  kategorie: KategorieSchema,
  steuersatz: SteuersatzSchema,
  status: VarianteStatusSchema,
  varianten: z.array(VarianteSchema),
  createdAt: DateStringSchema,
  updatedAt: DateStringSchema,
})
export type Produkt = z.infer<typeof ProduktSchema>
