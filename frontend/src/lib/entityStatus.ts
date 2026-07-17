import { z } from 'zod'

// Status einer Stammdaten-Entität (Produkt, Variante), gespiegelt am
// Backend-`produkt.Status` bzw. DB-Enum `EntityStatus`. Soft-gelöschte Entitäten
// liefert das Backend nie an das Frontend aus, daher nur die beiden im UI
// erreichbaren Werte.
export const EntityStatusSchema = z.enum(['active', 'inactive'])
export type EntityStatus = z.infer<typeof EntityStatusSchema>
