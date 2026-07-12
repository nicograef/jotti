import { z } from 'zod'

import { DateStringSchema } from '@/lib/utils'

export const TischIdSchema = z.number().int().min(1)
const TischNameSchema = z
  .string()
  .min(3, { message: 'Das sieht nicht nach einem echten Namen aus.' })
  .max(100, { message: 'Der Name ist zu lang.' })

export const TischStatus = {
  ACTIVE: 'active',
  INACTIVE: 'inactive',
} as const
export type TischStatus = (typeof TischStatus)[keyof typeof TischStatus]
const TischStatusSchema = z.enum(TischStatus)

export const TischSchema = z.object({
  id: TischIdSchema,
  name: TischNameSchema,
  status: TischStatusSchema,
  // Offener Saldo des Tischs in der aktuell offenen Kassensitzung (Cent; 0 ohne
  // offenen Saldo oder ohne offene Sitzung). Reine Backend-Projektion.
  saldoCents: z.number().int().min(0),
  createdAt: DateStringSchema,
  updatedAt: DateStringSchema,
})
export type Tisch = z.infer<typeof TischSchema>

// Der Namenspräfix eines Tischs ist alles vor der abschließenden Zahl, ohne
// Trennzeichen ('Zelt 3' → 'Zelt', 'Biergarten 12' → 'Biergarten'). Ohne
// abschließende Zahl gibt es keinen Präfix — solche Tische landen in 'Weitere'.
// Bewusst hier, damit die reine Gruppierungsfunktion sie isoliert testen kann.
export const WEITERE_GRUPPE = 'Weitere'
