import { useCallback, useState } from 'react'
import { toast } from 'sonner'

import { BackendSingleton } from '@/lib/Backend'

import { ReportingBackend } from './ReportingBackend'
import type { ReportingData } from './types'

const reportingBackend = new ReportingBackend(BackendSingleton)

/** Manages reporting filter state and fetches data on demand. */
export function useReporting() {
  const [kassensitzungNr, setKassensitzungNr] = useState<number | null>(null)
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<ReportingData | null>(null)

  const auswerten = useCallback(async () => {
    if (!kassensitzungNr || kassensitzungNr < 1) {
      toast.error('Bitte eine gültige Kassensitzungs-Nr eingeben.')
      return
    }

    setLoading(true)
    try {
      const data = await reportingBackend.getReporting(kassensitzungNr)
      setResult(data)
    } catch {
      toast.error('Fehler beim Laden des Reportings.')
    } finally {
      setLoading(false)
    }
  }, [kassensitzungNr])

  return {
    kassensitzungNr,
    setKassensitzungNr,
    loading,
    result,
    auswerten: () => void auswerten(),
  }
}
