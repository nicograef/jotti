import { z } from 'zod'

export const SummarySchema = z.object({
  gesamtUmsatzCents: z.number().int(),
  gesamtAuszahlungenCents: z.number().int(),
  gesamtBestellungenCents: z.number().int(),
  gesamtStornierungenCents: z.number().int(),
  offeneSaldiCents: z.number().int(),
  ausstehendAuszahlungenCents: z.number().int(),
  anzahlOffeneTische: z.number().int(),
  anzahlBestellungen: z.number().int(),
  anzahlStornierungen: z.number().int(),
})

export const UmsatzServicekraftSchema = z.object({
  userId: z.number().int(),
  userName: z.string(),
  zahlungenCents: z.number().int(),
  auszahlungenCents: z.number().int(),
  anzahlZahlungen: z.number().int(),
})
export type UmsatzServicekraft = z.infer<typeof UmsatzServicekraftSchema>

export const StornierungPositionSchema = z.object({
  produktName: z.string(),
  varianteName: z.string(),
  menge: z.number().int(),
  einzelpreis: z.number().int(),
})

export const StornierungDetailSchema = z.object({
  zeitpunkt: z.string(),
  tischId: z.number().int(),
  tischName: z.string(),
  userId: z.number().int(),
  userName: z.string(),
  betragCents: z.number().int(),
  kommentar: z.string(),
  positionen: z.array(StornierungPositionSchema),
})
export type StornierungDetail = z.infer<typeof StornierungDetailSchema>

export const UmsatzTischSchema = z.object({
  tischId: z.number().int(),
  tischName: z.string(),
  zahlungenCents: z.number().int(),
  auszahlungenCents: z.number().int(),
  anzahlZahlungen: z.number().int(),
})
export type UmsatzTisch = z.infer<typeof UmsatzTischSchema>

export const KassensitzungSchema = z.object({
  zNr: z.number().int(),
  datum: z.string(),
  bezeichnung: z.string(),
  status: z.string(),
})
export type Kassensitzung = z.infer<typeof KassensitzungSchema>

export const ReportingDataSchema = z.object({
  kassensitzungNr: z.number().int(),
  summary: SummarySchema,
  breakdowns: z.object({
    umsatzProServicekraft: z.array(UmsatzServicekraftSchema),
    umsatzProTisch: z.array(UmsatzTischSchema),
  }),
  stornierungen: z.array(StornierungDetailSchema),
})
export type ReportingData = z.infer<typeof ReportingDataSchema>
