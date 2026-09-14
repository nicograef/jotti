import { isRouteErrorResponse, useNavigate, useRouteError } from 'react-router'

import { AuthLayout } from '@/components/common/AuthLayout'
import { Wortmarke } from '@/components/common/Wortmarke'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardFooter, CardHeader } from '@/components/ui/card'

interface FehlerAnzeigeProps {
  titel: string
  text: string
}

// Verhindert React Routers rohen "Unexpected Application Error!"-Bildschirm.
function FehlerAnzeige({ titel, text }: FehlerAnzeigeProps) {
  const navigate = useNavigate()

  return (
    <AuthLayout>
      <Card className="w-full max-w-sm">
        <CardHeader>
          <Wortmarke as="h1" className="text-[38px] text-center" />
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
