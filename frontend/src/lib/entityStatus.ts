import { z } from 'zod'

// Status of CRUD master data (Produkt, Variante, Tisch, Benutzer). The backend never returns soft-deleted
// entities (every read path filters `status != 'deleted'`), so only the two values the UI can reach exist here.
export const EntityStatus = {
  ACTIVE: 'active',
  INACTIVE: 'inactive',
} as const
export type EntityStatus = (typeof EntityStatus)[keyof typeof EntityStatus]
export const EntityStatusSchema = z.enum(EntityStatus)
