import { isRouteErrorResponse, useNavigate, useRouteError } from 'react-router'

import { AuthLayout } from '@/components/common/AuthLayout'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardFooter, CardHeader } from '@/components/ui/card'

interface FehlerAnzeigeProps {
  titel: string
  text: string
}

// Gemeinsames Layout für die 404-Route und die Router-ErrorBoundary,
// damit kein "Unexpected Application Error!"-Rohbildschirm mehr erreichbar ist.
function FehlerAnzeige({ titel, text }: FehlerAnzeigeProps) {
  const navigate = useNavigate()

  return (
    <AuthLayout>
      <Card className="w-full max-w-sm">
        <CardHeader>
          <h1 className="text-4xl text-center font-extrabold">jotti</h1>
        </CardHeader>
        <CardContent className="text-center space-y-2">
          <h2 className="text-lg font-semibold">{titel}</h2>
          <p className="text-muted-foreground text-sm">{text}</p>
        </CardContent>
        <CardFooter>
          <Button className="w-full" onClick={() => void navigate('/')}>
            Zurück zur Startseite
          </Button>
        </CardFooter>
      </Card>
    </AuthLayout>
  )
}

// Fängt Render-/Loader-Fehler der Root-Route ab (ErrorBoundary).
export function ErrorPage() {
  const error = useRouteError()
  const istUnbekannterPfad = isRouteErrorResponse(error) && error.status === 404

  return istUnbekannterPfad ? (
    <NotFoundPage />
  ) : (
    <FehlerAnzeige
      titel="Ein Fehler ist aufgetreten"
      text="Etwas ist schiefgelaufen. Bitte versuche es erneut."
    />
  )
}

// Catch-all-Route für unbekannte Pfade (kein Router-Fehler, daher eigene Route).
export function NotFoundPage() {
  return (
    <FehlerAnzeige
      titel="Seite nicht gefunden"
      text="Die aufgerufene Seite existiert nicht."
    />
  )
}
