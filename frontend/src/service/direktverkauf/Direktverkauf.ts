import { z } from 'zod'

export const VerkaufPositionInputSchema = z.object({
  produktId: z.number().int().min(1),
  varianteId: z.number().int().min(1),
  menge: z.number().int().min(1),
})
export type VerkaufPositionInput = z.infer<typeof VerkaufPositionInputSchema>

export const DirektverkaufTaetigenSchema = z.object({
  positionen: VerkaufPositionInputSchema.array().min(1),
  kommentar: z.string().max(100),
})
export type DirektverkaufTaetigen = z.infer<typeof DirektverkaufTaetigenSchema>

export const PositionRefSchema = z.object({
  positionId: z.uuid(),
  menge: z.number().int().min(1),
})
export type PositionRef = z.infer<typeof PositionRefSchema>

export const DirektverkaufStornierenSchema = z.object({
  verkaufId: z.uuid(),
  positionen: PositionRefSchema.array().min(1),
  kommentar: z.string().min(3).max(100),
})
export type DirektverkaufStornieren = z.infer<
  typeof DirektverkaufStornierenSchema
>

export const VerkaufPositionSchema = z.object({
  positionId: z.uuid(),
  varianteId: z.number().int().min(1),
  produktName: z.string().min(1).max(100),
  varianteName: z.string().min(1).max(100),
  kategorie: z.enum(['essen', 'getraenk', 'sonstiges']),
  einzelpreis: z.number().int().min(0),
  menge: z.number().int().min(1),
})
export type VerkaufPosition = z.infer<typeof VerkaufPositionSchema>

export const DirektverkaufHistorieEintragSchema = z.object({
  verkaufId: z.uuid(),
  userName: z.string(),
  getaetigtAm: z.string().refine((date) => !isNaN(Date.parse(date)), {
    message: 'Ungültiges Datumsformat',
  }),
  positionen: VerkaufPositionSchema.array(),
  gesamtbetragCents: z.number().int().min(0),
  kommentar: z.string(),
  offenePositionen: VerkaufPositionSchema.array(),
  gesamtStorniertCents: z.number().int().min(0),
})
export type DirektverkaufHistorieEintrag = z.infer<
  typeof DirektverkaufHistorieEintragSchema
>
