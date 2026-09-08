import { z } from 'zod'

// Namensschema für Tisch-, Produkt-, Varianten- und Benutzernamen: die eine
// Quelle der Meldungen und der Mindestlänge, weil dieselbe Regel sonst in jedem
// Bereich neu geschrieben wird. Nur die Obergrenze unterscheidet sich, deshalb
// ist sie das Argument.
//
// Gespiegelt am zog-Gegenstück des Backends (Regel 5): `.trim()` entspricht
// dessen `.Trim()`, damit ein Name aus Leerzeichen schon im Formular auffällt
// statt als anonymer validation_error zurückzukommen. Die Grenzen sind 100
// Zeichen für Tisch, Produkt und Variante (domain/tisch, domain/produkt) und 50
// für den Benutzernamen (domain/user).
export function createNameSchema(maxLength: number) {
  return z
    .string()
    .trim()
    .min(3, { message: 'Das sieht nicht nach einem echten Namen aus.' })
    .max(maxLength, { message: 'Der Name ist zu lang.' })
}
