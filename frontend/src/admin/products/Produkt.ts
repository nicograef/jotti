import { z } from 'zod'

import { createNameSchema } from '@/lib/nameSchema'
import { Kategorie, Steuersatz } from '@/lib/produktSchemas'

export const STEUERSATZ_LABEL: Record<Steuersatz, string> = {
  regel: 'Regelsteuersatz (19 %)',
  ermaessigt: 'Ermäßigter Steuersatz (7 %)',
  befreit: 'Steuerbefreit (0 %)',
  kombi: 'Kombi (70/30)',
}

export const VarianteStatus = {
  ACTIVE: 'active',
  INACTIVE: 'inactive',
} as const
export type VarianteStatus =
  (typeof VarianteStatus)[keyof typeof VarianteStatus]

// Die Anzeigereihenfolge selbst liefert das Backend fertig sortiert; das
// Frontend kennt sie nicht als Wert.
export const Richtung = {
  HOCH: 'hoch',
  RUNTER: 'runter',
} as const
export type Richtung = (typeof Richtung)[keyof typeof Richtung]
export const RichtungSchema = z.enum([Richtung.HOCH, Richtung.RUNTER])

// Eingaberegeln gespiegelt an den zog-Grenzen des Backends (Regel 5): das
// Formular nennt die Grenze, statt einen anonymen validation_error abzuwarten.
// Produkt- und Variantenname teilen dieselbe Regel wie NameSchema in domain/produkt.
export const NameEingabeSchema = createNameSchema(100)
export const PreisCentsEingabeSchema = z
  .number()
  .int()
  .min(1, { message: 'Preis muss mindestens 1 Cent betragen.' })
  .max(99999, { message: 'Preis darf maximal 999,99 € betragen.' })

export function defaultSteuersatzByKategorie(kategorie: Kategorie): Steuersatz {
  if (kategorie === Kategorie.ESSEN) {
    return Steuersatz.ERMAESSIGT
  }

  return Steuersatz.REGEL
}
