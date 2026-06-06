import { useQuery } from '@tanstack/react-query'

import { BackendSingleton } from '@/lib/Backend'

import { ReportingBackend } from './ReportingBackend'
import type { Kassensitzung, LiveReportingData, ReportingData } from './types'

const reportingBackend = new ReportingBackend(BackendSingleton)

export function useKassensitzungen() {
  const { data: kassensitzungen = [] as Kassensitzung[], isPending } = useQuery(
    {
      queryKey: ['kassensitzungen'],
      queryFn: () => reportingBackend.getAllKassensitzungen(),
    },
  )
  return { kassensitzungen, isPending }
}

export function useReport(kassensitzungNr: number | null) {
  const { data: result = null as ReportingData | null, isPending } = useQuery({
    queryKey: ['report', kassensitzungNr],
    queryFn: () => reportingBackend.getReporting(kassensitzungNr ?? 0),
    enabled: kassensitzungNr !== null,
  })
  return { result, isPending }
}

export function useLiveReporting() {
  const { data: liveData = null as LiveReportingData | null, isPending } =
    useQuery({
      queryKey: ['live-reporting'],
      queryFn: () => reportingBackend.getLiveReporting(),
    })
  return { liveData, isPending }
}
