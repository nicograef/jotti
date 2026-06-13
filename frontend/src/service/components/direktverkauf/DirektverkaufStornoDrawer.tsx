import { Minus, Plus } from 'lucide-react'
import { useState } from 'react'

import { Button } from '@/components/ui/button'
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
} from '@/components/ui/drawer'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Spinner } from '@/components/ui/spinner'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { formatCents, formatPositionName } from '@/lib/utils'

import type {
  DirektverkaufHistorieEintrag,
  VerkaufPosition,
} from '../../direktverkauf/Direktverkauf'
import type { DirektverkaufBackend } from '../../direktverkauf/DirektverkaufBackend'
import { KommentarField } from '../table/CommentField'

interface DirektverkaufStornoDrawerProps {
  backend: Pick<DirektverkaufBackend, 'direktverkaufStornieren'>
  verkauf: DirektverkaufHistorieEintrag
  onClose: () => void
  onStorniert: () => void
}

export function DirektverkaufStornoDrawer({
  backend,
  verkauf,
  onClose,
  onStorniert,
}: DirektverkaufStornoDrawerProps) {
  const [kommentar, setKommentar] = useState('')
  const [mengen, setMengen] = useState<Record<string, number>>({})

  const selectedPositionen = verkauf.offenePositionen
    .map((position) => ({
      ...position,
      menge: mengen[position.positionId] || 0,
    }))
    .filter((position) => position.menge > 0)
  const totalPrice = selectedPositionen.reduce(
    (sum, position) => sum + position.einzelpreis * position.menge,
    0,
  )
  const noPositionenSelected = selectedPositionen.length === 0
  const kommentarInvalid = kommentar.trim().length < 3

  const onAdd = (positionId: string, maxMenge: number) => {
    setMengen((prev) => {
      const current = prev[positionId] || 0
      if (current >= maxMenge) return prev
      return { ...prev, [positionId]: current + 1 }
    })
  }

  const onRemove = (positionId: string) => {
    setMengen((prev) => {
      const current = prev[positionId] || 0
      if (current <= 0) return prev
      return { ...prev, [positionId]: current - 1 }
    })
  }

  const { loading, run } = useActionSubmit({
    actionLabel: 'Stornierung ausführen',
    byCode: {
      position_nicht_stornierbar:
        'Mindestens eine Position ist nicht mehr stornierbar. Bitte Auswahl aktualisieren.',
      kasse_nicht_geoeffnet:
        'Es ist keine Kassensitzung geöffnet. Bitte zuerst die Kasse öffnen.',
      verkauf_not_found: 'Der Verkauf wurde nicht gefunden.',
    },
    onSuccess: () => {
      onStorniert()
    },
  })

  const onSubmit = async () => {
    await run(async () => {
      await backend.direktverkaufStornieren({
        verkaufId: verkauf.verkaufId,
        positionen: selectedPositionen.map((position) => ({
          positionId: position.positionId,
          menge: position.menge,
        })),
        kommentar,
      })
    })
  }

  return (
    <Drawer
      open={true}
      onOpenChange={(isOpen) => {
        if (!isOpen) onClose()
      }}
    >
      <DrawerContent>
        <div className="mx-auto w-full max-w-sm">
          <DrawerHeader>
            <DrawerTitle>Verkauf stornieren</DrawerTitle>
            <DrawerDescription>
              Positionen aus diesem Verkauf zum Stornieren auswählen.
            </DrawerDescription>
          </DrawerHeader>
          <ScrollArea className="max-h-72">
            <div className="px-4 space-y-2">
              {verkauf.offenePositionen.map((position: VerkaufPosition) => {
                const selected = mengen[position.positionId] || 0
                return (
                  <div
                    key={position.positionId}
                    className="flex items-center justify-between border-b pb-2 last:border-0"
                  >
                    <div className="flex-1 min-w-0">
                      <div className="text-sm font-medium truncate">
                        {formatPositionName(
                          position.produktName,
                          position.varianteName,
                        )}
                      </div>
                      <div className="text-xs text-muted-foreground">
                        {formatCents(position.einzelpreis)}&nbsp;€ ·{' '}
                        {position.menge}&nbsp;Stück
                      </div>
                    </div>
                    <div className="flex items-center gap-1 ml-2">
                      <Button
                        variant="secondary"
                        size="icon"
                        className="h-8 w-8"
                        aria-label={`${formatPositionName(position.produktName, position.varianteName)} verringern`}
                        onClick={() => {
                          onRemove(position.positionId)
                        }}
                      >
                        <Minus
                          className={selected > 0 ? '' : 'opacity-50'}
                          size={16}
                        />
                      </Button>
                      <span className="font-bold tabular-nums text-center w-6 text-sm">
                        {selected}
                      </span>
                      <Button
                        variant="secondary"
                        size="icon"
                        className="h-8 w-8"
                        aria-label={`${formatPositionName(position.produktName, position.varianteName)} hinzufügen`}
                        onClick={() => {
                          onAdd(position.positionId, position.menge)
                        }}
                      >
                        <Plus size={16} />
                      </Button>
                    </div>
                  </div>
                )
              })}
            </div>
          </ScrollArea>
          {!noPositionenSelected && (
            <div className="flex justify-between font-bold px-4 pt-2 pb-2 border-t-2">
              <div>Stornierung gesamt</div>
              <div>{formatCents(totalPrice)}&nbsp;€</div>
            </div>
          )}
          <div className="px-4">
            <KommentarField
              required
              invalid={kommentarInvalid}
              onChange={(value) => {
                setKommentar(value)
              }}
            />
          </div>
          <DrawerFooter>
            <Button
              variant="destructive"
              disabled={loading || noPositionenSelected || kommentarInvalid}
              onClick={() => {
                void onSubmit()
              }}
            >
              {loading ? <Spinner /> : null} Stornierung erteilen
            </Button>
            <DrawerClose asChild>
              <Button variant="outline" disabled={loading}>
                Abbrechen
              </Button>
            </DrawerClose>
          </DrawerFooter>
        </div>
      </DrawerContent>
    </Drawer>
  )
}
