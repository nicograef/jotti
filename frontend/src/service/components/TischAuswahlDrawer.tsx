import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useNavigate } from 'react-router'

import {
  Drawer,
  DrawerBody,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
} from '@/components/ui/drawer'
import { Input } from '@/components/ui/input'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { BackendSingleton } from '@/lib/Backend'
import { formatCents } from '@/lib/utils'

import {
  AKTIVE_TISCHE_MIT_FAVORITEN_KEY,
  MEINE_TISCHE_STATE_KEY,
  useAktiveTischeMitFavoriten,
} from '../table/hooks'
import type { AktiverTischMitFavorit } from '../table/Tisch'
import { TischBackend } from '../table/TischBackend'

const tischBackend = new TischBackend(BackendSingleton)

// Reihenfolge im Alle-Tische-Drawer: Favoriten zuerst, dann nach offenem Saldo
// (absteigend), zuletzt nach Name. Reine Darstellungssortierung bereits
// vollständig geladener Daten.
function sortiereTische(
  a: AktiverTischMitFavorit,
  b: AktiverTischMitFavorit,
): number {
  if (a.istFavorit !== b.istFavorit) return a.istFavorit ? -1 : 1
  if (a.saldoCents !== b.saldoCents) return b.saldoCents - a.saldoCents
  return a.name.localeCompare(b.name, 'de')
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
  const [suche, setSuche] = useState('')
  const { tische } = useAktiveTischeMitFavoriten()
  const { loading: favoritLoading, run: runToggleFavorit } = useActionSubmit({
    actionLabel: 'Favorit ändern',
  })

  const gefilterteTische = tische
    .filter((t) => t.name.toLowerCase().includes(suche.toLowerCase()))
    .sort(sortiereTische)

  const favoritMutation = useMutation({
    mutationFn: (tisch: AktiverTischMitFavorit) =>
      tisch.istFavorit
        ? tischBackend.favoritEntfernen(tisch.id)
        : tischBackend.favoritHinzufuegen(tisch.id),
    onSuccess: () =>
      Promise.all([
        queryClient.invalidateQueries({
          queryKey: [AKTIVE_TISCHE_MIT_FAVORITEN_KEY],
        }),
        queryClient.invalidateQueries({ queryKey: [MEINE_TISCHE_STATE_KEY] }),
      ]),
  })

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
        {/* Suchfeld außerhalb von DrawerBody, damit es beim Scrollen der Liste sichtbar bleibt. */}
        <div className="px-4 pb-3">
          <Input
            placeholder="Tisch suchen..."
            value={suche}
            onChange={(e) => {
              setSuche(e.target.value)
            }}
          />
        </div>
        <DrawerBody className="flex flex-col gap-0 px-4 pb-6">
          {gefilterteTische.map((tisch) => (
            <div
              key={tisch.id}
              className="flex items-center border-b last:border-b-0"
            >
              <button
                type="button"
                className="flex size-11 shrink-0 items-center justify-center text-xl leading-none"
                disabled={favoritMutation.isPending || favoritLoading}
                onClick={() => {
                  void runToggleFavorit(async () => {
                    await favoritMutation.mutateAsync(tisch)
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
                  {formatCents(tisch.saldoCents)} €
                </span>
              </button>
            </div>
          ))}
        </DrawerBody>
      </DrawerContent>
    </Drawer>
  )
}
