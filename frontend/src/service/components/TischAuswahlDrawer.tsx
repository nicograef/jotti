import { useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router'

import {
  Drawer,
  DrawerBody,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
} from '@/components/ui/drawer'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { BackendSingleton } from '@/lib/Backend'
import { formatEuro } from '@/lib/utils'

import {
  AKTIVE_TISCHE_MIT_FAVORITEN_KEY,
  MEINE_TISCHE_STATE_KEY,
  useAktiveTischeMitFavoriten,
} from '../table/hooks'
import type { AktiverTischMitFavorit } from '../table/Tisch'
import { TischBackend } from '../table/TischBackend'

const tischBackend = new TischBackend(BackendSingleton)

// Reihenfolge im Alle-Tische-Drawer: durchgehend nach Tischname mit
// numerischem Vergleich („Tisch 2" vor „Tisch 10"). Favoriten und Saldo
// werden pro Zeile weiter angezeigt, aber nicht mehr zur Sortierung genutzt —
// so bleibt die Reihenfolge stabil und vorhersehbar. Reine Darstellungs-
// sortierung bereits vollständig geladener Daten.
function sortiereTische(
  a: AktiverTischMitFavorit,
  b: AktiverTischMitFavorit,
): number {
  return a.name.localeCompare(b.name, 'de', { numeric: true })
}

interface TischAuswahlDrawerProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function TischAuswahlDrawer({
  open,
  onOpenChange,
}: TischAuswahlDrawerProps) {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { tische } = useAktiveTischeMitFavoriten()
  const { loading: favoritLoading, run: runToggleFavorit } = useActionSubmit({
    actionLabel: 'Favorit ändern',
  })

  // Reine Durchblätter-/Favorisier-Liste — die Suche über alle Tische liegt
  // jetzt auf der Hauptseite (TableSelectionPage), kein zweites Suchfeld hier.
  const sortierteTische = [...tische].sort(sortiereTische)

  const toggleFavorit = async (tisch: AktiverTischMitFavorit) => {
    if (tisch.istFavorit) {
      await tischBackend.favoritEntfernen(tisch.id)
    } else {
      await tischBackend.favoritHinzufuegen(tisch.id)
    }
    await Promise.all([
      queryClient.invalidateQueries({
        queryKey: [AKTIVE_TISCHE_MIT_FAVORITEN_KEY],
      }),
      queryClient.invalidateQueries({ queryKey: [MEINE_TISCHE_STATE_KEY] }),
    ])
  }

  const handleTischClick = (tisch: AktiverTischMitFavorit) => {
    void navigate(`/service/tische/${tisch.id.toString()}`)
    onOpenChange(false)
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerContent aria-describedby={undefined}>
        <DrawerHeader>
          <DrawerTitle>Alle Tische</DrawerTitle>
        </DrawerHeader>
        <DrawerBody className="flex flex-col gap-0 px-4 pb-6">
          {sortierteTische.map((tisch) => (
            <div
              key={tisch.id}
              className="flex items-center border-b last:border-b-0"
            >
              <button
                type="button"
                className="flex size-11 shrink-0 items-center justify-center text-xl leading-none"
                disabled={favoritLoading}
                onClick={() => {
                  void runToggleFavorit(async () => {
                    await toggleFavorit(tisch)
                  })
                }}
                aria-label={
                  tisch.istFavorit
                    ? `${tisch.name} aus Favoriten entfernen`
                    : `${tisch.name} zu Favoriten hinzufügen`
                }
              >
                {tisch.istFavorit ? '★' : '☆'}
              </button>
              <button
                type="button"
                className="flex-1 flex items-center justify-between text-left hover:bg-accent/50 rounded px-2 py-3 transition-colors"
                onClick={() => {
                  handleTischClick(tisch)
                }}
              >
                <span className="font-medium">{tisch.name}</span>
                <span
                  className={
                    tisch.saldoCents < 0
                      ? 'text-sm text-destructive'
                      : 'text-sm text-muted-foreground'
                  }
                >
                  {formatEuro(tisch.saldoCents)}
                </span>
              </button>
            </div>
          ))}
        </DrawerBody>
      </DrawerContent>
    </Drawer>
  )
}
