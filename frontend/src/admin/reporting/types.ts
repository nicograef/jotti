import { z } from 'zod'

export const SummarySchema = z.object({
  gesamtUmsatzCents: z.number().int(),
  gesamtBestellungenCents: z.number().int(),
  gesamtStornierungenCents: z.number().int(),
  geldtransitCents: z.number().int(),
  anzahlBestellungen: z.number().int(),
  anzahlStornierungen: z.number().int(),
  anzahlDirektverkaeufe: z.number().int(),
  direktverkaufUmsatzCents: z.number().int(),
})

export const UmsatzServicekraftSchema = z.object({
  userId: z.number().int(),
  userName: z.string(),
  name: z.string(),
  zahlungenCents: z.number().int(),
  anzahlZahlungen: z.number().int(),
})
export type UmsatzServicekraft = z.infer<typeof UmsatzServicekraftSchema>

export const StornierungPositionSchema = z.object({
  produktName: z.string(),
  varianteName: z.string(),
  menge: z.number().int(),
  einzelpreisCents: z.number().int(),
})

export const StornierungDetailSchema = z.object({
  zeitpunkt: z.string(),
  quelle: z.enum(['tisch', 'direktverkauf']),
  barRueckgabe: z.boolean(),
  tischId: z.number().int(),
  tischName: z.string(),
  userId: z.number().int(),
  userName: z.string(),
  name: z.string(),
  betragCents: z.number().int(),
  kommentar: z.string(),
  positionen: z.array(StornierungPositionSchema),
})
export type StornierungDetail = z.infer<typeof StornierungDetailSchema>

export const UmsatzSteuersatzSchema = z.object({
  satz: z.enum(['regel', 'ermaessigt', 'befreit', 'kombi']),
  bruttoCents: z.number().int(),
  nettoCents: z.number().int(),
  steuerCents: z.number().int(),
})
export type UmsatzSteuersatz = z.infer<typeof UmsatzSteuersatzSchema>

export {
  type Kassensitzung,
  KassensitzungSchema,
  kassensitzungStatusLabel,
} from '@/admin/kasse/Kassensitzung'

export const OffenerTischSchema = z.object({
  tischId: z.number().int(),
  tischName: z.string(),
  saldoCents: z.number().int(),
})
export type OffenerTisch = z.infer<typeof OffenerTischSchema>

export const OffeneArbeitTischSchema = z.object({
  tischId: z.number().int(),
  tischName: z.string(),
  anzahlUnbezahlt: z.number().int(),
  anzahlOffen: z.number().int(),
})
export type OffeneArbeitTisch = z.infer<typeof OffeneArbeitTischSchema>

// ServicekraftLive führt den kassierten Umsatz mit der offenen eigenen Arbeit
// zusammen; erledigt ist true, wenn keine offene eigene Arbeit mehr besteht.
export const ServicekraftLiveSchema = z.object({
  userId: z.number().int(),
  userName: z.string(),
  name: z.string(),
  zahlungenCents: z.number().int(),
  anzahlZahlungen: z.number().int(),
  offeneTische: z.array(OffeneArbeitTischSchema),
  erledigt: z.boolean(),
})
export type ServicekraftLive = z.infer<typeof ServicekraftLiveSchema>

export const LiveReportingDataSchema = z.object({
  kassensitzungNr: z.number().int(),
  bezeichnung: z.string(),
  datum: z.string(),
  offeneTische: z.array(OffenerTischSchema),
  offeneSaldiCents: z.number().int(),
  summary: SummarySchema,
  breakdowns: z.object({
    servicekraefte: z.array(ServicekraftLiveSchema),
  }),
  stornierungen: z.array(StornierungDetailSchema),
})
export type LiveReportingData = z.infer<typeof LiveReportingDataSchema>

export const ReportingDataSchema = z.object({
  kassensitzungNr: z.number().int(),
  summary: SummarySchema,
  breakdowns: z.object({
    umsatzProServicekraft: z.array(UmsatzServicekraftSchema),
  }),
  umsatzProSteuersatz: z.array(UmsatzSteuersatzSchema),
  stornierungen: z.array(StornierungDetailSchema),
})
export type ReportingData = z.infer<typeof ReportingDataSchema>
