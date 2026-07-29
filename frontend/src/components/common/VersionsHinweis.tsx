import { useVersion } from '@/hooks/use-version'
import { CLIENT_VERSION, istVersionsabweichung } from '@/lib/version'

/**
 * Nicht-blockierender Hinweis, dass der Server eine andere Release-Version
 * ausliefert als dieser Client geladen hat.
 *
 * Erscheint nur, wenn beide Seiten echte Releases sind — in Dev, in E2E und in
 * Tests steht auf beiden Seiten der Default `dev`, dort bleibt der Hinweis aus.
 *
 * Der Hinweis steht im Fluss über dem Seitenlayout und schwebt bewusst nicht
 * als fixierte Leiste: Am unteren Rand läge er über dem ServiceDock und der
 * Fußleiste der Tischauswahl, am oberen über den fixierten Kopfleisten — überall
 * dort sitzen Bedienelemente. Im Fluss verdrängt er sie, statt sie zu verdecken.
 *
 * Der Text nennt bewusst keine Richtung: `istVersionsabweichung` meldet jede
 * Abweichung, nach einem Rollback also auch eine ältere Serverversion.
 */
export function VersionsHinweis() {
  const serverVersion = useVersion()

  if (
    serverVersion === undefined ||
    !istVersionsabweichung(CLIENT_VERSION, serverVersion)
  ) {
    return null
  }

  return (
    <div
      role="status"
      className="bg-primary px-4 py-2 text-center text-sm text-primary-foreground print:hidden"
    >
      Der Server läuft mit Version {serverVersion}, diese Seite mit einer
      anderen. Bitte die Seite neu laden.
    </div>
  )
}
