import { Minus, Plus } from 'lucide-react'
import { useMemo, useState } from 'react'
import { toast } from 'sonner'

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
import { NativeSelect, NativeSelectOption } from '@/components/ui/native-select'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Spinner } from '@/components/ui/spinner'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { formatCents, formatPositionName } from '@/lib/utils'

import type { Bestellung, Position } from '../../table/Bestellung'
import { useAktiveTische } from '../../table/hooks'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import {
  calculateTotalPrice,
  selectPositionen,
  toPositionRefs,
} from './drawerUtils'

interface HistorieUmbuchungDrawerProps {
  backend: Pick<TischBackend, 'bestellungUmbuchen'>
  tisch: Tisch
  bestellung: Bestellung
  positionen: Position[]
  onClose: () => void
  onBestellungUmgebucht: () => void
}

function createDefaultMengen(positionen: Position[]): Record<string, number> {
  return positionen.reduce<Record<string, number>>((acc, position) => {
    acc[position.positionId] = position.menge
    return acc
  }, {})
}

export function HistorieUmbuchungDrawer({
  backend,
  tisch,
  bestellung,
  positionen,
  onClose,
  onBestellungUmgebucht,
}: HistorieUmbuchungDrawerProps) {
  const [mengen, setMengen] = useState<Record<string, number>>(() =>
    createDefaultMengen(positionen),
  )
  const [zielTischIdOverride, setZielTischIdOverride] = useState<number | null>(
    null,
  )
  const { tische, isPending: tischeLoading } = useAktiveTische()

  const zielTische = useMemo(
    () => tische.filter((candidate) => candidate.id !== tisch.id),
    [tisch.id, tische],
  )
  const zielTischId: number | null = useMemo(() => {
    if (
      zielTischIdOverride !== null &&
      zielTische.some((candidate) => candidate.id === zielTischIdOverride)
    ) {
      return zielTischIdOverride
    }

    return zielTische.length > 0 ? zielTische[0].id : null
  }, [zielTischIdOverride, zielTische])

  const selectedPositionen = selectPositionen(positionen, mengen)
  const totalPrice = calculateTotalPrice(selectedPositionen)
  const noPositionenSelected = selectedPositionen.length === 0

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
    actionLabel: 'Umbuchung ausführen',
    byCode: {
      position_nicht_umbuchbar:
        'Mindestens eine Position ist nicht mehr umbuchbar. Bitte Auswahl aktualisieren.',
    },
    onSuccess: () => {
      toast.success('Bestellung umgebucht.')
      onBestellungUmgebucht()
    },
  })

  const onSubmit = async () => {
    if (zielTischId === null) {
      return
    }

    await run(async () => {
      await backend.bestellungUmbuchen({
        quellTischId: tisch.id,
        zielTischId,
        positionen: toPositionRefs(selectedPositionen),
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
              Umbuchung aus Bestellung {bestellung.id.slice(0, 8)}
            </DrawerTitle>
            <DrawerDescription>
              Positionen auswählen und auf einen Ziel-Tisch umbuchen.
            </DrawerDescription>
          </DrawerHeader>
          <ScrollArea className="max-h-72">
            <div className="px-4 space-y-2">
              {positionen.map((position) => {
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
              <div>Umbuchung gesamt</div>
              <div>{formatCents(totalPrice)}&nbsp;€</div>
            </div>
          )}
          <div className="px-4 pb-2 space-y-1">
            <p className="text-sm font-medium">Ziel-Tisch</p>
            <NativeSelect
              className="w-full"
              value={zielTischId === null ? '' : String(zielTischId)}
              onChange={(event) => {
                setZielTischIdOverride(Number(event.target.value))
              }}
              disabled={loading || tischeLoading || zielTische.length === 0}
            >
              {zielTische.length === 0 ? (
                <NativeSelectOption value="">
                  Kein aktiver Ziel-Tisch verfügbar
                </NativeSelectOption>
              ) : (
                zielTische.map((candidate) => (
                  <NativeSelectOption
                    key={candidate.id}
                    value={String(candidate.id)}
                  >
                    {candidate.name}
                  </NativeSelectOption>
                ))
              )}
            </NativeSelect>
          </div>
          <DrawerFooter>
            <Button
              variant="secondary"
              disabled={
                loading ||
                noPositionenSelected ||
                zielTischId === null ||
                zielTische.length === 0
              }
              onClick={() => {
                void onSubmit()
              }}
            >
              {loading ? <Spinner /> : null} Umbuchung ausführen
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
