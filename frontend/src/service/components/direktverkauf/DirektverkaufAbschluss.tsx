import { useEffect, useRef, useState } from 'react'

import { Button } from '@/components/ui/button'
import { DrawerBody, DrawerClose, DrawerFooter } from '@/components/ui/drawer'
import { Spinner } from '@/components/ui/spinner'
import { useActionSubmit } from '@/hooks/use-action-submit'
import { parseCents } from '@/lib/utils'

import type { VerkaufPositionInput } from '../../direktverkauf/Direktverkauf'
import type { DirektverkaufBackend } from '../../direktverkauf/DirektverkaufBackend'
import { AbschlussContainer } from '../table/AbschlussContainer'
import { AbschlussHeader } from '../table/AbschlussHeader'
import { AbschlussLeer } from '../table/AbschlussLeer'
import { BarzahlungFelder } from '../table/BarzahlungFelder'
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

  // verkaufId je logischem Vorgang: neu, sobald eine Zusammenstellung aus dem
  // Leerzustand beginnt, und — weil ein erfolgreicher Abschluss die Auswahl leert
  // — erneut beim nächsten Aufbau. Ein Retry desselben Vorgangs behält seinen
  // Schlüssel, weil die Auswahl dabei nicht leer wird. Mit dem neuen Schlüssel
  // starten auch die Eingaben leer: In der dauerhaften Spalte überlebt der State
  // sonst über einen Auswahl-Reset hinweg und würde Erhalten/Kommentar eines
  // abgebrochenen Vorgangs in den nächsten tragen (der Idempotenz-Schlüssel und
  // die Eingaben bleiben so an derselben Vorgangsgrenze konsistent).
  const [verkaufId, setVerkaufId] = useState(() => crypto.randomUUID())
  const warLeerRef = useRef(noPositionenSelected)
  useEffect(() => {
    if (warLeerRef.current && !noPositionenSelected) {
      setVerkaufId(crypto.randomUUID())
      setErhaltenEuro('')
      setZielbetragEuro('')
      setAndererAktiv(false)
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
    actionLabel: 'Verkauf abschließen',
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
            <BarzahlungFelder
              gesamtCents={props.totalCents}
              erhaltenEuro={erhaltenEuro}
              onErhaltenEuroChange={setErhaltenEuro}
              zielbetragEuro={zielbetragEuro}
              onZielbetragEuroChange={setZielbetragEuro}
              andererAktiv={andererAktiv}
              onAndererAktivChange={setAndererAktiv}
              rueckgeldCents={rueckgeldCents}
              trinkgeldCents={trinkgeldCents}
            />
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

  return (
    <AbschlussContainer variant={props.variant} pending={loading}>
      {inhalt}
    </AbschlussContainer>
  )
}
