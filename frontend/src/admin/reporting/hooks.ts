import { useMutation, useQuery } from '@tanstack/react-query'
import { toast } from 'sonner'

import { BackendSingleton } from '@/lib/Backend'
import { triggerBrowserDownload } from '@/lib/download'
import { getActionErrorMessage } from '@/lib/errorMessages'

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
