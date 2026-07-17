import { useMutation, useQuery } from '@tanstack/react-query'
import { toast } from 'sonner'

import { BackendSingleton } from '@/lib/Backend'
import { triggerBrowserDownload } from '@/lib/download'
import { getActionErrorMessage } from '@/lib/errorMessages'

import { ReportingBackend } from './ReportingBackend'
import type {
  AbgeschlosseneSitzung,
  LiveReportingData,
  ReportingData,
} from './types'

const reportingBackend = new ReportingBackend(BackendSingleton)

// Query-Keys der Reporting-Ansichten.
export const ABGESCHLOSSENE_KASSENSITZUNGEN_KEY =
  'abgeschlossene-kassensitzungen'
export const REPORT_KEY = 'report'
export const LIVE_REPORTING_KEY = 'live-reporting'

export function useAbgeschlosseneKassensitzungen() {
  const { data: kassensitzungen = [] as AbgeschlosseneSitzung[], isPending } =
    useQuery({
      queryKey: [ABGESCHLOSSENE_KASSENSITZUNGEN_KEY],
      queryFn: () => reportingBackend.getAbgeschlosseneKassensitzungen(),
    })
  return { kassensitzungen, isPending }
}

export function useReport(kassensitzungNr: number | null) {
  const { data: result = null as ReportingData | null, isPending } = useQuery({
    queryKey: [REPORT_KEY, kassensitzungNr],
    queryFn: () => reportingBackend.getReporting(kassensitzungNr ?? 0),
    enabled: kassensitzungNr !== null,
  })
  return { result, isPending }
}

export function useLiveReporting() {
  const {
    data: liveData = null as LiveReportingData | null,
    isPending,
    dataUpdatedAt,
    refetch,
  } = useQuery({
    queryKey: [LIVE_REPORTING_KEY],
    queryFn: () => reportingBackend.getLiveReporting(),
    // Auto-Refresh: das Live-Dashboard aktualisiert sich alle 30 s ohne
    // Interaktion (lokaler Server, eine Admin-Session).
    refetchInterval: 30_000,
  })
  return { liveData, isPending, dataUpdatedAt, refetch }
}

export function useDsfinvkExport() {
  const mutation = useMutation({
    mutationFn: (kassensitzungNr: number | null) =>
      reportingBackend.exportDsfinvk(kassensitzungNr),
    onSuccess: ({ blob, filename }) => {
      triggerBrowserDownload(blob, filename)
      toast.success('DSFinV-K-Archiv heruntergeladen.')
    },
    onError: (error) => {
      toast.error(
        getActionErrorMessage({
          actionLabel: 'DSFinV-K-Export',
          error,
          byCode: {
            leere_kassensitzung:
              'Diese Kassensitzung enthält keine Vorgänge zum Exportieren.',
            kassensitzung_nicht_gefunden:
              'Die gewählte Kassensitzung wurde nicht gefunden. Bitte neu auswählen und erneut versuchen.',
          },
        }),
      )
    },
  })
  return { exportieren: mutation.mutate, isPending: mutation.isPending }
}
