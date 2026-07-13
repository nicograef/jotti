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
})
export type UmsatzServicekraft = z.infer<typeof UmsatzServicekraftSchema>

// StornierungServicekraft aggregiert die Stornierungen einer Servicekraft
// (Anzahl und Betrag) — Kontroll-Signal für Live-Dashboard und Kassenberichte.
export const StornierungServicekraftSchema = z.object({
  userId: z.number().int(),
  userName: z.string(),
  name: z.string(),
  anzahlStornierungen: z.number().int(),
  stornierungenCents: z.number().int(),
})
export type StornierungServicekraft = z.infer<
  typeof StornierungServicekraftSchema
>

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

// AbgeschlosseneSitzung ist ein Eintrag der Kassenberichte-Sitzungsliste: die
// abgeschlossene Kassensitzung mit Gesamtumsatz und Abschlusszeitpunkt aus dem
// Tagesabschluss-Event. Status entfällt (alle Einträge sind abgeschlossen).
export const AbgeschlosseneSitzungSchema = z.object({
  zNr: z.number().int(),
  datum: z.string(),
  bezeichnung: z.string(),
  umsatzGesamtCents: z.number().int(),
  abgeschlossenAm: z.string().nullable(),
})
export type AbgeschlosseneSitzung = z.infer<typeof AbgeschlosseneSitzungSchema>

// Metadaten sind die Kopfdaten des formalen Tagesberichts, rein aus den
// Journal-Events projiziert. Alle Felder sind optional, solange die zugehörigen
// Events fehlen (z. B. bei einer noch offenen Sitzung).
export const MetadatenSchema = z.object({
  eroeffnetAm: z.string().nullable(),
  abgeschlossenAm: z.string().nullable(),
  abgeschlossenVon: z.string(),
  kassensturzDifferenzCents: z.number().int().nullable(),
})
export type Metadaten = z.infer<typeof MetadatenSchema>

export const OffenerTischSchema = z.object({
  tischId: z.number().int(),
  tischName: z.string(),
  saldoCents: z.number().int(),
})
export type OffenerTisch = z.infer<typeof OffenerTischSchema>

// OffeneArbeitTisch trägt nur den Tisch-Namen für die Inline-Anzeige der
// offenen Tische einer Servicekraft; der offene Betrag wird auf
// Servicekraft-Ebene (ServicekraftLive.offenCents) vom Backend aggregiert.
export const OffeneArbeitTischSchema = z.object({
  tischId: z.number().int(),
  tischName: z.string(),
})
export type OffeneArbeitTisch = z.infer<typeof OffeneArbeitTischSchema>

// ServicekraftLive führt den kassierten Umsatz mit der offenen eigenen Arbeit
// zusammen; erledigt ist true, wenn keine offene eigene Arbeit mehr besteht.
export const ServicekraftLiveSchema = z.object({
  userId: z.number().int(),
  userName: z.string(),
  name: z.string(),
  zahlungenCents: z.number().int(),
  offenCents: z.number().int(),
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
    stornierungenProServicekraft: z.array(StornierungServicekraftSchema),
  }),
  stornierungen: z.array(StornierungDetailSchema),
})
export type LiveReportingData = z.infer<typeof LiveReportingDataSchema>

export const ReportingDataSchema = z.object({
  kassensitzungNr: z.number().int(),
  metadaten: MetadatenSchema,
  summary: SummarySchema,
  breakdowns: z.object({
    umsatzProServicekraft: z.array(UmsatzServicekraftSchema),
    stornierungenProServicekraft: z.array(StornierungServicekraftSchema),
  }),
  umsatzProSteuersatz: z.array(UmsatzSteuersatzSchema),
  stornierungen: z.array(StornierungDetailSchema),
})
export type ReportingData = z.infer<typeof ReportingDataSchema>
