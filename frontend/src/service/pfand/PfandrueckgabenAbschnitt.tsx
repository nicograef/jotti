import { LadefehlerAlert } from '@/components/common/LadefehlerAlert'

import type { PositionRef } from '../schemas'
import { VorgangHistorie } from '../VorgangHistorie'
import { pfandBackend, usePfandrueckgabeHistorie } from './hooks'

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

  const stornieren = (
    rueckgabeId: string,
    stornierungId: string,
    positionen: PositionRef[],
    kommentar: string,
  ) =>
    pfandBackend.pfandrueckgabeStornieren({
      stornierungId,
      rueckgabeId,
      positionen,
      kommentar,
    })

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
