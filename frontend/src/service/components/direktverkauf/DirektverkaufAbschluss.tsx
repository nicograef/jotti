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
  verkaufAbgeschlossen: () => void
  variant: 'sheet' | 'spalte'
}

export function DirektverkaufAbschluss(props: DirektverkaufAbschlussProps) {
  const [erhaltenEuro, setErhaltenEuro] = useState('')
  const [zielbetragEuro, setZielbetragEuro] = useState('')
  const [andererAktiv, setAndererAktiv] = useState(false)
  const [kommentar, setKommentar] = useState('')

  const noPositionenSelected = props.positionen.length === 0

  // verkaufId je logischem Vorgang: neu, sobald eine Zusammenstellung aus dem
  // Leerzustand beginnt (ein erfolgreicher Abschluss leert die Auswahl). Ein
  // Retry desselben Vorgangs behält seinen Schlüssel; mit dem neuen Schlüssel
  // starten auch die Eingaben leer, damit in der dauerhaften Spalte nichts aus
  // einem abgebrochenen Vorgang übertragen wird.
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
