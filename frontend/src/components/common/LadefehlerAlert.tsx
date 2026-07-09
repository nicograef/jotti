import { TriangleAlert } from 'lucide-react'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'

interface LadefehlerAlertProps {
  titel: string
  onErneutVersuchen: () => void
  className?: string
}

// LadefehlerAlert zeigt einen destruktiven Fehlerhinweis mit Titel und einem
// „Erneut versuchen“-Button an. Genutzt, wenn eine Query fehlschlägt und ein
// leerer Standardzustand irreführend wäre (z. B. Tisch scheinbar abgerechnet).
export function LadefehlerAlert({
  titel,
  onErneutVersuchen,
  className,
}: LadefehlerAlertProps) {
  return (
    <Alert variant="destructive" className={className}>
      <TriangleAlert className="size-4" />
      <AlertTitle>{titel}</AlertTitle>
      <AlertDescription>
        <p>Bitte die Verbindung prüfen und erneut versuchen.</p>
        <Button variant="outline" size="sm" onClick={onErneutVersuchen}>
          Erneut versuchen
        </Button>
      </AlertDescription>
    </Alert>
  )
}
