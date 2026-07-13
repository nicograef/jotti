interface ActionHintProps {
  // Grund, warum die Primäraktion deaktiviert ist; null, sobald keine
  // Bedingung mehr fehlt.
  reason: string | null
}

// Nennt sichtbar den Grund, warum die Primäraktion gesperrt ist, direkt über
// dem Button — ohne auf eine Berührung zu warten. Rendert nichts, sobald die
// Bedingung erfüllt ist (reason === null).
export function ActionHint({ reason }: ActionHintProps) {
  if (reason === null) {
    return null
  }

  return (
    <p role="status" className="text-center text-sm text-muted-foreground">
      {reason}
    </p>
  )
}
