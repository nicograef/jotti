import { useCallback, useState } from 'react'

import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useIsMobile } from '@/hooks/use-mobile'
import { BackendSingleton } from '@/lib/Backend'

import { Direktverkauf } from './components/direktverkauf/Direktverkauf'
import { DirektverkaufHistorie } from './components/direktverkauf/DirektverkaufHistorie'
import { ErfolgsPop } from './components/ErfolgsPop'
import { ServiceDock } from './components/ServiceDock'
import { DirektverkaufBackend } from './direktverkauf/DirektverkaufBackend'
import { useDirektverkaufHistorie } from './direktverkauf/hooks'
import { useAktiveProdukte } from './product/hooks'

const direktverkaufBackend = new DirektverkaufBackend(BackendSingleton)

// Unterer Freiraum der Tab-Inhalte in Dock-Höhe (Aktionsbutton plus TabsList
// plus Innenabstände), damit die letzte Zeile über dem fixierten ServiceDock
// endet und antippbar bleibt. Nur im Handy-Layout (unter lg) relevant.
const dockFreiraum = 'pb-[calc(9rem+env(safe-area-inset-bottom,0px))]'

export function DirektverkaufPage() {
  const isMobile = useIsMobile()
  const { produkte, isPending } = useAktiveProdukte()
  const {
    historie,
    isPending: historieLoading,
    refetch: reloadHistorie,
  } = useDirektverkaufHistorie()

  // Erfolgs-Pop: Der abgeschlossene Verkauf öffnet ihn mit seiner Meldung (statt
  // eines Erfolgs-Toasts). Der Refetch der Historie läuft erst beim Schließen,
  // damit der neue Eintrag dem Pop folgt. Der Storno-Pfad lädt weiterhin sofort.
  const [erfolg, setErfolg] = useState({ open: false, text: '' })
  const zeigeErfolg = useCallback((nachricht: string) => {
    setErfolg({ open: true, text: nachricht })
  }, [])
  const erfolgSchliessen = useCallback(() => {
    setErfolg((prev) => ({ ...prev, open: false }))
    void reloadHistorie()
  }, [reloadHistorie])

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

  const verkaufenInhalt = (
    <Direktverkauf
      backend={direktverkaufBackend}
      products={produkte}
      productsLoading={isPending}
      onErfolg={zeigeErfolg}
    />
  )
  const historieInhalt = (
    <DirektverkaufHistorie
      historie={historie}
      historieLoading={historieLoading}
      backend={direktverkaufBackend}
      onStorniert={() => {
        void reloadHistorie()
      }}
    />
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
          <>
            <div className="mb-4 max-w-md">{tabTrigger}</div>
            <TabsContent value="verkaufen">{verkaufenInhalt}</TabsContent>
            <TabsContent value="historie">{historieInhalt}</TabsContent>
          </>
        )}
      </Tabs>
      <ErfolgsPop
        open={erfolg.open}
        text={erfolg.text}
        onDismiss={erfolgSchliessen}
      />
    </>
  )
}
