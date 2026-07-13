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
  // Transienter Barrierestatus, den KasseAbschliessen hält (Saldo-Prüfung,
  // Reporting, TSE-Signierung). Ein eigenständiger Wert, nicht dasselbe wie
  // abgeschlossen.
  WIRD_ABGESCHLOSSEN: 'wird_abgeschlossen',
  ABGESCHLOSSEN: 'abgeschlossen',
} as const
export type KassensitzungStatus =
  (typeof KassensitzungStatus)[keyof typeof KassensitzungStatus]

// Anzeige-Label je Status. Der transiente Zwischenstatus wird_abgeschlossen
// bekommt ein eigenes Symbol (🟡) und darf nie wie abgeschlossen (🔴) aussehen.
export function kassensitzungStatusLabel(status: KassensitzungStatus): {
  symbol: string
  text: string
} {
  switch (status) {
    case KassensitzungStatus.OFFEN:
      return { symbol: '🟢', text: 'offen' }
    case KassensitzungStatus.WIRD_ABGESCHLOSSEN:
      return { symbol: '🟡', text: 'wird abgeschlossen…' }
    case KassensitzungStatus.ABGESCHLOSSEN:
      return { symbol: '🔴', text: 'abgeschlossen' }
  }
}

export const BezeichnungSchema = z
  .string()
  .min(1, { message: 'Bezeichnung ist erforderlich.' })
  .max(200, { message: 'Bezeichnung darf maximal 200 Zeichen lang sein.' })

export const BetragCentsSchema = z
  .number()
  .int()
  .gte(0, { message: 'Betrag muss mindestens 0 Cent sein.' })

export const KommentarSchema = z
  .string()
  .min(3, { message: 'Kommentar muss mindestens 3 Zeichen lang sein.' })
  .max(200, { message: 'Kommentar darf maximal 200 Zeichen lang sein.' })

// Canonical Kassensitzung record (zNr, datum, bezeichnung, status). Reporting
// re-exports this so both areas share one definition.
export const KassensitzungSchema = z.object({
  zNr: z.number().int(),
  datum: z.string(),
  bezeichnung: z.string(),
  status: z.enum([
    KassensitzungStatus.OFFEN,
    KassensitzungStatus.WIRD_ABGESCHLOSSEN,
    KassensitzungStatus.ABGESCHLOSSEN,
  ]),
})
export type Kassensitzung = z.infer<typeof KassensitzungSchema>

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

// Eine einzelne gebuchte Bargeldbewegung (Einlage/Entnahme) für die
// Bewegungsliste; gebuchtVon ist der eingefrorene Anzeigename aus dem Kassenjournal.
export const GeldtransitBuchungSchema = z.object({
  zeitpunkt: z.string(),
  richtung: GeldtransitRichtungSchema,
  betragCents: z.number().int(),
  kommentar: z.string(),
  gebuchtVon: z.string(),
})
export type GeldtransitBuchung = z.infer<typeof GeldtransitBuchungSchema>
