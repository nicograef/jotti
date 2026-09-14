import { toast } from 'sonner'
import { z } from 'zod'

// eingereiht = Druckauftrag angelegt; ausstehend = die TSE-Signatur fehlt noch,
// es entstand kein Druckauftrag — derselbe Endpunkt wird erneut aufgerufen.
export const BelegStatusSchema = z.enum(['eingereiht', 'ausstehend'])
export type BelegStatus = z.infer<typeof BelegStatusSchema>

export const BelegDruckenResponseSchema = z.object({
  status: BelegStatusSchema,
})

const NACHFASS_INTERVALL_MS = 1_500
const NACHFASS_VERSUCHE = 6

const warte = (ms: number) =>
  new Promise<void>((resolve) => setTimeout(resolve, ms))

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

export function meldeBelegStatus(status: BelegStatus, erfolgMeldung: string) {
  if (status === 'eingereiht') {
    toast.success(erfolgMeldung)
  } else {
    toast.info(
      'Die TSE-Signatur steht noch aus. Bitte den Beleg gleich erneut anfordern.',
    )
  }
}
