import { z } from 'zod'

import { maxUtf8Bytes } from '@/lib/utf8'

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
  // Transienter Barrierestatus während des Abschlusses (Saldo-Prüfung,
  // Reporting, TSE-Signierung) — nicht dasselbe wie abgeschlossen.
  WIRD_ABGESCHLOSSEN: 'wird_abgeschlossen',
  ABGESCHLOSSEN: 'abgeschlossen',
} as const
export type KassensitzungStatus =
  (typeof KassensitzungStatus)[keyof typeof KassensitzungStatus]

// Das Backend zählt beide Grenzen in UTF-8-Bytes.
export const BezeichnungSchema = maxUtf8Bytes(
  z.string().min(1, { message: 'Bezeichnung ist erforderlich.' }),
  200,
  'Bezeichnung darf maximal 200 Zeichen lang sein.',
)

export const BetragCentsSchema = z
  .number()
  .int()
  .gte(0, { message: 'Betrag muss mindestens 0 Cent sein.' })

export const KommentarSchema = maxUtf8Bytes(
  z
    .string()
    .min(3, { message: 'Kommentar muss mindestens 3 Zeichen lang sein.' }),
  200,
  'Kommentar darf maximal 200 Zeichen lang sein.',
)

// KasseBackend erweitert diesen Datensatz zu AktiveKassensitzung.
export const KassensitzungSchema = z.object({
  zNr: z.number().int(),
  datum: z.string(),
  bezeichnung: z.string(),
  status: z.enum(KassensitzungStatus),
})

// Der Kassenbestand (Soll) samt Aufschlüsselung. Es gilt (vor dem Kassensturz):
// anfangsbestand + bareinnahmen + einlagen − entnahmen = sollBestand.
export const KassenbestandSchema = z.object({
  sollBestandCents: z.number().int(),
  anfangsbestandCents: z.number().int(),
  bareinnahmenCents: z.number().int(),
  einlagenCents: z.number().int(),
  entnahmenCents: z.number().int(),
})
export type Kassenbestand = z.infer<typeof KassenbestandSchema>

// gebuchtVon ist der eingefrorene Anzeigename aus dem Kassenjournal.
export const GeldtransitBuchungSchema = z.object({
  zeitpunkt: z.string(),
  richtung: GeldtransitRichtungSchema,
  betragCents: z.number().int(),
  kommentar: z.string(),
  gebuchtVon: z.string(),
})
export type GeldtransitBuchung = z.infer<typeof GeldtransitBuchungSchema>
