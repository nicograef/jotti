import { BackendSingleton } from '@/lib/Backend'
import { useFetch } from '@/lib/useFetch'

import { KasseBackend } from './KasseBackend'
import type { KassensitzungState } from './types'

const kasseBackend = new KasseBackend(BackendSingleton)

export { kasseBackend }

export function useOffeneKassensitzung() {
  const {
    data: kassensitzung,
    setData: setKassensitzung,
    ...rest
  } = useFetch(
    () => kasseBackend.getOffeneKassensitzung(),
    null as KassensitzungState | null,
  )
  return { ...rest, kassensitzung, setKassensitzung }
}

export function useKassenbestand(kassensitzungNr: number | null) {
  const { data: kassenbestand, ...rest } = useFetch(
    async () => {
      if (!kassensitzungNr) return null
      return kasseBackend.getKassenbestand(kassensitzungNr)
    },
    null,
    [kassensitzungNr],
  )
  return { ...rest, kassenbestand }
}
