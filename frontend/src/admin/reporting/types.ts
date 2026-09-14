import { z } from 'zod'

import { SteuersatzSchema } from '@/lib/produktSchemas'

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

// abzugebenCents = kassiertCents − ruecknahmenCents. anzahlStornierungen zählt
// beide Tisch-Storno-Arten als Kontroll-Zähler. Direktverkäufe sind nicht
// enthalten.
export const AbrechnungServicekraftSchema = z.object({
  userId: z.number().int(),
  userName: z.string(),
  name: z.string(),
  kassiertCents: z.number().int(),
  anzahlZahlungen: z.number().int(),
  ruecknahmenCents: z.number().int(),
  anzahlStornierungen: z.number().int(),
  abzugebenCents: z.number().int(),
})
export type AbrechnungServicekraft = z.infer<
  typeof AbrechnungServicekraftSchema
>

// userName ist der eingefrorene Username, name der live aufgelöste Klarname.
export const ServicekraftRefSchema = z.object({
  userId: z.number().int(),
  userName: z.string(),
  name: z.string(),
})
export type ServicekraftRef = z.infer<typeof ServicekraftRefSchema>

export const StornierungPositionSchema = z.object({
  produktName: z.string(),
  varianteName: z.string(),
  menge: z.number().int(),
  einzelpreisCents: z.number().int(),
})

// akteur hat den Storno ausgelöst; betroffene sind die Servicekräfte, deren
// Vorgang er rückgängig macht (Storno-Zuordnung). betroffene liefert das
// Backend nie leer.
export const StornierungDetailSchema = z.object({
  zeitpunkt: z.string(),
  quelle: z.enum(['tisch', 'direktverkauf']),
  barRueckgabe: z.boolean(),
  tischId: z.number().int(),
  tischName: z.string(),
  akteur: ServicekraftRefSchema,
  betroffene: z.array(ServicekraftRefSchema),
  betragCents: z.number().int(),
  kommentar: z.string(),
  positionen: z.array(StornierungPositionSchema),
})
export type StornierungDetail = z.infer<typeof StornierungDetailSchema>

export const UmsatzSteuersatzSchema = z.object({
  satz: SteuersatzSchema,
  bruttoCents: z.number().int(),
  nettoCents: z.number().int(),
  steuerCents: z.number().int(),
})
export type UmsatzSteuersatz = z.infer<typeof UmsatzSteuersatzSchema>

// ausgegebeneMenge (Produktion) und umsatzCents (Einnahmen) ruhen bewusst auf
// getrennten Grundlagen.
export const VarianteStatistikSchema = z.object({
  varianteId: z.number().int(),
  varianteName: z.string(),
  ausgegebeneMenge: z.number().int(),
  umsatzCents: z.number().int(),
})
export type VarianteStatistik = z.infer<typeof VarianteStatistikSchema>

// Vom Backend fertig gruppiert und sortiert; Ein-Varianten-Produkte tragen
// genau eine Variante.
export const ProduktStatistikSchema = z.object({
  kategorie: z.string(),
  produktName: z.string(),
  ausgegebeneMenge: z.number().int(),
  umsatzCents: z.number().int(),
  varianten: z.array(VarianteStatistikSchema),
})
export type ProduktStatistik = z.infer<typeof ProduktStatistikSchema>

// Nur abgeschlossene Sitzungen; abgeschlossenAm stammt aus dem
// Tagesabschluss-Event.
export const AbgeschlosseneSitzungSchema = z.object({
  zNr: z.number().int(),
  datum: z.string(),
  bezeichnung: z.string(),
  umsatzGesamtCents: z.number().int(),
  abgeschlossenAm: z.string().nullable(),
})
export type AbgeschlosseneSitzung = z.infer<typeof AbgeschlosseneSitzungSchema>

// Aus den Journal-Events projiziert; die Felder bleiben leer, solange das
// zugehörige Event fehlt (etwa bei einer offenen Sitzung).
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

// Ohne Betrag: der offene Saldo wird auf Servicekraft-Ebene aggregiert
// (ServicekraftLive.offenCents).
export const OffeneArbeitTischSchema = z.object({
  tischId: z.number().int(),
  tischName: z.string(),
})
export type OffeneArbeitTisch = z.infer<typeof OffeneArbeitTischSchema>

// erledigt ist true, wenn keine offene eigene Arbeit mehr besteht.
export const ServicekraftLiveSchema = z.object({
  userId: z.number().int(),
  userName: z.string(),
  name: z.string(),
  kassiertCents: z.number().int(),
  ruecknahmenCents: z.number().int(),
  anzahlStornierungen: z.number().int(),
  abzugebenCents: z.number().int(),
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
  }),
  stornierungen: z.array(StornierungDetailSchema),
  produktStatistik: z.array(ProduktStatistikSchema),
})
export type LiveReportingData = z.infer<typeof LiveReportingDataSchema>

export const ReportingDataSchema = z.object({
  kassensitzungNr: z.number().int(),
  metadaten: MetadatenSchema,
  summary: SummarySchema,
  breakdowns: z.object({
    abrechnungProServicekraft: z.array(AbrechnungServicekraftSchema),
  }),
  umsatzProSteuersatz: z.array(UmsatzSteuersatzSchema),
  stornierungen: z.array(StornierungDetailSchema),
  produktStatistik: z.array(ProduktStatistikSchema),
})
export type ReportingData = z.infer<typeof ReportingDataSchema>
