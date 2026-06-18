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
import { useMengen } from '@/hooks/use-mengen'
import { formatCents, formatPositionName } from '@/lib/utils'

import type { Bestellung, Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { KommentarField } from './CommentField'
import {
  calculateTotalPrice,
  selectPositionen,
  toPositionRefs,
} from './drawerUtils'

interface HistorieStornierungDrawerProps {
  backend: Pick<TischBackend, 'stornierungErteilen'>
  tisch: Tisch
  bestellung: Bestellung
  positionen: Position[]
  onClose: () => void
  onStornierungErteilt: () => void
}

export function HistorieStornierungDrawer({
  backend,
  tisch,
  bestellung,
  positionen,
  onClose,
  onStornierungErteilt,
}: HistorieStornierungDrawerProps) {
  const [kommentar, setKommentar] = useState('')
  const { mengen, add, remove } = useMengen<string>(
    (positionId) =>
      positionen.find((p) => p.positionId === positionId)?.menge ?? 0,
  )

  const selectedPositionen = selectPositionen(positionen, mengen)
  const totalPrice = calculateTotalPrice(selectedPositionen)
  const noPositionenSelected = selectedPositionen.length === 0
  const kommentarInvalid = kommentar.trim().length < 3

  const { loading, run } = useActionSubmit({
    actionLabel: 'Stornierung ausführen',
    byCode: {
      position_nicht_stornierbar:
        'Mindestens eine Position ist nicht mehr stornierbar. Bitte Auswahl aktualisieren.',
    },
    onSuccess: () => {
      onStornierungErteilt()
    },
  })

  const onSubmit = async () => {
    await run(async () => {
      await backend.stornierungErteilen({
        tischId: tisch.id,
        positionen: toPositionRefs(selectedPositionen),
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
            <DrawerTitle>
              Stornierung aus Bestellung {bestellung.id.slice(0, 8)}
            </DrawerTitle>
            <DrawerDescription>
              Positionen aus dieser Bestellung zum Stornieren auswählen.
            </DrawerDescription>
          </DrawerHeader>
          <ScrollArea className="max-h-72">
            <div className="px-4 space-y-2">
              {positionen.map((position: Position) => {
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
                        {position.menge}
                        &nbsp;Stück
                      </div>
                    </div>
                    <div className="flex items-center gap-1 ml-2">
                      <Button
                        variant="secondary"
                        size="icon"
                        className="h-8 w-8"
                        onClick={() => {
                          remove(position.positionId)
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
                        onClick={() => {
                          add(position.positionId)
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
