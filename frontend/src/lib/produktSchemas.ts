import { z } from 'zod'

import { EntityStatusSchema } from '@/lib/entityStatus'
import { createNameSchema } from '@/lib/nameSchema'

// Response-Vertrag beider Produkt-Lesepfade — `admin/get-all-produkte` und
// `service/get-aktive-produkte` liefern dasselbe DTO, deshalb liegen die
// Schemas hier; die Formularregeln stehen in src/admin/products/.

export const Kategorie = {
  ESSEN: 'essen',
  GETRAENK: 'getraenk',
  SONSTIGES: 'sonstiges',
} as const
export type Kategorie = (typeof Kategorie)[keyof typeof Kategorie]
export const KategorieSchema = z.enum([
  Kategorie.ESSEN,
  Kategorie.GETRAENK,
  Kategorie.SONSTIGES,
])

// Eine Quelle für Produktliste, Preisliste und Verkaufsstatistik. Der
// Stationsname eines Bondruckers weicht bewusst ab (Singular „Getränk").
export const KATEGORIE_LABEL: Record<Kategorie, string> = {
  essen: 'Essen',
  getraenk: 'Getränke',
  sonstiges: 'Sonstiges',
}
export const KATEGORIE_ORDER: Kategorie[] = [
  Kategorie.ESSEN,
  Kategorie.GETRAENK,
  Kategorie.SONSTIGES,
]

export const Steuersatz = {
  REGEL: 'regel',
  ERMAESSIGT: 'ermaessigt',
  BEFREIT: 'befreit',
  KOMBI: 'kombi',
} as const
export type Steuersatz = (typeof Steuersatz)[keyof typeof Steuersatz]
export const SteuersatzSchema = z.enum([
  Steuersatz.REGEL,
  Steuersatz.ERMAESSIGT,
  Steuersatz.BEFREIT,
  Steuersatz.KOMBI,
])

export const ProduktIdSchema = z.number().int().min(1)
export const VarianteIdSchema = z.number().int().min(1)

const NameSchema = createNameSchema(100)

// Gelesene Preise decken den persistierten Bereich ab (DB-CHECK
// `preis_cents >= 0`); die engere Grenze gehört ins Admin-Formular und darf
// einen bestehenden Datensatz nicht unlesbar machen.
const PreisCentsSchema = z.number().int().min(0)

// pfandVarianteId: die verknüpfte Pfand-Variante oder null. Der Service-Lesepfad
// liefert null, wenn das Ziel nicht bestellbar ist.
export const VarianteSchema = z.object({
  id: VarianteIdSchema,
  name: NameSchema,
  preisCents: PreisCentsSchema,
  status: EntityStatusSchema,
  pfandVarianteId: VarianteIdSchema.nullable(),
})
export type Variante = z.infer<typeof VarianteSchema>

export const ProduktSchema = z.object({
  id: ProduktIdSchema,
  name: NameSchema,
  kategorie: KategorieSchema,
  steuersatz: SteuersatzSchema,
  pfand: z.boolean(),
  status: EntityStatusSchema,
  varianten: z.array(VarianteSchema),
})
export type Produkt = z.infer<typeof ProduktSchema>
