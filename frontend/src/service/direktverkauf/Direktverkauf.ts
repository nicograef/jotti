import { z } from 'zod'

import { KategorieSchema } from '../product/Produkt'
import {
  DateStringSchema,
  PositionRefSchema,
  SteuersatzSchema,
} from '../schemas'

export const VerkaufPositionInputSchema = z.object({
  produktId: z.number().int().min(1),
  varianteId: z.number().int().min(1),
  menge: z.number().int().min(1),
})
export type VerkaufPositionInput = z.infer<typeof VerkaufPositionInputSchema>

export const DirektverkaufTaetigenSchema = z.object({
  verkaufId: z.uuid(),
  positionen: VerkaufPositionInputSchema.array().min(1),
  kommentar: z.string().max(100),
})
export type DirektverkaufTaetigen = z.infer<typeof DirektverkaufTaetigenSchema>

export const DirektverkaufStornierenSchema = z.object({
  // Client-erzeugter Schlüssel des fachlichen Vorgangs. Ein Wiederholversuch
  // nach Verbindungsabbruch trägt denselben Schlüssel und bucht kein zweites Mal.
  vorgangId: z.uuid(),
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
  einzelpreisCents: z.number().int().min(1),
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
