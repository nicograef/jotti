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
