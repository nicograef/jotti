import { CircleCheck } from 'lucide-react'
import { useMemo, useState } from 'react'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import {
  Drawer,
  DrawerBody,
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
import { useMengen } from '@/hooks/use-mengen'
import {
  formatAlleAuswaehlenLabel,
  formatEuro,
  formatRelativeTime,
} from '@/lib/utils'

import type { Bestellung, Position } from '../../table/Bestellung'
import { useAktiveTische } from '../../table/hooks'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import type { Umbuchung } from '../../table/Umbuchung'
import { PositionAuswahlListe } from '../PositionAuswahlListe'
import { ActionHint } from './ActionHint'
import { KommentarField } from './CommentField'
import {
  calculateTotalPrice,
  quelleTitel,
  quelleZeitpunkt,
  selectPositionen,
  toAuswahlPositionen,
  toPositionRefs,
} from './drawerUtils'

interface HistorieUmbuchungDrawerProps {
  backend: Pick<TischBackend, 'bestellungUmbuchen'>
  tisch: Tisch
  // Ursprungsvorgang (Bestellung oder Umbuchungs-Zugang), dessen Positionen umgebucht
  // werden; beschriftet den Drawer.
  quelle: Bestellung | Umbuchung
  onClose: () => void
  onBestellungUmgebucht: () => void
}

// Volle umbuchbare Menge je Position — Basis für den „Alle auswählen"-Button.
function createDefaultMengen(positionen: Position[]): Record<string, number> {
  return positionen.reduce<Record<string, number>>((acc, position) => {
    acc[position.positionId] = position.menge
    return acc
  }, {})
}

export function HistorieUmbuchungDrawer({
  backend,
  tisch,
  quelle,
  onClose,
  onBestellungUmgebucht,
}: HistorieUmbuchungDrawerProps) {
  const positionen = quelle.umbuchbarePositionen
  const [zielTischId, setZielTischId] = useState<number | null>(null)
  const [kommentar, setKommentar] = useState('')
  const { tische, isPending: tischeLoading } = useAktiveTische()

  const umbuchbareMengen = useMemo(
    () => createDefaultMengen(positionen),
    [positionen],
  )
  const {
    mengen,
    add: onAdd,
    remove: onRemove,
    reset,
    setAll,
  } = useMengen<string>((positionId) => umbuchbareMengen[positionId] || 0)

  const zielTische = useMemo(
    () => tische.filter((candidate) => candidate.id !== tisch.id),
    [tisch.id, tische],
  )

  const selectedPositionen = selectPositionen(positionen, mengen)
  const totalPrice = calculateTotalPrice(selectedPositionen)
  const noPositionenSelected = selectedPositionen.length === 0
  const keinZielTischVerfuegbar = zielTische.length === 0
  // Grund am Button nur für die behebbaren Bedingungen: Fehlt gänzlich ein
  // Ziel-Tisch, nennt bereits der Select-Platzhalter den Grund — ein zweiter
  // Hinweis wäre redundant (gleiche Dedup wie in HistorieStornierungDrawer).
  const disabledReason = noPositionenSelected
    ? 'Positionen auswählen'
    : keinZielTischVerfuegbar
      ? null
      : zielTischId === null
        ? 'Ziel-Tisch wählen'
        : null

  const alleVollAusgewaehlt =
    positionen.length > 0 &&
    positionen.every(
      (position) => (mengen[position.positionId] || 0) === position.menge,
    )
  const umbuchbarGesamt = calculateTotalPrice(positionen)

  const alleAuswaehlen = () => {
    if (alleVollAusgewaehlt) {
      reset()
      return
    }
    setAll(umbuchbareMengen)
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
        benutzerKommentar: kommentar,
      })
    })
  }

  return (
    <Drawer
      open
      onOpenChange={(isOpen) => {
        if (!isOpen) onClose()
      }}
    >
      <DrawerContent pending={loading}>
        <DrawerHeader className="mx-auto w-full max-w-sm">
          <DrawerTitle>
            {quelleTitel(quelle)} ·{' '}
            {formatRelativeTime(quelleZeitpunkt(quelle))} · {quelle.userName}
          </DrawerTitle>
          <DrawerDescription>
            Positionen auswählen und auf einen Ziel-Tisch umbuchen.
          </DrawerDescription>
        </DrawerHeader>
        <DrawerBody className="mx-auto w-full max-w-sm">
          {positionen.length > 0 && (
            <div className="px-4 pb-2">
              <button
                type="button"
                onClick={alleAuswaehlen}
                className="flex h-11 w-full items-center justify-center gap-2 rounded-lg border border-primary/50 bg-primary/5 px-4 text-sm font-medium text-primary"
              >
                <CircleCheck className="size-5" />
                {alleVollAusgewaehlt
                  ? 'Auswahl aufheben'
                  : formatAlleAuswaehlenLabel(
                      positionen.length,
                      umbuchbarGesamt,
                    )}
              </button>
            </div>
          )}
          <PositionAuswahlListe
            positionen={toAuswahlPositionen(positionen)}
            mengen={mengen}
            onAdd={onAdd}
            onRemove={onRemove}
          />
          <div className="px-4">
            <KommentarField
              onChange={(value) => {
                setKommentar(value)
              }}
            />
          </div>
        </DrawerBody>
        <DrawerFooter className="mx-auto w-full max-w-sm">
          {!noPositionenSelected && (
            <div className="flex justify-between border-t-2 pt-2 font-bold">
              <div>Umbuchung gesamt</div>
              <div>{formatEuro(totalPrice)}</div>
            </div>
          )}
          <div className="space-y-1">
            <p className="text-sm font-medium">Ziel-Tisch</p>
            <NativeSelect
              className="w-full"
              value={zielTischId === null ? '' : String(zielTischId)}
              onChange={(event) => {
                setZielTischId(Number(event.target.value))
              }}
              disabled={loading || tischeLoading || keinZielTischVerfuegbar}
            >
              <NativeSelectOption value="" disabled>
                {keinZielTischVerfuegbar
                  ? 'Kein aktiver Ziel-Tisch verfügbar'
                  : 'Ziel-Tisch wählen…'}
              </NativeSelectOption>
              {zielTische.map((candidate) => (
                <NativeSelectOption
                  key={candidate.id}
                  value={String(candidate.id)}
                >
                  {candidate.name}
                </NativeSelectOption>
              ))}
            </NativeSelect>
          </div>
          <ActionHint reason={disabledReason} />
          <Button
            disabled={loading || noPositionenSelected || zielTischId === null}
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
      </DrawerContent>
    </Drawer>
  )
}
