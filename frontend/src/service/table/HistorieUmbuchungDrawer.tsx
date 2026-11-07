import { useMemo, useState } from 'react'

import { LadefehlerAlert } from '@/components/common/LadefehlerAlert'
import { Button } from '@/components/ui/button'
import {
  Drawer,
  DrawerBody,
  DrawerClose,
  DrawerContent,
  DrawerFooter,
} from '@/components/ui/drawer'
import { NativeSelect, NativeSelectOption } from '@/components/ui/native-select'
import { Spinner } from '@/components/ui/spinner'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { useMengen } from '@/hooks/use-mengen'
import { useOffenerVorgang } from '@/hooks/use-offener-vorgang'

import { ActionHint } from '../ActionHint'
import { KommentarField } from '../CommentField'
import {
  calculateTotalPrice,
  selectPositionen,
  toAuswahlPositionen,
  toMengen,
  toPositionRefs,
} from '../drawerUtils'
import { GesamtZeile } from '../GesamtZeile'
import { PositionAuswahlListe } from '../PositionAuswahlListe'
import { mitVersuchsSchluessel } from '../versuchsSchluessel'
import { AlleAuswaehlenButton } from './AlleAuswaehlenButton'
import type { Bestellung } from './Bestellung'
import { tischBackend, useAktiveTischeMitFavoriten } from './hooks'
import { QuelleDrawerHeader } from './QuelleDrawerHeader'
import type { Tisch } from './Tisch'
import type { Umbuchung } from './Umbuchung'

interface HistorieUmbuchungDrawerProps {
  tisch: Tisch
  // Ursprungsvorgang (Bestellung oder Umbuchungs-Zugang), dessen Positionen umgebucht
  // werden; beschriftet den Drawer.
  quelle: Bestellung | Umbuchung
  onClose: () => void
  onBestellungUmgebucht: (zielName: string) => void
}

export function HistorieUmbuchungDrawer({
  tisch,
  quelle,
  onClose,
  onBestellungUmgebucht,
}: HistorieUmbuchungDrawerProps) {
  const positionen = quelle.umbuchbarePositionen
  const [zielTischId, setZielTischId] = useState<number | null>(null)
  const [kommentar, setKommentar] = useState('')
  const {
    tische,
    isPending: tischeLoading,
    isError: tischeError,
    refetch: reloadTische,
  } = useAktiveTischeMitFavoriten()

  const umbuchbareMengen = useMemo(() => toMengen(positionen), [positionen])
  const auswahl = useMengen<string>(
    (positionId) => umbuchbareMengen[positionId] || 0,
  )

  // Die Positionsauswahl meldet bereits useMengen; der getippte Kommentar
  // kommt hinzu, weil er ein zweites Mal formuliert werden müsste.
  useOffenerVorgang(kommentar.trim() !== '')

  const zielTische = useMemo(
    () => tische.filter((candidate) => candidate.id !== tisch.id),
    [tisch.id, tische],
  )

  const { mengen, add, remove } = auswahl
  const selectedPositionen = selectPositionen(positionen, mengen)
  const totalPrice = calculateTotalPrice(selectedPositionen)
  const noPositionenSelected = selectedPositionen.length === 0
  const keinZielTischVerfuegbar = zielTische.length === 0
  // Grund am Button nur für die behebbaren Bedingungen: fehlt gänzlich ein
  // Ziel-Tisch, nennt ihn bereits der Select-Platzhalter bzw. der LadefehlerAlert.
  const disabledReason = noPositionenSelected
    ? 'Positionen auswählen'
    : keinZielTischVerfuegbar
      ? null
      : zielTischId === null
        ? 'Ziel-Tisch wählen'
        : null

  const { loading, run } = useActionSubmit({
    actionLabel: 'Umbuchung ausführen',
    onSuccess: () => {
      const zielName =
        zielTische.find((candidate) => candidate.id === zielTischId)?.name ?? ''
      onBestellungUmgebucht(zielName)
    },
  })

  const onSubmit = async () => {
    if (zielTischId === null) {
      return
    }

    await run(async () => {
      // Der Kommentar gehört nicht zum Schlüssel: wer ihn nach einer verlorenen
      // Antwort ändert, wiederholt dieselbe Umbuchung.
      const positionen = toPositionRefs(selectedPositionen)
      await mitVersuchsSchluessel(
        ['umbuchung', tisch.id, zielTischId, positionen],
        (umbuchungId) =>
          tischBackend.bestellungUmbuchen({
            umbuchungId,
            quellTischId: tisch.id,
            zielTischId,
            positionen,
            benutzerKommentar: kommentar,
          }),
      )
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
        <QuelleDrawerHeader quelle={quelle}>
          Positionen auswählen und auf einen Ziel-Tisch umbuchen.
        </QuelleDrawerHeader>
        <DrawerBody className="mx-auto w-full max-w-sm">
          {positionen.length > 0 && (
            <div className="px-4 pb-2">
              <AlleAuswaehlenButton
                positionen={positionen}
                steuerung={auswahl}
              />
            </div>
          )}
          <PositionAuswahlListe
            positionen={toAuswahlPositionen(positionen)}
            mengen={mengen}
            onAdd={add}
            onRemove={remove}
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
            <GesamtZeile label="Umbuchung gesamt" betrag={totalPrice} />
          )}
          <div className="space-y-1">
            <p className="text-sm font-medium">Ziel-Tisch</p>
            {/* Expliziter Fehlerzustand statt der leeren Auswahl — sonst läse
                sich der Ladefehler als „kein Ziel-Tisch vorhanden". */}
            {tischeError ? (
              <LadefehlerAlert
                titel="Ziel-Tische konnten nicht geladen werden"
                onErneutVersuchen={() => {
                  void reloadTische()
                }}
              />
            ) : (
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
            )}
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
