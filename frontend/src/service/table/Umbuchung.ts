import { z } from 'zod'

import { PositionRefSchema } from './Bestellung'

export const BestellungUmbuchenSchema = z.object({
  quellTischId: z.number().int().min(1),
  zielTischId: z.number().int().min(1),
  positionen: PositionRefSchema.array().min(1),
})
export type BestellungUmbuchen = z.infer<typeof BestellungUmbuchenSchema>
