import { useEffect, useRef, useState } from 'react'

import { useAnzahlOffeneVorgaenge } from '@/hooks/use-anzahl-offene-vorgaenge'
import { useVersion } from '@/hooks/use-version'
import { seiteNeuLaden } from '@/lib/reload'
import { CLIENT_VERSION, istVersionsabweichung } from '@/lib/version'

/**
 * Schlüssel des Vermerks, unter dem der Guard die Zielversion des letzten
 * erzwungenen Reloads ablegt.
 *
 * `sessionStorage` und nicht `localStorage`: Der Vermerk soll den Reload
 * überleben, aber nicht das Schließen des Tabs.
 */
export const RELOAD_VERMERK_SCHLUESSEL = 'JOTTI_RELOAD_ZIELVERSION'

/**
 * Was der Versions-Handshake gerade zu tun hat.
 *
 * - `aus` — kein Anlass: Die Versionen passen, oder es gibt noch keine Antwort.
 * - `laedt` — die Seite lädt neu; bis zum Entladen ist nichts zu zeigen.
 * - `wartet` — es gibt eine Abweichung, aber noch einen offenen Vorgang. Der
 *   Reload folgt von selbst, sobald das Register leer wird.
 * - `gebremst` — ein Reload ist bereits wirkungslos geblieben. Dieser Client
 *   lädt nicht mehr von selbst; es bleibt der Weg von Hand.
 */
export type VersionsZustand = 'aus' | 'laedt' | 'wartet' | 'gebremst'

/**
 * Wertet den Vermerk des letzten erzwungenen Reloads aus und meldet, ob die
 * Schleifenbremse greift.
 *
 * Trägt dieser Client immer noch nicht die vermerkte Zielversion, war der
 * Reload wirkungslos — dann darf kein zweiter folgen. Eingelöst wird der
 * Vermerk nicht hier, sondern erst bei Einigkeit mit dem Server (siehe
 * `useVersionsGuard`).
 *
 * Ohne diese Bremse entsteht im Update-Fenster eine Endlosschleife: Beim
 * Selbsthosting-Update wird das Backend garantiert vor dem Frontend ersetzt
 * (`docker-compose.prod.yml`, `depends_on`), das neue Backend meldet also
 * bereits die neue Version, während der alte Frontend-Container weiterhin das
 * alte Bundle ausliefert. Ein Client darin lädt neu, bekommt dasselbe alte
 * Bundle, sieht dieselbe Abweichung und lädt wieder — auf jedem Helfer-Handy,
 * in der heikelsten Minute eines Updates. „Genau einmal je Seitenleben" schützt
 * nicht davor, weil jeder Reload ein neues Seitenleben beginnt.
 */
function bremseAuswerten(): boolean {
  const zielVersion = sessionStorage.getItem(RELOAD_VERMERK_SCHLUESSEL)
  return zielVersion !== null && zielVersion !== CLIENT_VERSION
}

function bestimmeVersionsZustand(
  serverVersion: string | undefined,
  gebremst: boolean,
  anzahlOffeneVorgaenge: number,
): VersionsZustand {
  // undefined heißt: keine erfolgreich beantwortete Abfrage. Ein Serverneustart
  // lässt sie scheitern und darf keinen Reload erzwingen — nur ein
  // tatsächlicher Versionswechsel.
  if (serverVersion === undefined) return 'aus'
  if (!istVersionsabweichung(CLIENT_VERSION, serverVersion)) return 'aus'
  if (gebremst) return 'gebremst'
  if (anzahlOffeneVorgaenge > 0) return 'wartet'
  return 'laedt'
}

/**
 * Erzwingt den Reload, sobald Server und Client verschiedene Releases tragen —
 * aber nie über einen offenen Vorgang hinweg.
 *
 * Ist das Vorgangs-Register leer, lädt die Seite sofort neu. Ist es das nicht,
 * meldet der Hook `wartet`, und der Reload holt sich seinen Moment, sobald der
 * letzte Vorgang abgeschlossen oder verworfen ist. Der Zustand ist reine
 * Ableitung aus Serverversion, Bremse und Register; ausgelöst wird im Effekt,
 * denn ein Reload ist eine Nebenwirkung und gehört nicht ins Rendern.
 */
export function useVersionsGuard(): VersionsZustand {
  const serverVersion = useVersion()
  const anzahlOffeneVorgaenge = useAnzahlOffeneVorgaenge()
  // Genau einmal je Seitenleben, bevor der Guard zum ersten Mal entscheidet.
  const [gebremst, setGebremst] = useState(bremseAuswerten)
  const bereitsGeladen = useRef(false)

  // Einigkeit mit dem Server beendet den Handshake und löst den Vermerk ein —
  // gleich, welche Version in ihm steht. Nach einem misslungenen Update landet
  // der Client per Rollback oder Vorwärts-Korrektur auf einer anderen als der
  // vermerkten Zielversion; löste nur der exakte Treffer den Vermerk ein,
  // bliebe er für die restliche Lebensdauer des Tabs stehen und entschärfte die
  // Erkennung dauerhaft: Jeder spätere echte Versionswechsel endete im
  // gebremsten Hinweis statt im Reload.
  const einigMitServer =
    serverVersion !== undefined &&
    !istVersionsabweichung(CLIENT_VERSION, serverVersion)

  // Mit dem Vermerk fällt auch das eingefrorene Flag dieses Seitenlebens: In
  // der als App installierten jotti bleibt ein Tab wochenlang offen, ein
  // zweites Seitenleben kommt also womöglich nie. Eine Reload-Schleife kann
  // daraus nicht entstehen — sie setzt eine Abweichung voraus, und genau die
  // gibt es hier nicht. Die Rücknahme gehört ins Rendern und nicht in einen
  // Effekt: React sieht das Anpassen von Zustand während des Renderns dafür
  // vor, und am Ergebnis dieses Renders ändert sie nichts, denn bei Einigkeit
  // meldet der Guard ohnehin `aus`.
  if (gebremst && einigMitServer) setGebremst(false)

  const versionsZustand = bestimmeVersionsZustand(
    serverVersion,
    gebremst,
    anzahlOffeneVorgaenge,
  )

  useEffect(() => {
    if (!einigMitServer) return
    sessionStorage.removeItem(RELOAD_VERMERK_SCHLUESSEL)
  }, [einigMitServer])

  useEffect(() => {
    // `laedt` gibt es nur mit beantworteter Abfrage; die Prüfung auf undefined
    // ist hier bloß der Typnachweis für den Vermerk.
    if (versionsZustand !== 'laedt' || serverVersion === undefined) return
    // Zwischen dem Aufruf und dem tatsächlichen Entladen läuft die Anwendung
    // weiter und der Effekt kann erneut laufen. Ein zweiter Reload darf daraus
    // nie folgen.
    if (bereitsGeladen.current) return

    bereitsGeladen.current = true
    sessionStorage.setItem(RELOAD_VERMERK_SCHLUESSEL, serverVersion)
    seiteNeuLaden()
  }, [versionsZustand, serverVersion])

  return versionsZustand
}
