import { z } from 'zod'

import type { BackendClient, DownloadResult } from '@/lib/Backend'

import {
  type AbgeschlosseneSitzung,
  AbgeschlosseneSitzungSchema,
  type LiveReportingData,
  LiveReportingDataSchema,
  type ReportingData,
  ReportingDataSchema,
} from './types'

export class ReportingBackend {
  private readonly backend: BackendClient

  constructor(backend: BackendClient) {
    this.backend = backend
  }

  public async getReporting(kassensitzungNr: number): Promise<ReportingData> {
    return this.backend.post(
      'admin/get-abrechnung',
      { kassensitzungNr },
      ReportingDataSchema,
    )
  }

  public async getLiveReporting(): Promise<LiveReportingData | null> {
    return this.backend.post(
      'admin/get-live-reporting',
      {},
      LiveReportingDataSchema.nullable(),
    )
  }

  public async getAbgeschlosseneKassensitzungen(): Promise<
    AbgeschlosseneSitzung[]
  > {
    const response = await this.backend.post(
      'admin/get-abgeschlossene-kassensitzungen',
      {},
      z.object({ kassensitzungen: z.array(AbgeschlosseneSitzungSchema) }),
    )
    return response.kassensitzungen
  }

  // exportDsfinvk lädt das DSFinV-K-Archiv der gewählten Kassensitzung herunter.
  // Ohne Nummer wählt das Backend die Standard-Sitzung.
  public async exportDsfinvk(
    kassensitzungNr: number | null,
  ): Promise<DownloadResult> {
    return this.backend.download(
      'admin/export/dsfinvk',
      kassensitzungNr ? { kassensitzungNr } : {},
    )
  }
}
