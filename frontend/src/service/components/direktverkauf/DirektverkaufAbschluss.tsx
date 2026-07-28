import { useState } from 'react'

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
import { useVorgangId } from '@/hooks/use-vorgang-id'
import { formatEuro, parseCents } from '@/lib/utils'

import type { VerkaufPositionInput } from '../../direktverkauf/Direktverkauf'
import type { DirektverkaufBackend } from '../../direktverkauf/DirektverkaufBackend'
import { AbschlussHeader } from '../table/AbschlussHeader'
import { AbschlussLeer } from '../table/AbschlussLeer'
import { AufrundenChips } from '../table/AufrundenChips'
import { KommentarField } from '../table/CommentField'
import { calculateZahlungsbetraege } from '../table/drawerUtils'
import type { ReceiptPosition } from '../table/Receipt'
import { Receipt } from '../table/Receipt'

interface DirektverkaufAbschlussProps {
  backend: Pick<DirektverkaufBackend, 'direktverkaufTaetigen'>
  receiptItems: ReceiptPosition[]
  positionen: VerkaufPositionInput[]
  totalCents: number
  // Nach erfolgreichem Verkauf: Auswahl zurücksetzen und Erfolgs-Pop auslösen
  // (im Handy-Container zusätzlich den Drawer schließen).
  verkaufAbgeschlossen: () => void
  // Der Server hat den Vorgang unter diesem Schlüssel bereits gebucht (409
  // `vorgang_daten_abweichend`), nur mit der zuerst gesendeten Auswahl. Räumt
  // Auswahl und Historie ab — siehe Direktverkauf bzw. DirektverkaufPage.
  vorgangBereitsGebucht: () => void
  // 'sheet' rendert den Bottom-Sheet-Drawer-Inhalt (Handy), 'spalte' die feste
  // Abschluss-Spalte (ab lg). Einzige Quelle des Abschluss-Inhalts; die beiden
  // Varianten unterscheiden sich nur im umschließenden Container.
  variant: 'sheet' | 'spalte'
}

// Presentation-neutraler Abschluss-Inhalt des Direktverkaufs (Beleg,
// Erhalten/Rückgeld, Kommentar, „Verkauf abschließen"). Trägt den vollständigen
// Zustand samt verkaufId-Lebenszyklus und Submit-/Fehler-/Retry-Verhalten und
// wird sowohl im Handy-Drawer als auch in der festen Spalte gerendert.
export function DirektverkaufAbschluss(props: DirektverkaufAbschlussProps) {
  const [erhaltenEuro, setErhaltenEuro] = useState('')
  const [zielbetragEuro, setZielbetragEuro] = useState('')
  const [andererAktiv, setAndererAktiv] = useState(false)
  const [kommentar, setKommentar] = useState('')

  const noPositionenSelected = props.positionen.length === 0

  // Beginnt eine Zusammenstellung aus dem Leerzustand, starten die Eingaben
  // leer: In der dauerhaften Spalte überlebt der Eingabe-State sonst einen
  // Auswahl-Reset und würde Erhalten/Kommentar eines abgebrochenen Vorgangs in
  // den nächsten tragen. React-idiomatischer State-Sync im Render (wie beim
  // Tischwechsel in TablePage), nicht per Effekt.
  const [warLeer, setWarLeer] = useState(noPositionenSelected)
  if (warLeer !== noPositionenSelected) {
    setWarLeer(noPositionenSelected)
    if (warLeer) {
      setErhaltenEuro('')
      setZielbetragEuro('')
      setAndererAktiv(false)
      setKommentar('')
    }
  }

  // verkaufId je Zusammenstellung: Jeder Wiederholversuch behält den Schlüssel —
  // auch mit inzwischen geänderter Auswahl, denn genau diese Abweichung erkennt
  // und meldet der Server. Ein neuer Schlüssel entsteht erst, wenn nach dem
  // Leeren der Auswahl (auch nach einem erfolgreichen Verkauf) eine neue
  // Zusammenstellung beginnt.
  const verkaufId = useVorgangId(noPositionenSelected)

  const { rueckgeldCents, trinkgeldCents } = calculateZahlungsbetraege(
    props.totalCents,
    parseCents(erhaltenEuro),
    parseCents(zielbetragEuro),
  )

  const { loading, run } = useActionSubmit({
    actionLabel: 'Verkauf abschließen',
    byCode: {
      kasse_nicht_geoeffnet:
        'Es ist keine Kassensitzung geöffnet. Bitte zuerst die Kasse öffnen.',
      produkt_not_found:
        'Ein ausgewähltes Produkt ist nicht mehr verfügbar. Bitte Auswahl aktualisieren.',
    },
    onCode: {
      // Der Vorgang ist gebucht, nur seine Antwort ging verloren — clientseitig
      // ist er damit abgeschlossen, auch wenn die zuletzt gesendete Auswahl eine
      // andere war. Ohne das Abräumen liefe jeder weitere Versuch wieder in
      // denselben 409, und die Differenz bliebe unverkauft.
      vorgang_daten_abweichend: () => {
        setErhaltenEuro('')
        setZielbetragEuro('')
        setAndererAktiv(false)
        setKommentar('')
        props.vorgangBereitsGebucht()
      },
    },
    onSuccess: () => {
      setErhaltenEuro('')
      setZielbetragEuro('')
      setAndererAktiv(false)
      setKommentar('')
      props.verkaufAbgeschlossen()
    },
  })

  const onSubmit = async () => {
    await run(async () => {
      await props.backend.direktverkaufTaetigen({
        verkaufId,
        positionen: props.positionen,
        kommentar,
      })
    })
  }

  const inhalt = (
    <>
      <AbschlussHeader
        variant={props.variant}
        eyebrow="Verkauf abschließen"
        title="Direktverkauf"
        description="Verkauf abschließen"
      />
      <DrawerBody className="mx-auto w-full max-w-sm">
        {noPositionenSelected ? (
          <AbschlussLeer>Produkte auswählen, um zu kassieren.</AbschlussLeer>
        ) : (
          <>
            <Receipt
              positionen={props.receiptItems}
              totalPrice={props.totalCents}
            />
            <div className="flex flex-col gap-2 px-4 pt-3">
              <div className="flex items-center justify-between gap-3">
                <Label htmlFor="erhalten">Erhalten</Label>
                <EuroInput
                  id="erhalten"
                  value={erhaltenEuro}
                  onValueChange={setErhaltenEuro}
                  className="w-28"
                />
              </div>
              <AufrundenChips
                gesamtCents={props.totalCents}
                zielbetragEuro={zielbetragEuro}
                onZielbetragEuroChange={setZielbetragEuro}
                andererAktiv={andererAktiv}
                onAndererAktivChange={setAndererAktiv}
              />
              {rueckgeldCents !== null && (
                <div className="flex items-baseline justify-between pt-1">
                  <div className="text-[15px] font-semibold">Rückgeld</div>
                  <div className="text-xl font-bold tabular-nums">
                    {formatEuro(rueckgeldCents)}
                  </div>
                </div>
              )}
              {trinkgeldCents !== null && (
                <>
                  <div className="flex justify-between font-medium">
                    <div>Trinkgeld</div>
                    <div className="tabular-nums">
                      {formatEuro(trinkgeldCents)}
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
        <Button
          disabled={loading || noPositionenSelected}
          onClick={() => {
            void onSubmit()
          }}
        >
          {loading ? <Spinner /> : null} Verkauf abschließen
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

  // Feste Spalte: dieselben Header/Body/Footer-Primitive wie im Sheet, nur in
  // einem eigenen, unabhängig scrollenden Container. group/drawer-content +
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
