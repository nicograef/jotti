import { useState } from 'react'
import { useNavigate } from 'react-router'

import {
  Drawer,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
} from '@/components/ui/drawer'
import { Input } from '@/components/ui/input'
import { BackendSingleton } from '@/lib/Backend'
import { formatCents } from '@/lib/utils'

import { useAktiveTischeMitFavoriten } from '../table/hooks'
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
  const [suche, setSuche] = useState('')
  const { tische, setData } = useAktiveTischeMitFavoriten()

  const gefilterteTische = tische.filter((t) =>
    t.name.toLowerCase().includes(suche.toLowerCase()),
  )

  const handleFavoritToggle = async (tisch: AktiverTischMitFavorit) => {
    // Optimistic update
    setData((prev: AktiverTischMitFavorit[]) =>
      prev.map((t: AktiverTischMitFavorit) =>
        t.id === tisch.id ? { ...t, istFavorit: !t.istFavorit } : t,
      ),
    )

    try {
      if (tisch.istFavorit) {
        await tischBackend.favoritEntfernen(tisch.id)
      } else {
        await tischBackend.favoritHinzufuegen(tisch.id)
      }
    } catch {
      // Revert on error
      setData((prev: AktiverTischMitFavorit[]) =>
        prev.map((t: AktiverTischMitFavorit) =>
          t.id === tisch.id ? { ...t, istFavorit: tisch.istFavorit } : t,
        ),
      )
    }
  }

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
                onClick={() => void handleFavoritToggle(tisch)}
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
