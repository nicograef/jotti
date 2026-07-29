import { Button } from '@/components/ui/button'
import { useVersionsGuard } from '@/hooks/use-versions-guard'
import { seiteNeuLaden } from '@/lib/reload'

/**
 * Hinweis zum Versions-Handshake: Er erscheint genau dann, wenn der erzwungene
 * Reload ansteht, aber nicht laufen kann.
 *
 * Der Regelfall ist unsichtbar — bei leerem Vorgangs-Register lädt der Guard
 * sofort neu, ohne dass hier je etwas zu sehen ist. Sichtbar wird der Hinweis
 * nur, solange ein Vorgang offen ist (`wartet`) oder ein Reload wirkungslos
 * geblieben ist (`gebremst`).
 *
 * Der Hinweis ist nicht wegklickbar: Er trägt keine Schließen-Aktion, hört
 * weder auf Escape noch auf einen Klick daneben und verschwindet allein, wenn
 * der Guard ihn zurücknimmt. Bewusst ist er dabei **kein** modaler Dialog —
 * `AlertDialog` fängt in Radix zwingend den Fokus und legt die Seite darunter
 * still (`modal: true` ist dort fest verdrahtet). Genau das darf hier nicht
 * passieren: Der Hinweis wartet auf das Abschließen oder Verwerfen des
 * laufenden Vorgangs und darf die Bedienung, die dafür nötig ist, nicht
 * blockieren.
 *
 * Er steht deshalb wie sein Vorgänger im Fluss über dem Seitenlayout und
 * schwebt nicht als fixierte Leiste: Am unteren Rand läge er über dem
 * ServiceDock und der Fußleiste der Tischauswahl, am oberen über den fixierten
 * Kopfleisten — überall dort sitzen Bedienelemente. Im Fluss verdrängt er sie,
 * statt sie zu verdecken.
 *
 * Der Text nennt bewusst keine Richtung: Der Handshake meldet jede Abweichung,
 * nach einem Rollback also auch eine ältere Serverversion.
 */
export function VersionsHinweis() {
  const versionsZustand = useVersionsGuard()

  if (versionsZustand === 'aus' || versionsZustand === 'laedt') return null

  return (
    <div
      role="alert"
      // `relative z-[60] pointer-events-auto` hebt den Hinweis über ein
      // offenes Radix-Modal. Mehrere der meldenden Vorgänge leben nur, solange
      // ein Dialog oder Drawer offen ist; dessen Overlay liegt als
      // `fixed inset-0 z-50` mit Backdrop-Blur darüber und legt den Rest der
      // Seite per `body { pointer-events: none }` still. Ohne eigenen
      // Stapelplatz stünde der Hinweis verwaschen dahinter, und „Jetzt neu
      // laden" — der einzige Ausweg aus dem gebremsten Zustand — wäre nicht zu
      // treffen. Alle drei zusammen sind nötig: `z-index` wirkt nur auf
      // positionierten Elementen (`relative`), und die Zeigersperre hebt allein
      // `pointer-events-auto` auf. Ein Tippen auf die Hinweisfläche schließt
      // damit das Modal nicht mehr — gewollt, denn dort liegt die Schaltfläche.
      className="relative z-[60] pointer-events-auto flex flex-col items-center justify-center gap-2 bg-primary px-4 py-2 text-center text-sm text-primary-foreground print:hidden sm:flex-row"
    >
      {versionsZustand === 'wartet' ? (
        <p>
          Der Server läuft mit einer anderen Version als diese Seite. Bitte den
          laufenden Vorgang abschließen oder verwerfen — danach lädt sich die
          Seite von selbst neu.
        </p>
      ) : (
        <>
          <p>
            Der Server läuft mit einer anderen Version als diese Seite. Das
            automatische Neuladen hat nicht geklappt — bitte von Hand neu laden.
          </p>
          <Button variant="secondary" size="sm" onClick={seiteNeuLaden}>
            Jetzt neu laden
          </Button>
        </>
      )}
    </div>
  )
}
