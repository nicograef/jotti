import { Check, Copy, Info } from 'lucide-react'
import { useState } from 'react'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Label } from '@/components/ui/label'

import { useKassenidentitaet } from './hooks'

const LEITFADEN_URL = 'https://jotti.rocks/docs/leitfaden/finanzamt-anmelden/'

export function KassenidentitaetSection() {
  const { data: kassenidentitaet, isPending, error } = useKassenidentitaet()
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    if (!kassenidentitaet) return
    await navigator.clipboard.writeText(kassenidentitaet.seriennummer)
    setCopied(true)
    setTimeout(() => {
      setCopied(false)
    }, 2000)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Kassenidentität</CardTitle>
        <CardDescription>
          Die Seriennummer identifiziert diese jotti-Instanz eindeutig. Sie wird
          für die ELSTER-Meldung (§ 146a AO) benötigt und erscheint zusätzlich
          auf jedem Kassenbeleg. Das Anlegedatum dokumentiert, wann diese
          Kassenidentität in jotti erzeugt wurde.
        </CardDescription>
      </CardHeader>
      <CardContent className="grid gap-4">
        {isPending && (
          <p className="text-muted-foreground text-sm">Lade Kassenidentität…</p>
        )}
        {error && (
          <p className="text-destructive text-sm">
            Fehler beim Laden der Kassenidentität.
          </p>
        )}
        {kassenidentitaet && (
          <div className="grid gap-4">
            <div className="grid gap-1.5">
              <Label>Seriennummer</Label>
              <div className="flex items-center gap-2">
                <code className="flex-1 rounded-md border bg-muted px-3 py-2 font-mono text-sm">
                  {kassenidentitaet.seriennummer}
                </code>
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => void handleCopy()}
                  aria-label="Seriennummer kopieren"
                >
                  {copied ? (
                    <Check className="h-4 w-4" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </Button>
              </div>
            </div>
            <div className="grid gap-1.5">
              <Label>Anlegedatum</Label>
              <p className="text-sm">
                {new Date(kassenidentitaet.angelegtAm).toLocaleDateString(
                  'de-DE',
                )}
              </p>
            </div>
          </div>
        )}
        <Alert>
          <Info className="size-4" />
          <AlertTitle>Meldepflicht beim Finanzamt</AlertTitle>
          <AlertDescription>
            Diese Kasse muss innerhalb eines Monats nach Inbetriebnahme über
            ELSTER beim Finanzamt gemeldet werden (§ 146a Abs. 4 AO). Verwendet
            dafür die Seriennummer oben.{' '}
            <a href={LEITFADEN_URL} target="_blank" rel="noopener noreferrer">
              Schritt-für-Schritt-Anleitung
            </a>
            .
          </AlertDescription>
        </Alert>
      </CardContent>
    </Card>
  )
}
