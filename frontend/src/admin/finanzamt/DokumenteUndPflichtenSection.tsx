import { ExternalLink } from 'lucide-react'

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

const DOCS_BASE = 'https://jotti.rocks/docs'

const dokumente = [
  {
    titel: 'Leitfaden für Vereine',
    beschreibung: 'Was ihr als Verein gegenüber dem Finanzamt erledigen müsst.',
    url: `${DOCS_BASE}/leitfaden/was-ist-jotti/`,
  },
  {
    titel: 'Compliance-Überblick',
    beschreibung: 'Rechtsgrundlagen (KassenSichV, GoBD, DSFinV-K, ELSTER).',
    url: `${DOCS_BASE}/compliance/`,
  },
]

export function DokumenteUndPflichtenSection() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Dokumente und Pflichten</CardTitle>
        <CardDescription>
          Alle Kassendaten müssen zehn Jahre aufbewahrt werden (§ 147 AO). Sorgt
          für regelmäßige Backups, damit die Daten auch nach der Veranstaltung
          vollständig und unverändert verfügbar bleiben.
        </CardDescription>
      </CardHeader>
      <CardContent className="grid gap-3">
        {dokumente.map((dokument) => (
          <a
            key={dokument.url}
            href={dokument.url}
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-start gap-2 rounded-md border px-3 py-2 hover:bg-accent"
          >
            <ExternalLink className="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" />
            <span>
              <span className="font-medium underline underline-offset-4">
                {dokument.titel}
              </span>
              <span className="block text-sm text-muted-foreground">
                {dokument.beschreibung}
              </span>
            </span>
          </a>
        ))}
      </CardContent>
    </Card>
  )
}
