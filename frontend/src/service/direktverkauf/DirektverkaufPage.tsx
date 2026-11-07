import { useQueryClient } from '@tanstack/react-query'
import { useCallback } from 'react'

import { LadefehlerAlert } from '@/components/common/LadefehlerAlert'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useErfolgsPop } from '@/hooks/use-erfolgs-pop'
import { useIsMobile } from '@/hooks/use-mobile'
import { useVorgang } from '@/hooks/use-vorgang'
import { cn } from '@/lib/utils'

import { ErfolgsPop } from '../ErfolgsPop'
import { PFANDRUECKGABE_HISTORIE_KEY } from '../pfand/hooks'
import { PfandrueckgabenAbschnitt } from '../pfand/PfandrueckgabenAbschnitt'
import { useAktiveProdukte } from '../product/hooks'
import { dockFreiraum, ServiceDock } from '../ServiceDock'
import { seitenHoehe } from '../ServiceLayout'
import { VorgangHistorie } from '../VorgangHistorie'
import { DirektverkaufTab } from './DirektverkaufTab'
import { direktverkaufBackend, useDirektverkaufHistorie } from './hooks'

export function DirektverkaufPage() {
  const isMobile = useIsMobile()
  const queryClient = useQueryClient()
  const {
    produkte,
    pfandProdukte,
    isPending,
    isError: produkteError,
    refetch: reloadProdukte,
  } = useAktiveProdukte()
  const {
    historie,
    isPending: historieLoading,
    isError: historieError,
    refetch: reloadHistorie,
  } = useDirektverkaufHistorie()
  // Korb, verkaufId und Pfand-Vorgang liegen hier, damit sie einen Blick in die
  // Historie überstehen; ein Retry nach verlorener Antwort buchte sonst doppelt.
  const verkauf = useVorgang<number>()
  const pfandVorgang = useVorgang<number>()

  const reload = useCallback(() => {
    void reloadHistorie()
    void queryClient.invalidateQueries({
      queryKey: [PFANDRUECKGABE_HISTORIE_KEY],
    })
  }, [reloadHistorie, queryClient])
  const { zeigeErfolg, pop } = useErfolgsPop(reload)

  const tabTrigger = (
    <TabsList className="h-10 w-full">
      <TabsTrigger value="verkaufen" className="flex-1">
        Verkaufen
      </TabsTrigger>
      <TabsTrigger value="historie" className="flex-1">
        Historie
      </TabsTrigger>
    </TabsList>
  )

  // Ladefehler je Reiter: eine leere Produktliste sähe wie ein leeres Sortiment
  // aus, eine leere Historie wie ein Tag ohne Verkäufe.
  const verkaufenInhalt = produkteError ? (
    <LadefehlerAlert
      titel="Produkte konnten nicht geladen werden"
      onErneutVersuchen={() => {
        void reloadProdukte()
      }}
    />
  ) : (
    <DirektverkaufTab
      products={produkte}
      pfandProdukte={pfandProdukte}
      vorgang={verkauf}
      pfandVorgang={pfandVorgang}
      productsLoading={isPending}
      onErfolg={zeigeErfolg}
    />
  )
  const historieInhalt = historieError ? (
    <LadefehlerAlert
      titel="Historie konnte nicht geladen werden"
      onErneutVersuchen={() => {
        void reloadHistorie()
      }}
    />
  ) : (
    <>
      <VorgangHistorie
        eintraege={historie.map((verkauf) => ({
          ...verkauf,
          id: verkauf.verkaufId,
        }))}
        loading={historieLoading}
        bezeichnung="Verkauf"
        leerHinweis="Noch keine Direktverkäufe in dieser Kassensitzung."
        stornieren={(verkaufId, positionen, kommentar) =>
          direktverkaufBackend.direktverkaufStornieren({
            verkaufId,
            positionen,
            kommentar,
          })
        }
        belegDrucken={(verkaufId, stornierungId) =>
          direktverkaufBackend.kassenbelegDrucken({ verkaufId, stornierungId })
        }
        onErfolg={zeigeErfolg}
      />
      <PfandrueckgabenAbschnitt tischId={0} onErfolg={zeigeErfolg} />
    </>
  )

  return (
    <>
      <Tabs defaultValue="verkaufen">
        {isMobile ? (
          <ServiceDock leiste={tabTrigger}>
            <TabsContent value="verkaufen" className={dockFreiraum}>
              {verkaufenInhalt}
            </TabsContent>
            <TabsContent value="historie" className={dockFreiraum}>
              {historieInhalt}
            </TabsContent>
          </ServiceDock>
        ) : (
          // Höhenbegrenzte Flex-Spalte; der Split füllt via h-full den Rest.
          <div className={cn('flex flex-col', seitenHoehe)}>
            <div className="mb-4 max-w-md">{tabTrigger}</div>
            <TabsContent value="verkaufen" className="min-h-0 flex-1">
              {verkaufenInhalt}
            </TabsContent>
            <TabsContent
              value="historie"
              className="min-h-0 flex-1 overflow-y-auto"
            >
              {historieInhalt}
            </TabsContent>
          </div>
        )}
      </Tabs>
      <ErfolgsPop {...pop} />
    </>
  )
}
