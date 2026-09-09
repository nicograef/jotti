import { EuroInput } from '@/components/common/EuroInput'
import { Label } from '@/components/ui/label'
import { formatEuro } from '@/lib/utils'

import { AufrundenChips } from './AufrundenChips'

interface BarzahlungFelderProps {
  gesamtCents: number
  erhaltenEuro: string
  onErhaltenEuroChange: (wert: string) => void
  zielbetragEuro: string
  onZielbetragEuroChange: (wert: string) => void
  andererAktiv: boolean
  onAndererAktivChange: (aktiv: boolean) => void
  // Aus Erhalten und Zielbetrag berechnet (calculateZahlungsbetraege); `null`
  // heißt: noch nichts anzuzeigen.
  rueckgeldCents: number | null
  trinkgeldCents: number | null
}

// Bargeld-Eingaben eines Abschlusses: Erhalten, Aufrunden-Chips, Rückgeld und
// der Trinkgeld-Hinweis. Geteilt von Tisch-Kassieren und Direktverkauf, die
// beide bar kassieren; der Zustand liegt beim jeweiligen Abschluss.
export function BarzahlungFelder({
  gesamtCents,
  erhaltenEuro,
  onErhaltenEuroChange,
  zielbetragEuro,
  onZielbetragEuroChange,
  andererAktiv,
  onAndererAktivChange,
  rueckgeldCents,
  trinkgeldCents,
}: BarzahlungFelderProps) {
  return (
    <div className="flex flex-col gap-2 px-4 pt-3">
      <div className="flex items-center justify-between gap-3">
        <Label htmlFor="erhalten">Erhalten</Label>
        <EuroInput
          id="erhalten"
          value={erhaltenEuro}
          onValueChange={onErhaltenEuroChange}
          className="w-28"
        />
      </div>
      <AufrundenChips
        gesamtCents={gesamtCents}
        zielbetragEuro={zielbetragEuro}
        onZielbetragEuroChange={onZielbetragEuroChange}
        andererAktiv={andererAktiv}
        onAndererAktivChange={onAndererAktivChange}
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
            <div className="tabular-nums">{formatEuro(trinkgeldCents)}</div>
          </div>
          <p className="text-xs text-muted-foreground">
            Trinkgeld wird nicht als Kasseneinnahme gebucht und gehört nicht in
            die Kassenlade.
          </p>
        </>
      )}
    </div>
  )
}
