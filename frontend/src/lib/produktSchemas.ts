import { z } from 'zod'

import { createNameSchema } from '@/lib/nameSchema'
import { DateStringSchema } from '@/lib/utils'

// Response-Vertrag der Produkt-Lesepfade — `admin/get-all-produkte` und
// `service/get-aktive-produkte` liefern dasselbe DTO (backend
// api/stammdaten/produkt/http/query_handler.go), deshalb liegen die Schemas
// hier und nicht je Bereich. Formular- und Eingaberegeln der Admin-Formulare
// (strengere Preisgrenze, eigene Meldungen) stehen in src/admin/products/.

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

// Status einer Stammdaten-Entität (Produkt, Variante), gespiegelt am
// Backend-`produkt.Status` bzw. DB-Enum `EntityStatus`. Soft-gelöschte Entitäten
// liefert das Backend nie an das Frontend aus (beide Lesepfade filtern
// `status != 'deleted'`), daher nur die beiden im UI erreichbaren Werte.
export const EntityStatusSchema = z.enum(['active', 'inactive'])
export type EntityStatus = z.infer<typeof EntityStatusSchema>

export const ProduktIdSchema = z.number().int().min(1)
export const VarianteIdSchema = z.number().int().min(1)

const NameSchema = createNameSchema(100)

// Gelesene Preise decken den persistierten Bereich ab (DB-CHECK
// `preis_cents >= 0`); die engere Formulargrenze gehört ins Admin-Formular und
// darf einen bestehenden Datensatz nicht unlesbar machen.
const PreisCentsSchema = z.number().int().min(0)

export const VarianteSchema = z.object({
  id: VarianteIdSchema,
  name: NameSchema,
  preisCents: PreisCentsSchema,
  status: EntityStatusSchema,
  createdAt: DateStringSchema,
  updatedAt: DateStringSchema,
})
export type Variante = z.infer<typeof VarianteSchema>

export const ProduktSchema = z.object({
  id: ProduktIdSchema,
  name: NameSchema,
  kategorie: KategorieSchema,
  steuersatz: SteuersatzSchema,
  status: EntityStatusSchema,
  varianten: z.array(VarianteSchema),
  createdAt: DateStringSchema,
  updatedAt: DateStringSchema,
})
export type Produkt = z.infer<typeof ProduktSchema>
