import { z } from 'zod'

// === Dashboard ===
export const DashboardSchema = z.object({
  gesamtUmsatzCents: z.number(),
  anzahlOffeneTische: z.number(),
  anzahlBestellungen: z.number(),
  anzahlStornierungen: z.number(),
  gesamtBestellungenCents: z.number(),
  gesamtStornierungenCents: z.number(),
})
export type Dashboard = z.infer<typeof DashboardSchema>

// === Tagesabrechnung ===
export const UmsatzServicekraftSchema = z.object({
  userId: z.number(),
  userName: z.string(),
  zahlungenCents: z.number(),
  anzahlZahlungen: z.number(),
})
export type UmsatzServicekraft = z.infer<typeof UmsatzServicekraftSchema>

export const StornierungPositionSchema = z.object({
  produktName: z.string(),
  varianteName: z.string(),
  menge: z.number(),
  einzelpreis: z.number(),
})

export const StornierungDetailSchema = z.object({
  zeitpunkt: z.string(),
  tischId: z.number(),
  tischName: z.string(),
  userId: z.number(),
  userName: z.string(),
  betragCents: z.number(),
  kommentar: z.string(),
  positionen: z.array(StornierungPositionSchema),
})
export type StornierungDetail = z.infer<typeof StornierungDetailSchema>

export const TagesabrechnungSchema = z.object({
  zeitraum: z.object({ von: z.string(), bis: z.string() }),
  gesamtUmsatzCents: z.number(),
  gesamtBestellungenCents: z.number(),
  gesamtStornierungenCents: z.number(),
  offeneSaldiCents: z.number(),
  anzahlBestellungen: z.number(),
  anzahlStornierungen: z.number(),
  umsatzProServicekraft: z.array(UmsatzServicekraftSchema),
  stornierungen: z.array(StornierungDetailSchema),
})
export type Tagesabrechnung = z.infer<typeof TagesabrechnungSchema>
