import { useEffect, useRef, useState } from 'react'

import { useAnzahlOffeneVorgaenge } from '@/hooks/use-anzahl-offene-vorgaenge'
import { useVersion } from '@/hooks/use-version'
import { Seite } from '@/lib/reload'
import { CLIENT_VERSION, istVersionsabweichung } from '@/lib/version'

/**
 * `sessionStorage` und nicht `localStorage`: Der Vermerk der letzten
 * Reload-Zielversion soll den Reload überleben, nicht das Schließen des Tabs.
 */
export const RELOAD_VERMERK_SCHLUESSEL = 'JOTTI_RELOAD_ZIELVERSION'

/**
 * `wartet`: mismatch with an open Vorgang; the reload follows once the register is empty.
 * `gebremst`: a reload had no effect, so this client no longer reloads on its own.
 */
export type VersionsZustand = 'aus' | 'laedt' | 'wartet' | 'gebremst'

/**
 * A client without the noted target version had an ineffective reload, so no second one may follow.
 * Without this brake the update window loops forever (docs/handbuch.md §6.8, Schleifenbremse).
 */
function bremseAuswerten(clientVersion: string): boolean {
  const zielVersion = sessionStorage.getItem(RELOAD_VERMERK_SCHLUESSEL)
  return zielVersion !== null && zielVersion !== clientVersion
}

function bestimmeVersionsZustand(
  clientVersion: string,
  serverVersion: string | undefined,
  gebremst: boolean,
  anzahlOffeneVorgaenge: number,
): VersionsZustand {
  // undefined heißt: keine beantwortete Abfrage. Ein Serverneustart lässt sie
  // scheitern und darf keinen Reload erzwingen, nur ein Versionswechsel.
  if (serverVersion === undefined) return 'aus'
  if (!istVersionsabweichung(clientVersion, serverVersion)) return 'aus'
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
  const [gebremst, setGebremst] = useState(() =>
    bremseAuswerten(CLIENT_VERSION),
  )
  const bereitsGeladen = useRef(false)

  // Agreement clears the note whatever version it holds: after a rollback or forward fix the client differs from the target.
  // Clearing only on an exact match would leave the note for the tab's lifetime and disarm every later detection.
  const einigMitServer =
    serverVersion !== undefined &&
    !istVersionsabweichung(CLIENT_VERSION, serverVersion)

  // The frozen flag falls with the note: an installed jotti tab stays open for weeks, so a second page life may never come.
  // No loop can follow, since a loop needs a mismatch; adjusting state during render is the intended React pattern here.
  if (gebremst && einigMitServer) setGebremst(false)

  const versionsZustand = bestimmeVersionsZustand(
    CLIENT_VERSION,
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
    Seite.neuLaden()
  }, [versionsZustand, serverVersion])

  return versionsZustand
}
