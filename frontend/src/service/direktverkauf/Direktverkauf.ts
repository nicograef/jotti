import { z } from 'zod'

import { KategorieSchema, SteuersatzSchema } from '@/lib/produktSchemas'

import { DateStringSchema, PositionRefSchema } from '../schemas'

// max(999) spiegelt kasse.PositionEingabeSchema im Backend; siehe
// BestellPositionInputSchema: die Grenze gilt nur auf dem Eingabeweg, die
// gelesene VerkaufPositionSchema bleibt offen.
export const VerkaufPositionInputSchema = z.object({
  produktId: z.number().int().min(1),
  varianteId: z.number().int().min(1),
  menge: z.number().int().min(1).max(999),
})
export type VerkaufPositionInput = z.infer<typeof VerkaufPositionInputSchema>

// Die im selben Abschluss verrechnete Pfandrückgabe von Zahlung und Verkauf.
export const PfandVerrechnungSchema = z.object({
  rueckgabeId: z.uuid(),
  positionen: VerkaufPositionInputSchema.array().min(1),
})
export type PfandVerrechnung = z.infer<typeof PfandVerrechnungSchema>

export const DirektverkaufTaetigenSchema = z.object({
  verkaufId: z.uuid(),
  positionen: VerkaufPositionInputSchema.array().min(1),
  kommentar: z.string().max(100),
  pfand: PfandVerrechnungSchema.optional(),
})
export type DirektverkaufTaetigen = z.infer<typeof DirektverkaufTaetigenSchema>

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
  einzelpreisCents: z.number().int().min(1),
  menge: z.number().int().min(1),
})
export type VerkaufPosition = z.infer<typeof VerkaufPositionSchema>

export const DirektverkaufStornierungSchema = z.object({
  stornierungId: z.uuid(),
  storniertAm: DateStringSchema,
  gesamtStornierungCents: z.number().int().min(0),
})

// Historie eines Vorgangs mit eigenem Stream: Direktverkauf oder Pfandrückgabe.
export const VorgangHistorieBasisSchema = z.object({
  userName: z.string(),
  getaetigtAm: DateStringSchema,
  positionen: VerkaufPositionSchema.array(),
  gesamtbetragCents: z.number().int().min(0),
  kommentar: z.string(),
  offenePositionen: VerkaufPositionSchema.array(),
  gesamtStorniertCents: z.number().int().min(0),
  stornierungen: DirektverkaufStornierungSchema.array(),
})
export type VorgangHistorieBasis = z.infer<typeof VorgangHistorieBasisSchema>

export const DirektverkaufHistorieEintragSchema =
  VorgangHistorieBasisSchema.extend({ verkaufId: z.uuid() })
export type DirektverkaufHistorieEintrag = z.infer<
  typeof DirektverkaufHistorieEintragSchema
>
