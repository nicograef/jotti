import { z } from 'zod'

export const KassensitzungStateSchema = z.object({
  zNr: z.number().int(),
  datum: z.string(),
  bezeichnung: z.string(),
  status: z.enum(['offen', 'abgeschlossen']),
})
export type KassensitzungState = z.infer<typeof KassensitzungStateSchema>

export const KassenbestandSchema = z.object({
  sollBestandCents: z.number().int(),
})
export type Kassenbestand = z.infer<typeof KassenbestandSchema>
