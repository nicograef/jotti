import { useEffect, useRef, useState } from 'react'

import { EuroInput } from '@/components/common/EuroInput'
import { Button } from '@/components/ui/button'
import {
  DrawerBody,
  DrawerClose,
  DrawerContent,
  DrawerFooter,
} from '@/components/ui/drawer'
import { Label } from '@/components/ui/label'
import { Spinner } from '@/components/ui/spinner'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { formatCents, parseCents } from '@/lib/utils'

import type { Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { AbschlussHeader } from './AbschlussHeader'
import { AbschlussLeer } from './AbschlussLeer'
import { KommentarField } from './CommentField'
import {
  calculateZahlungsbetraege,
  toPositionRefs,
  toReceiptItems,
} from './drawerUtils'
import { Receipt } from './Receipt'
import { RestbetragZeile } from './RestbetragZeile'

interface ZahlungAbschlussProps {
  backend: Pick<TischBackend, 'zahlungKassieren'>
  tisch: Tisch
  // Die ausgewählten, zu kassierenden Positionen (mit Auswahl-Menge).
  positionenToPay: Position[]
  totalCents: number
  restNachZahlungCents: number
  zahlungKassiert: () => void
  // 'sheet' rendert den Bottom-Sheet-Drawer-Inhalt (Handy), 'spalte' die feste
  // Abschluss-Spalte (ab lg). Einzige Quelle des Abschluss-Inhalts; die beiden
  // Varianten unterscheiden sich nur im umschließenden Container.
  variant: 'sheet' | 'spalte'
}

// Presentation-neutraler Abschluss-Inhalt des Tisch-Kassierens (Beleg, Erhalten,
// Zielbetrag inkl. Trinkgeld, Rückgeld, Trinkgeld-Hinweis, Kommentar,
// „Kassieren"). Trägt den vollständigen Eingabe-State und das Submit-/Fehler-/
// Retry-Verhalten und wird sowohl im Handy-Drawer als auch in der festen Spalte
// gerendert. Kein Client-Idempotenz-Schlüssel: die Idempotenz ist zustandsbasiert
// (bereits bezahlte Positionen → position_nicht_bezahlbar), und der
// Loading-Guard verhindert den Doppel-Submit.
export function ZahlungAbschluss(props: ZahlungAbschlussProps) {
  const [kommentar, setKommentar] = useState('')
  const [erhaltenEuro, setErhaltenEuro] = useState('')
  const [zielbetragEuro, setZielbetragEuro] = useState('')

  const noPositionenSelected = props.positionenToPay.length === 0

  // In der dauerhaften Spalte überlebt der Eingabe-State sonst über einen
  // Auswahl-Reset hinweg. Beim Beginn einer neuen Zusammenstellung (aus dem
  // Leerzustand) starten die Eingaben deshalb leer, damit nichts aus einer
  // abgebrochenen Zahlung übertragen wird.
  const warLeerRef = useRef(noPositionenSelected)
  useEffect(() => {
    if (warLeerRef.current && !noPositionenSelected) {
      setErhaltenEuro('')
      setZielbetragEuro('')
      setKommentar('')
    }
    warLeerRef.current = noPositionenSelected
  }, [noPositionenSelected])

  const { rueckgeldCents, trinkgeldCents } = calculateZahlungsbetraege(
    props.totalCents,
    parseCents(erhaltenEuro),
    parseCents(zielbetragEuro),
  )

  const { loading, run } = useActionSubmit({
    actionLabel: 'Zahlung kassieren',
    byCode: {
      position_nicht_bezahlbar:
        'Mindestens eine Position ist nicht mehr bezahlbar. Bitte Auswahl aktualisieren.',
    },
    onSuccess: () => {
      setErhaltenEuro('')
      setZielbetragEuro('')
      setKommentar('')
      props.zahlungKassiert()
    },
  })

  const onSubmit = async () => {
    await run(async () => {
      await props.backend.zahlungKassieren({
        tischId: props.tisch.id,
        positionen: toPositionRefs(props.positionenToPay),
        kommentar,
      })
    })
  }

  const inhalt = (
    <>
      <AbschlussHeader
        variant={props.variant}
        eyebrow="Zahlung für"
        title={props.tisch.name}
        description={`Zahlung für ${props.tisch.name}`}
      />
      <DrawerBody className="mx-auto w-full max-w-sm">
        {noPositionenSelected ? (
          <AbschlussLeer>Positionen auswählen, um zu kassieren.</AbschlussLeer>
        ) : (
          <>
            <Receipt
              positionen={toReceiptItems(props.positionenToPay)}
              totalPrice={props.totalCents}
            />
            <div className="px-4 pt-3 flex flex-col gap-2">
              <div className="flex items-center justify-between gap-3">
                <Label htmlFor="erhalten">Erhalten</Label>
                <EuroInput
                  id="erhalten"
                  value={erhaltenEuro}
                  onValueChange={setErhaltenEuro}
                  className="w-28"
                />
              </div>
              <div className="flex items-center justify-between gap-3">
                <Label htmlFor="zielbetrag">Zahlbetrag inkl. Trinkgeld</Label>
                <EuroInput
                  id="zielbetrag"
                  value={zielbetragEuro}
                  onValueChange={setZielbetragEuro}
                  aria-describedby="zielbetrag-hinweis"
                  className="w-28"
                />
              </div>
              <p
                id="zielbetrag-hinweis"
                className="text-xs text-muted-foreground"
              >
                Nur ausfüllen, wenn der Gast aufrundet: den vollen Betrag
                inklusive Trinkgeld eintragen, dann rechnet die Kasse das
                Rückgeld passend.
              </p>
              {rueckgeldCents !== null && (
                <div className="flex items-baseline justify-between pt-1">
                  <div className="text-[15px] font-semibold">Rückgeld</div>
                  <div className="text-xl font-bold tabular-nums">
                    {formatCents(rueckgeldCents)}&nbsp;€
                  </div>
                </div>
              )}
              {trinkgeldCents !== null && (
                <>
                  <div className="flex justify-between font-medium">
                    <div>Trinkgeld</div>
                    <div className="tabular-nums">
                      {formatCents(trinkgeldCents)}&nbsp;€
                    </div>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Trinkgeld wird nicht als Kasseneinnahme gebucht und gehört
                    nicht in die Kassenlade.
                  </p>
                </>
              )}
            </div>
            <div className="px-4 pt-3">
              <KommentarField
                value={kommentar}
                onChange={(value) => {
                  setKommentar(value)
                }}
              />
            </div>
          </>
        )}
      </DrawerBody>
      <DrawerFooter className="mx-auto w-full max-w-sm">
        {props.variant === 'spalte' && (
          <RestbetragZeile cents={props.restNachZahlungCents} />
        )}
        <Button
          disabled={loading || noPositionenSelected}
          onClick={() => {
            void onSubmit()
          }}
        >
          {loading ? <Spinner /> : null} Kassieren
        </Button>
        {props.variant === 'sheet' && (
          <DrawerClose asChild>
            <Button variant="outline" disabled={loading}>
              Abbrechen
            </Button>
          </DrawerClose>
        )}
      </DrawerFooter>
    </>
  )

  if (props.variant === 'sheet') {
    return <DrawerContent pending={loading}>{inhalt}</DrawerContent>
  }

  // Feste Spalte: dieselben Body/Footer-Primitive wie im Sheet, nur in einem
  // eigenen, unabhängig scrollenden Container. group/drawer-content +
  // data-pending übernehmen das Body-Dimming des Drawers während des Submits.
  return (
    <aside
      data-pending={loading || undefined}
      className="group/drawer-content flex min-h-0 flex-col overflow-hidden rounded-xl border bg-popover text-sm text-popover-foreground"
    >
      {inhalt}
    </aside>
  )
}
