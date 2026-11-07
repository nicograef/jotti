import { z } from 'zod'

import { maxUtf8Bytes } from '@/lib/utf8'

export { DateStringSchema } from '@/lib/utils'

// Die eingefrorenen Event-Schemas begrenzen den Kommentar auf 100 UTF-8-Bytes.
export const KOMMENTAR_MAX_BYTES = 100

export const KommentarSchema = maxUtf8Bytes(z.string(), KOMMENTAR_MAX_BYTES)
export const PflichtKommentarSchema = maxUtf8Bytes(
  z.string().min(3),
  KOMMENTAR_MAX_BYTES,
)

export const PositionRefSchema = z.object({
  positionId: z.uuid(),
  menge: z.number().int().min(1),
})
export type PositionRef = z.infer<typeof PositionRefSchema>
