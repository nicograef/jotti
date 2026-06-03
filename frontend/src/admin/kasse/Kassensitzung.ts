import { z } from 'zod'

export const GeldtransitRichtung = {
  EINLAGE: 'einlage',
  ENTNAHME: 'entnahme',
} as const
export type GeldtransitRichtung =
  (typeof GeldtransitRichtung)[keyof typeof GeldtransitRichtung]

export const GeldtransitRichtungSchema = z.enum([
  GeldtransitRichtung.EINLAGE,
  GeldtransitRichtung.ENTNAHME,
])

export const KassensitzungStatus = {
  OFFEN: 'offen',
  ABGESCHLOSSEN: 'abgeschlossen',
} as const
export type KassensitzungStatus =
  (typeof KassensitzungStatus)[keyof typeof KassensitzungStatus]

export const BezeichnungSchema = z
  .string()
  .min(1, { message: 'Bezeichnung ist erforderlich.' })
  .max(200, { message: 'Bezeichnung darf maximal 200 Zeichen lang sein.' })

export const BetragCentsSchema = z
  .number()
  .int()
  .gte(0, { message: 'Betrag muss mindestens 0 Cent sein.' })

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
