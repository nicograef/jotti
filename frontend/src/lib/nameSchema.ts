import { z } from 'zod'

// Die eine Quelle der Namensmeldungen und der Mindestlänge; nur die Obergrenze
// unterscheidet sich und ist deshalb das Argument (100 für Tisch, Produkt und
// Variante, 50 für den Namen eines Benutzers). Gespiegelt am zog-Gegenstück des
// Backends: `.trim()` entspricht dessen `.Trim()`, damit ein Name aus
// Leerzeichen im Formular auffällt statt als anonymer validation_error.
// Der Benutzername ist ein anderes Feld (lib/identity.ts).
export function createNameSchema(maxLength: number) {
  return z
    .string()
    .trim()
    .min(3, { message: 'Das sieht nicht nach einem echten Namen aus.' })
    .max(maxLength, { message: 'Der Name ist zu lang.' })
}
