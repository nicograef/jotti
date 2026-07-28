import { useQuery } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import type { Tisch } from './Tisch'
import { TischBackend } from './TischBackend'

export const tischBackend = new TischBackend(BackendSingleton)

export const ALLE_TISCHE_KEY = 'alle-tische'

// Kein Stammdatum trotz der Tisch-Stammdaten im Übrigen: Die Nutzlast trägt
// `saldoCents` — den offenen Tisch-Saldo der laufenden Kassensitzung. Er steuert
// den Löschen-/Deaktivieren-Guard und die Kopfzeile „N mit offenem Saldo"; ein
// veralteter Wert gäbe einen soeben bebuchten Tisch zum Deaktivieren frei.
export function useAllTische() {
  const { data: tische = [] as Tisch[], isPending } = useQuery({
    queryKey: [ALLE_TISCHE_KEY],
    queryFn: () => tischBackend.getAllTische(),
  })
  return { tische, isPending }
}
