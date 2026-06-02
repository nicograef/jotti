import { z } from 'zod'

export const KassenbewegungArt = {
  GELDTRANSIT: 'geldtransit',
  PRIVATENTNAHME: 'privatentnahme',
  PRIVATEINLAGE: 'privateinlage',
} as const
export type KassenbewegungArt =
  (typeof KassenbewegungArt)[keyof typeof KassenbewegungArt]

export const KassenbewegungArtSchema = z.enum([
  KassenbewegungArt.GELDTRANSIT,
  KassenbewegungArt.PRIVATENTNAHME,
  KassenbewegungArt.PRIVATEINLAGE,
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
