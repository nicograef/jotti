import { ExternalLink, Info } from 'lucide-react'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

const DOCS_BASE = 'https://jotti.rocks/docs'
const LEITFADEN_URL = `${DOCS_BASE}/leitfaden/was-ist-jotti/`

export function GutZuWissenSection() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Gut zu wissen</CardTitle>
      </CardHeader>
      <CardContent className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <a
          href={LEITFADEN_URL}
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-start gap-2 rounded-lg border px-4 py-3 hover:bg-accent"
        >
          <ExternalLink
            className="mt-0.5 size-4 shrink-0 text-muted-foreground"
            aria-hidden
          />
          <span>
            <span className="text-sm font-medium underline underline-offset-4">
              Leitfaden für Vereine
            </span>
            <span className="block text-sm text-muted-foreground">
              Eure Pflichten gegenüber dem Finanzamt, Schritt für Schritt.
            </span>
          </span>
        </a>
        <div className="flex items-start gap-2 rounded-lg border px-4 py-3">
          <Info
            className="mt-0.5 size-4 shrink-0 text-muted-foreground"
            aria-hidden
          />
          <span>
            <span className="text-sm font-medium">10 Jahre aufbewahren</span>
            <span className="block text-sm text-muted-foreground">
              Kassendaten regelmäßig sichern — auch nach dem Fest.{' '}
              <span className="text-xs">(§ 147 AO)</span>
            </span>
          </span>
        </div>
      </CardContent>
    </Card>
  )
}
