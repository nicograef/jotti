import type { ReactNode } from 'react'

// Die zwei Panels neben der Benutzertabelle (Design-Handoff 1e). Links die
// Tabelle, rechts das Onboarding-Verfahren und die Rechte-Erklärung — beide mit
// der Copy aus dem Handoff. Schlichte Karten im Stil der übrigen Admin-Seiten.
function OnboardingSchritt({
  nummer,
  children,
}: {
  nummer: number
  children: ReactNode
}) {
  return (
    <div className="flex gap-2.5 text-sm leading-relaxed text-muted-foreground">
      <span className="flex size-5 shrink-0 items-center justify-center rounded-full border bg-background text-xs font-bold text-foreground">
        {nummer}
      </span>
      <span>{children}</span>
    </div>
  )
}

export function HelferPanels() {
  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-col gap-2.5 rounded-lg border bg-sidebar p-4">
        <span className="text-sm font-semibold">So kommt ein Helfer rein</span>
        <OnboardingSchritt nummer={1}>
          „Neuer Helfer“ anlegen — jotti erzeugt ein{' '}
          <strong className="text-foreground">Einmalpasswort</strong>.
        </OnboardingSchritt>
        <OnboardingSchritt nummer={2}>
          Passwort dem Helfer zeigen (oder QR-Code scannen lassen).
        </OnboardingSchritt>
        <OnboardingSchritt nummer={3}>
          Helfer meldet sich am eigenen Handy an und wählt ein eigenes Passwort.
        </OnboardingSchritt>
      </div>

      <div className="flex flex-col gap-2 rounded-lg border p-4">
        <span className="text-sm font-semibold">Was Rollen dürfen</span>
        <span className="text-sm leading-relaxed text-muted-foreground">
          <strong className="text-foreground">Service</strong> bestellt &amp;
          kassiert · <strong className="text-foreground">Serviceleitung</strong>{' '}
          darf zusätzlich stornieren ·{' '}
          <strong className="text-foreground">Admin</strong> verwaltet alles
          hier.
        </span>
        <span className="text-sm leading-relaxed text-muted-foreground">
          Passwort vergessen? „···“ →{' '}
          <strong className="text-foreground">Passwort zurücksetzen</strong>{' '}
          erzeugt ein neues Einmalpasswort.
        </span>
      </div>
    </div>
  )
}
