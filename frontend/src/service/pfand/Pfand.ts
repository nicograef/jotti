import { z } from 'zod'

import {
  DirektverkaufStornierungSchema,
  VerkaufPositionInputSchema,
  VerkaufPositionSchema,
} from '../direktverkauf/Direktverkauf'
import { DateStringSchema, PositionRefSchema } from '../schemas'

// tischId 0 ist der Direktverkauf.
export const PfandZurueckgebenSchema = z.object({
  rueckgabeId: z.uuid(),
  tischId: z.number().int().min(0),
  positionen: VerkaufPositionInputSchema.array().min(1),
  kommentar: z.string().max(100),
})
export type PfandZurueckgeben = z.infer<typeof PfandZurueckgebenSchema>

export const PfandrueckgabeHistorieEintragSchema = z.object({
  rueckgabeId: z.uuid(),
  tischId: z.number().int().min(0),
  userName: z.string(),
  getaetigtAm: DateStringSchema,
  positionen: VerkaufPositionSchema.array(),
  gesamtbetragCents: z.number().int().min(0),
  kommentar: z.string(),
  offenePositionen: VerkaufPositionSchema.array(),
  gesamtStorniertCents: z.number().int().min(0),
  stornierungen: DirektverkaufStornierungSchema.array(),
})
export type PfandrueckgabeHistorieEintrag = z.infer<
  typeof PfandrueckgabeHistorieEintragSchema
>

export const PfandrueckgabeStornierenSchema = z.object({
  stornierungId: z.uuid(),
  rueckgabeId: z.uuid(),
  positionen: PositionRefSchema.array().min(1),
  kommentar: z.string().min(3).max(100),
})
export type PfandrueckgabeStornieren = z.infer<
  typeof PfandrueckgabeStornierenSchema
>

export const PfandbelegDruckenSchema = z.object({
  rueckgabeId: z.uuid(),
  stornierungId: z.uuid().optional(),
})
export type PfandbelegDrucken = z.infer<typeof PfandbelegDruckenSchema>
