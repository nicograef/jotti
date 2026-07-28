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

// Zeitlimit des DSFinV-K-Exports. Das Backend baut das Archiv einer ganzen
// Kassensitzung und überträgt es anschließend; für den Schreibvorgang der
// Antwort räumt es sich exportWriteTimeout = 5 Minuten ein
// (backend/api/fiskal/export/http/handler.go). Der Client wartet 30 s länger:
// Client-Budget = Schreibbudget des Servers + Netzreserve. Der Aufschlag gilt
// der Antwort: Ein spät, aber erfolgreich geschriebenes Archiv soll hier noch
// ankommen. Wäre dieses Budget nicht größer, hätte der Client in genau dem
// Fenster schon aufgegeben, in dem der Server gerade noch schreibt.
const EXPORT_TIMEOUT_MS = 330_000

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
      { zeitlimitMs: EXPORT_TIMEOUT_MS },
    )
  }
}
