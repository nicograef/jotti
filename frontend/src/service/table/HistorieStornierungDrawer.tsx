import { StornoDrawer } from '../StornoDrawer'
import type { Bestellung } from './Bestellung'
import { QuelleDrawerHeader } from './DrawerParts'
import type { Tisch } from './Tisch'
import type { TischBackend } from './TischBackend'
import type { Umbuchung } from './Umbuchung'

interface HistorieStornierungDrawerProps {
  backend: Pick<TischBackend, 'stornierungErteilen'>
  tisch: Tisch
  // Ursprungsvorgang (Bestellung oder Umbuchungs-Zugang), dessen Positionen storniert
  // werden; beschriftet den Drawer.
  quelle: Bestellung | Umbuchung
  onClose: () => void
  onStornierungErteilt: () => void
}

export function HistorieStornierungDrawer({
  backend,
  tisch,
  quelle,
  onClose,
  onStornierungErteilt,
}: HistorieStornierungDrawerProps) {
  return (
    <StornoDrawer
      kopf={
        <QuelleDrawerHeader quelle={quelle}>
          Positionen aus diesem Vorgang zum Stornieren auswählen.
        </QuelleDrawerHeader>
      }
      positionen={quelle.stornierbarePositionen}
      stornieren={(positionen, kommentar) =>
        backend.stornierungErteilen({
          tischId: tisch.id,
          positionen,
          kommentar,
        })
      }
      onClose={onClose}
      onStorniert={onStornierungErteilt}
    />
  )
}
