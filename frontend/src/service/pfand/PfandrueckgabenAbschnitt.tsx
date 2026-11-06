import { useRef } from 'react'

import { LadefehlerAlert } from '@/components/common/LadefehlerAlert'

import type { PositionRef } from '../schemas'
import { VorgangHistorie } from '../VorgangHistorie'
import { usePfandrueckgabeHistorie } from './hooks'
import { pfandBackend } from './PfandBackend'

interface PfandrueckgabenAbschnittProps {
  // 0 ist der Direktverkauf.
  tischId: number
  onErfolg: (nachricht: string) => void
}

export function PfandrueckgabenAbschnitt({
  tischId,
  onErfolg,
}: PfandrueckgabenAbschnittProps) {
  const { historie, isPending, isError, refetch } =
    usePfandrueckgabeHistorie(tischId)

  // Ein Wiederholversuch derselben Auswahl behält die stornierungId, damit eine
  // verlorene Antwort nicht doppelt bucht; eine andere Auswahl bekommt eine neue.
  const versuch = useRef<{ schluessel: string; stornierungId: string } | null>(
    null,
  )
  const stornieren = async (
    rueckgabeId: string,
    positionen: PositionRef[],
    kommentar: string,
  ) => {
    const schluessel = JSON.stringify([rueckgabeId, positionen])
    if (versuch.current?.schluessel !== schluessel) {
      versuch.current = { schluessel, stornierungId: crypto.randomUUID() }
    }
    await pfandBackend.pfandrueckgabeStornieren({
      stornierungId: versuch.current.stornierungId,
      rueckgabeId,
      positionen,
      kommentar,
    })
    versuch.current = null
  }

  if (isError) {
    return (
      <LadefehlerAlert
        titel="Pfandrückgaben konnten nicht geladen werden"
        onErneutVersuchen={() => {
          void refetch()
        }}
        className="my-4"
      />
    )
  }

  if (isPending || historie.length === 0) {
    return null
  }

  return (
    <section aria-label="Pfandrückgaben" className="mt-6">
      <h2 className="text-lg font-semibold">Pfandrückgaben</h2>
      <VorgangHistorie
        eintraege={historie.map((rueckgabe) => ({
          ...rueckgabe,
          id: rueckgabe.rueckgabeId,
        }))}
        loading={false}
        bezeichnung="Pfandrückgabe"
        leerHinweis="Noch keine Pfandrückgaben in dieser Kassensitzung."
        stornieren={stornieren}
        belegDrucken={(rueckgabeId, stornierungId) =>
          pfandBackend.pfandbelegDrucken({ rueckgabeId, stornierungId })
        }
        onErfolg={onErfolg}
      />
    </section>
  )
}
