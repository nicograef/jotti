import { z } from 'zod'

import { KategorieSchema, SteuersatzSchema } from '@/lib/produktSchemas'
import { DateStringSchema } from '@/lib/utils'
import { maxZeichen } from '@/lib/zeichen'

// Die eingefrorenen Event-Schemas begrenzen den Kommentar auf 100 Zeichen.
export const KOMMENTAR_MAX_ZEICHEN = 100

export const KommentarSchema = maxZeichen(z.string(), KOMMENTAR_MAX_ZEICHEN)
export const PflichtKommentarSchema = maxZeichen(
  z.string().min(3),
  KOMMENTAR_MAX_ZEICHEN,
)

export const TischIdSchema = z.number().int().min(1)

export const PositionRefSchema = z.object({
  positionId: z.uuid(),
  menge: z.number().int().min(1),
})
export type PositionRef = z.infer<typeof PositionRefSchema>

// Bestellte, verkaufte oder zurückgegebene Variante (eingabe.PositionInput).
// max(999) spiegelt kasse.PositionEingabeSchema und hält die Preissumme vom
// Überlauf fern. Es gilt nur auf dem Eingabeweg: GebuchtePositionSchema bleibt
// offen, damit bestehende größere Bestellungen weiter angezeigt werden.
export const PositionInputSchema = z.object({
  produktId: z.number().int().min(1),
  varianteId: z.number().int().min(1),
  menge: z.number().int().min(1).max(999),
})
export type PositionInput = z.infer<typeof PositionInputSchema>

// Die im selben Abschluss verrechnete Pfandrückgabe von Zahlung und Verkauf.
export const PfandVerrechnungSchema = z.object({
  rueckgabeId: z.uuid(),
  positionen: PositionInputSchema.array().min(1),
})
export type PfandVerrechnung = z.infer<typeof PfandVerrechnungSchema>

export const GebuchtePositionSchema = z.object({
  positionId: z.uuid(),
  varianteId: z.number().int().min(1),
  produktName: z.string().min(1).max(100),
  varianteName: z.string().min(1).max(100),
  kategorie: KategorieSchema,
  steuersatz: SteuersatzSchema,
  einzelpreisCents: z.number().int().min(1),
  menge: z.number().int().min(1),
})
export type GebuchtePosition = z.infer<typeof GebuchtePositionSchema>

const VorgangStornierungSchema = z.object({
  stornierungId: z.uuid(),
  storniertAm: DateStringSchema,
  gesamtStornierungCents: z.number().int().min(0),
})

// Historie eines Vorgangs mit eigenem Stream: Direktverkauf oder Pfandrückgabe.
export const VorgangHistorieBasisSchema = z.object({
  userName: z.string(),
  getaetigtAm: DateStringSchema,
  positionen: GebuchtePositionSchema.array(),
  gesamtbetragCents: z.number().int().min(0),
  kommentar: z.string(),
  offenePositionen: GebuchtePositionSchema.array(),
  gesamtStorniertCents: z.number().int().min(0),
  stornierungen: VorgangStornierungSchema.array(),
})
export type VorgangHistorieBasis = z.infer<typeof VorgangHistorieBasisSchema>
