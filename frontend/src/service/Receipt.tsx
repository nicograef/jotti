import { useState } from 'react'

import { Button } from '@/components/ui/button'
import { formatEuro } from '@/lib/utils'

import { PositionsKommentarFeld } from './CommentField'

export interface ReceiptPosition {
  // Eindeutig je Zeile: zwei Runden derselben Variante ergeben zwei Positionen
  // mit gleichem Namen und Preis.
  id: string
  name: string
  einzelpreisCents: number
  menge: number
  kommentar?: string
}

export function Receipt({
  positionen,
  totalPrice,
  kommentarBearbeiten,
}: {
  positionen: ReceiptPosition[]
  totalPrice?: number
  // Gesetzt, wenn jede Zeile ihren Positionskommentar bearbeiten lässt.
  kommentarBearbeiten?: (id: string, text: string) => void
}) {
  return (
    <>
      <div className="px-4 pt-2 pb-0 space-y-2">
        {positionen.map((position) => (
          <div key={position.id} className="border-b pb-2 last:border-0">
            <div className="flex justify-between">
              <div>
                {position.menge} x {position.name}
              </div>
              <div>
                {formatEuro(position.einzelpreisCents * position.menge)}
              </div>
            </div>
            {kommentarBearbeiten ? (
              <PositionsKommentar
                position={position}
                bearbeiten={kommentarBearbeiten}
              />
            ) : (
              position.kommentar && (
                <div className="text-sm text-muted-foreground">
                  „{position.kommentar}“
                </div>
              )
            )}
          </div>
        ))}
      </div>
      {totalPrice !== undefined && (
        <div className="flex justify-between font-bold px-4 pt-2 pb-4 border-t-2">
          <div>Gesamt</div>
          <div>{formatEuro(totalPrice)}</div>
        </div>
      )}
    </>
  )
}

// Eine Zeile mit Kommentar zeigt ihr Feld offen; einmal geöffnet, bleibt es offen,
// auch wenn der Text wieder leer wird.
function PositionsKommentar({
  position,
  bearbeiten,
}: {
  position: ReceiptPosition
  bearbeiten: (id: string, text: string) => void
}) {
  const [geoeffnet, setGeoeffnet] = useState(false)

  if (!geoeffnet && !position.kommentar) {
    return (
      <Button
        variant="link"
        size="sm"
        className="h-auto px-0"
        onClick={() => {
          setGeoeffnet(true)
        }}
      >
        Kommentar
      </Button>
    )
  }

  return (
    <div className="pt-1">
      <PositionsKommentarFeld
        defaultValue={position.kommentar ?? ''}
        autoFocus={geoeffnet}
        onChange={(text) => {
          setGeoeffnet(true)
          bearbeiten(position.id, text)
        }}
      />
    </div>
  )
}
