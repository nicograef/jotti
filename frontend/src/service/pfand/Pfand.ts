import { z } from 'zod'

import {
  KommentarSchema,
  PflichtKommentarSchema,
  PositionInputSchema,
  PositionRefSchema,
  VorgangHistorieBasisSchema,
} from '../schemas'

// tischId 0 ist der Direktverkauf.
export const PfandTischIdSchema = z.number().int().min(0)

export const PfandZurueckgebenSchema = z.object({
  rueckgabeId: z.uuid(),
  tischId: PfandTischIdSchema,
  positionen: PositionInputSchema.array().min(1),
  kommentar: KommentarSchema,
})
export type PfandZurueckgeben = z.infer<typeof PfandZurueckgebenSchema>

export const PfandrueckgabeHistorieEintragSchema =
  VorgangHistorieBasisSchema.extend({
    rueckgabeId: z.uuid(),
    tischId: PfandTischIdSchema,
  })
export type PfandrueckgabeHistorieEintrag = z.infer<
  typeof PfandrueckgabeHistorieEintragSchema
>

export const PfandrueckgabeStornierenSchema = z.object({
  stornierungId: z.uuid(),
  rueckgabeId: z.uuid(),
  positionen: PositionRefSchema.array().min(1),
  kommentar: PflichtKommentarSchema,
})
export type PfandrueckgabeStornieren = z.infer<
  typeof PfandrueckgabeStornierenSchema
>

export const PfandbelegDruckenSchema = z.object({
  rueckgabeId: z.uuid(),
  stornierungId: z.uuid().optional(),
})
export type PfandbelegDrucken = z.infer<typeof PfandbelegDruckenSchema>
