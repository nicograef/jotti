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
import { formatEuro, parseCents } from '@/lib/utils'

import type { Position } from '../../table/Bestellung'
import type { Tisch } from '../../table/Tisch'
import type { TischBackend } from '../../table/TischBackend'
import { AbschlussHeader } from './AbschlussHeader'
import { AbschlussLeer } from './AbschlussLeer'
import { AufrundenChips } from './AufrundenChips'
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
  // Idempotenz-Schlüssel je Zusammenstellung: Jeder Wiederholversuch behält den
  // Schlüssel — auch mit inzwischen geänderter Auswahl, denn genau diese
  // Abweichung erkennt und meldet der Server. Er wird von TablePage vergeben,
  // wo auch die Auswahl liegt: Der ausgehängte Tab-Inhalt darf ihn nicht mit
  // sich nehmen, sonst würde ein Wiederholversuch zur zweiten Buchung.
  vorgangId: string
  zahlungKassiert: () => void
  // Der Server hat den Vorgang unter diesem Schlüssel bereits gebucht (409
  // `vorgang_daten_abweichend`), nur mit der zuerst gesendeten Auswahl. Räumt
  // Auswahl und Tischzustand ab — siehe TablePage, wo beides liegt.
  vorgangBereitsGebucht: () => void
  // 'sheet' rendert den Bottom-Sheet-Drawer-Inhalt (Handy), 'spalte' die feste
  // Abschluss-Spalte (ab lg). Einzige Quelle des Abschluss-Inhalts; die beiden
  // Varianten unterscheiden sich nur im umschließenden Container.
  variant: 'sheet' | 'spalte'
}

// Presentation-neutraler Abschluss-Inhalt des Tisch-Kassierens (Beleg, Erhalten,
// Zielbetrag inkl. Trinkgeld, Rückgeld, Trinkgeld-Hinweis, Kommentar,
// „Kassieren"). Trägt den Eingabe-State und das Submit-/Fehler-/Retry-Verhalten
// und wird sowohl im Handy-Drawer als auch in der festen Spalte gerendert;
// Auswahl und vorgangId kommen wie bei BestellungAbschluss von TablePage.
export function ZahlungAbschluss(props: ZahlungAbschlussProps) {
  const [kommentar, setKommentar] = useState('')
  const [erhaltenEuro, setErhaltenEuro] = useState('')
  const [zielbetragEuro, setZielbetragEuro] = useState('')
  const [andererAktiv, setAndererAktiv] = useState(false)

  const noPositionenSelected = props.positionenToPay.length === 0

  // Beginnt eine Zusammenstellung aus dem Leerzustand, starten die Eingaben
  // leer: In der dauerhaften Spalte überlebt der Eingabe-State sonst einen
  // Auswahl-Reset und würde Werte aus einer abgebrochenen Zahlung übertragen.
  // React-idiomatischer State-Sync im Render (wie beim Tischwechsel in
  // TablePage), nicht per Effekt.
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

  const positionen = toPositionRefs(props.positionenToPay)

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
    onCode: {
      // Der Vorgang ist gebucht, nur seine Antwort ging verloren — clientseitig
      // ist er damit abgeschlossen, auch wenn die zuletzt gesendete Auswahl eine
      // andere war. Ohne das Abräumen liefe jeder weitere Versuch wieder in
      // denselben 409, und die Differenz bliebe unkassiert.
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
      props.zahlungKassiert()
    },
  })

  const onSubmit = async () => {
    await run(async () => {
      await props.backend.zahlungKassieren({
        vorgangId: props.vorgangId,
        tischId: props.tisch.id,
        positionen,
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
