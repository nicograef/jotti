import { useQuery } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import { ReportingBackend } from './ReportingBackend'
import type { Kassensitzung, ReportingData } from './types'

const reportingBackend = new ReportingBackend(BackendSingleton)

export function useKassensitzungen() {
  const { data: kassensitzungen = [] as Kassensitzung[], isPending } = useQuery(
    {
      queryKey: ['kassensitzungen'],
      queryFn: () => reportingBackend.getAllKassensitzungen(),
    },
  )
  return { kassensitzungen, loading: isPending }
}

export function useReport(kassensitzungNr: number | null) {
  const { data: result = null as ReportingData | null, isPending } = useQuery({
    queryKey: ['report', kassensitzungNr],
    queryFn: () => reportingBackend.getReporting(kassensitzungNr ?? 0),
    enabled: kassensitzungNr !== null,
  })
  return { result, loading: isPending }
}
