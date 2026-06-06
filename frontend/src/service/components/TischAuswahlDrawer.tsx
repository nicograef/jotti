import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useNavigate } from 'react-router'
import { toast } from 'sonner'

import {
  Drawer,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
} from '@/components/ui/drawer'
import { Input } from '@/components/ui/input'
import { BackendSingleton } from '@/lib/Backend'
import { getActionErrorMessage } from '@/lib/errorMessages'
import { formatCents } from '@/lib/utils'

import {
  AKTIVE_TISCHE_MIT_FAVORITEN_KEY,
  useAktiveTischeMitFavoriten,
} from '../table/hooks'
import type { AktiverTischMitFavorit } from '../table/Tisch'
import { TischBackend } from '../table/TischBackend'

const tischBackend = new TischBackend(BackendSingleton)

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

  const gefilterteTische = tische.filter((t) =>
    t.name.toLowerCase().includes(suche.toLowerCase()),
  )

  const favoritMutation = useMutation({
    mutationFn: (tisch: AktiverTischMitFavorit) =>
      tisch.istFavorit
        ? tischBackend.favoritEntfernen(tisch.id)
        : tischBackend.favoritHinzufuegen(tisch.id),
    onSuccess: () =>
      queryClient.invalidateQueries({
        queryKey: [AKTIVE_TISCHE_MIT_FAVORITEN_KEY],
      }),
    onError: (error) =>
      toast.error(
        getActionErrorMessage({ actionLabel: 'Favorit ändern', error }),
      ),
  })

  const handleTischClick = (tisch: AktiverTischMitFavorit) => {
    void navigate(`/service/tische/${tisch.id.toString()}`)
    onOpenChange(false)
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange} direction="bottom">
      <DrawerContent className="px-4 pb-6">
        <DrawerHeader>
          <DrawerTitle>Alle Tische</DrawerTitle>
        </DrawerHeader>
        <div className="mb-3">
          <Input
            placeholder="Tisch suchen..."
            value={suche}
            onChange={(e) => {
              setSuche(e.target.value)
            }}
          />
        </div>
        <div className="flex flex-col gap-0 overflow-y-auto max-h-[60vh]">
          {gefilterteTische.map((tisch) => (
            <div
              key={tisch.id}
              className="flex items-center gap-3 py-3 border-b last:border-b-0"
            >
              <button
                type="button"
                className="text-xl leading-none shrink-0 w-7 text-center"
                disabled={favoritMutation.isPending}
                onClick={() => {
                  favoritMutation.mutate(tisch)
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
                className="flex-1 flex items-center justify-between text-left hover:bg-accent/50 rounded px-2 py-1 transition-colors"
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
        </div>
      </DrawerContent>
    </Drawer>
  )
}
