import { useEffect, useRef, useState } from 'react'

import { useAnzahlOffeneVorgaenge } from '@/hooks/use-anzahl-offene-vorgaenge'
import { useVersion } from '@/hooks/use-version'
import { seiteNeuLaden } from '@/lib/reload'
import { CLIENT_VERSION, istVersionsabweichung } from '@/lib/version'

/**
 * `sessionStorage` und nicht `localStorage`: Der Vermerk der letzten
 * Reload-Zielversion soll den Reload überleben, nicht das Schließen des Tabs.
 */
export const RELOAD_VERMERK_SCHLUESSEL = 'JOTTI_RELOAD_ZIELVERSION'

/**
 * - `aus` — die Versionen passen, oder es gibt noch keine Antwort.
 * - `laedt` — die Seite lädt neu.
 * - `wartet` — Abweichung bei offenem Vorgang; der Reload folgt, sobald das
 *   Register leer wird.
 * - `gebremst` — ein Reload blieb wirkungslos; dieser Client lädt nicht mehr
 *   von selbst.
 */
export type VersionsZustand = 'aus' | 'laedt' | 'wartet' | 'gebremst'

/**
 * Trägt dieser Client nicht die vermerkte Zielversion, war der Reload
 * wirkungslos und es darf kein zweiter folgen. Eingelöst wird der Vermerk erst
 * bei Einigkeit mit dem Server (`useVersionsGuard`).
 *
 * Ohne die Bremse läuft das Update-Fenster in eine Endlosschleife: Das Backend
 * wird vor dem Frontend ersetzt (`docker-compose.prod.yml`, `depends_on`) und
 * meldet die neue Version, während der alte Container das alte Bundle
 * ausliefert. Ein Limit je Seitenleben hilft nicht — jeder Reload beginnt eines.
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
  // undefined heißt: keine beantwortete Abfrage. Ein Serverneustart lässt sie
  // scheitern und darf keinen Reload erzwingen, nur ein Versionswechsel.
  if (serverVersion === undefined) return 'aus'
  if (!istVersionsabweichung(CLIENT_VERSION, serverVersion)) return 'aus'
  if (gebremst) return 'gebremst'
  if (anzahlOffeneVorgaenge > 0) return 'wartet'
  return 'laedt'
}

/**
 * Erzwingt den Reload, sobald Server und Client verschiedene Releases tragen —
 * aber nie über einen offenen Vorgang hinweg: Bei leerem Register sofort, sonst
 * `wartet`, bis der letzte Vorgang abgeschlossen oder verworfen ist. Ausgelöst
 * wird im Effekt, denn ein Reload ist eine Nebenwirkung.
 */
export function useVersionsGuard(): VersionsZustand {
  const serverVersion = useVersion()
  const anzahlOffeneVorgaenge = useAnzahlOffeneVorgaenge()
  const [gebremst, setGebremst] = useState(bremseAuswerten)
  const bereitsGeladen = useRef(false)

  // Einigkeit löst den Vermerk ein, gleich welche Version in ihm steht: Nach
  // Rollback oder Vorwärts-Korrektur trägt der Client eine andere als die
  // vermerkte Zielversion. Löste nur der exakte Treffer ein, bliebe der Vermerk
  // für die Lebensdauer des Tabs stehen und entschärfte jede spätere Erkennung.
  const einigMitServer =
    serverVersion !== undefined &&
    !istVersionsabweichung(CLIENT_VERSION, serverVersion)

  // Mit dem Vermerk fällt auch das eingefrorene Flag: In der als App
  // installierten jotti bleibt ein Tab wochenlang offen, ein zweites
  // Seitenleben kommt womöglich nie. Eine Schleife kann daraus nicht entstehen
  // — sie setzt eine Abweichung voraus. Das Anpassen von Zustand während des
  // Renderns ist dafür vorgesehen und ändert am Ergebnis dieses Renders nichts.
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
    // Zwischen Aufruf und Entladen läuft die Anwendung weiter und der Effekt
    // kann erneut laufen. Ein zweiter Reload darf daraus nie folgen.
    if (bereitsGeladen.current) return

    bereitsGeladen.current = true
    sessionStorage.setItem(RELOAD_VERMERK_SCHLUESSEL, serverVersion)
    seiteNeuLaden()
  }, [versionsZustand, serverVersion])

  return versionsZustand
}
