import { z } from 'zod'

import {
  KommentarSchema,
  KommentiertePositionInputSchema,
  PfandVerrechnungSchema,
  PflichtKommentarSchema,
  PositionRefSchema,
  VorgangHistorieBasisSchema,
} from '../schemas'

export const DirektverkaufTaetigenSchema = z.object({
  verkaufId: z.uuid(),
  positionen: KommentiertePositionInputSchema.array().min(1),
  kommentar: KommentarSchema,
  pfand: PfandVerrechnungSchema.optional(),
})
export type DirektverkaufTaetigen = z.infer<typeof DirektverkaufTaetigenSchema>

export const DirektverkaufStornierenSchema = z.object({
  stornierungId: z.uuid(),
  verkaufId: z.uuid(),
  positionen: PositionRefSchema.array().min(1),
  kommentar: PflichtKommentarSchema,
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

export const DirektverkaufHistorieEintragSchema =
  VorgangHistorieBasisSchema.extend({ verkaufId: z.uuid() })
export type DirektverkaufHistorieEintrag = z.infer<
  typeof DirektverkaufHistorieEintragSchema
>
