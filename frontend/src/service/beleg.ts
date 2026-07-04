import { z } from 'zod'

// Antwort des Beleg-Abrufs (service/beleg-drucken): eingereiht = Druckauftrag
// angelegt; ausstehend = die TSE-Signatur liegt noch nicht am Signaturauftrag,
// es entstand kein Druckauftrag — derselbe Endpunkt wird erneut aufgerufen.
export const BelegStatusSchema = z.enum(['eingereiht', 'ausstehend'])
export type BelegStatus = z.infer<typeof BelegStatusSchema>

export const BelegDruckenResponseSchema = z.object({
  status: BelegStatusSchema,
})

// Nachfass-Takt bei ausstehender Signatur: alle 1,5 s erneut anfragen,
// insgesamt rund 10 s. Danach zeigt die UI einen Hinweis; der Button bleibt
// für erneutes Anfordern aktiv.
const NACHFASS_INTERVALL_MS = 1_500
const NACHFASS_VERSUCHE = 6

const warte = (ms: number) =>
  new Promise<void>((resolve) => setTimeout(resolve, ms))

// belegDruckenMitNachfassen ruft den Beleg-Abruf auf und fasst bei
// ausstehender TSE-Signatur automatisch nach, bis der Beleg eingereiht ist
// oder das Nachfass-Fenster endet. Liefert den letzten Status.
export async function belegDruckenMitNachfassen(
  anfordern: () => Promise<BelegStatus>,
): Promise<BelegStatus> {
  let status = await anfordern()
  for (
    let versuch = 0;
    status === 'ausstehend' && versuch < NACHFASS_VERSUCHE;
    versuch++
  ) {
    await warte(NACHFASS_INTERVALL_MS)
    status = await anfordern()
  }
  return status
}
