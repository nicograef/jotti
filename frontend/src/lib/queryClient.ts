import { QueryCache, QueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

// createQueryClient baut den globalen QueryClient mit zentralem Fehler-Handling:
// Ohne diesen Handler verschwinden Query-Fehler stumm und die Seiten zeigen
// Leer-Defaults (z. B. Saldo 0,00 €). Die feste Toast-ID sorgt dafür, dass
// mehrere gleichzeitig fehlschlagende Queries (z. B. bei Netzabbruch) nur
// einen Toast erzeugen.
export function createQueryClient(): QueryClient {
  return new QueryClient({
    queryCache: new QueryCache({
      onError: (error) => {
        console.error(error)
        toast.error(
          'Daten konnten nicht geladen werden. Bitte Verbindung prüfen und erneut versuchen.',
          { id: 'query-fehler' },
        )
      },
    }),
  })
}
