interface ActionHintProps {
  // Grund, warum die Primäraktion gesperrt ist; `null` = keine Bedingung fehlt.
  reason: string | null
}

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
