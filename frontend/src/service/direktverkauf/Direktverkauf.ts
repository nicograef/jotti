import { z } from 'zod'

import { KategorieSchema } from '../product/Produkt'
import { DateStringSchema, SteuersatzSchema } from '../schemas'

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

export const DirektverkaufKassenbelegDruckenSchema = z.object({
  verkaufId: z.uuid(),
  stornierungId: z.uuid().optional(),
})
export type DirektverkaufKassenbelegDrucken = z.infer<
  typeof DirektverkaufKassenbelegDruckenSchema
>

export const VerkaufPositionSchema = z.object({
  positionId: z.uuid(),
  varianteId: z.number().int().min(1),
  produktName: z.string().min(1).max(100),
  varianteName: z.string().min(1).max(100),
  kategorie: KategorieSchema,
  steuersatz: SteuersatzSchema,
  einzelpreis: z.number().int().min(0),
  menge: z.number().int().min(1),
})
export type VerkaufPosition = z.infer<typeof VerkaufPositionSchema>

export const DirektverkaufStornierungSchema = z.object({
  stornierungId: z.uuid(),
  storniertAm: DateStringSchema,
  gesamtStornierungCents: z.number().int().min(0),
})
export type DirektverkaufStornierung = z.infer<
  typeof DirektverkaufStornierungSchema
>

export const DirektverkaufHistorieEintragSchema = z.object({
  verkaufId: z.uuid(),
  userName: z.string(),
  getaetigtAm: DateStringSchema,
  positionen: VerkaufPositionSchema.array(),
  gesamtbetragCents: z.number().int().min(0),
  kommentar: z.string(),
  offenePositionen: VerkaufPositionSchema.array(),
  gesamtStorniertCents: z.number().int().min(0),
  stornierungen: DirektverkaufStornierungSchema.array(),
})
export type DirektverkaufHistorieEintrag = z.infer<
  typeof DirektverkaufHistorieEintragSchema
>
