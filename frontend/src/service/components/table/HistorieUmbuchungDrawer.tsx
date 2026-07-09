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
import { Spinner } from '@/components/ui/spinner'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { formatCents } from '@/lib/utils'

import type { Position } from '../../table/Bestellung'
import { useAktiveTische } from '../../table/hooks'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { PositionAuswahlListe } from '../PositionAuswahlListe'
import {
  calculateTotalPrice,
  selectPositionen,
  toAuswahlPositionen,
  toPositionRefs,
} from './drawerUtils'

interface HistorieUmbuchungDrawerProps {
  backend: Pick<TischBackend, 'bestellungUmbuchen'>
  tisch: Tisch
  // Ursprungsvorgang (Bestellung oder Umbuchungs-Zugang), dessen Positionen umgebucht
  // werden; beschriftet den Drawer.
  vorgangId: string
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
  vorgangId,
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

  const onAdd = (positionId: string) => {
    const maxMenge =
      positionen.find((p) => p.positionId === positionId)?.menge ?? 0
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
              Umbuchung aus Vorgang {vorgangId.slice(0, 8)}
            </DrawerTitle>
            <DrawerDescription>
              Positionen auswählen und auf einen Ziel-Tisch umbuchen.
            </DrawerDescription>
          </DrawerHeader>
          <PositionAuswahlListe
            positionen={toAuswahlPositionen(positionen)}
            mengen={mengen}
            onAdd={onAdd}
            onRemove={onRemove}
          />
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
