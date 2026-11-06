import { z } from 'zod'

import { VerkaufPositionInputSchema } from '../direktverkauf/Direktverkauf'

// tischId 0 ist der Direktverkauf.
export const PfandZurueckgebenSchema = z.object({
  rueckgabeId: z.uuid(),
  tischId: z.number().int().min(0),
  positionen: VerkaufPositionInputSchema.array().min(1),
  kommentar: z.string().max(100),
})
export type PfandZurueckgeben = z.infer<typeof PfandZurueckgebenSchema>
