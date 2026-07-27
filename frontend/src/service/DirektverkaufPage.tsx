import { useCallback, useState } from 'react'

import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useIsMobile } from '@/hooks/use-mobile'
import { BackendSingleton } from '@/lib/Backend'

import { Direktverkauf } from './components/direktverkauf/Direktverkauf'
import { DirektverkaufHistorie } from './components/direktverkauf/DirektverkaufHistorie'
import { ErfolgsPop } from './components/ErfolgsPop'
import { dockFreiraum, ServiceDock } from './components/ServiceDock'
import { DirektverkaufBackend } from './direktverkauf/DirektverkaufBackend'
import { useDirektverkaufHistorie } from './direktverkauf/hooks'
import { useAktiveProdukte } from './product/hooks'

const direktverkaufBackend = new DirektverkaufBackend(BackendSingleton)

export function DirektverkaufPage() {
  const isMobile = useIsMobile()
  const { produkte, isPending } = useAktiveProdukte()
  const {
    historie,
    isPending: historieLoading,
    isError: historieError,
    refetch: reloadHistorie,
  } = useDirektverkaufHistorie()

  // Erfolgs-Pop: Der abgeschlossene Verkauf und der Storno öffnen ihn mit ihrer
  // Meldung (statt eines Erfolgs-Toasts). Der Refetch der Historie läuft erst
  // beim Schließen, damit die Änderung dem Pop folgt.
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
      historieError={historieError}
      onErneutVersuchen={() => {
        void reloadHistorie()
      }}
      backend={direktverkaufBackend}
      onErfolg={zeigeErfolg}
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
          // Höhenbegrenzte Flex-Spalte (Viewport minus Header und Content-
          // Padding aus ServiceLayout); die Reiter-Höhe ergibt sich per Flex,
          // der Split füllt via h-full den Rest und scrollt in seinen Spalten.
          <div className="flex h-[calc(100dvh-5.5rem)] flex-col xl:h-[calc(100dvh-6.5rem)]">
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
      <ErfolgsPop
        open={erfolg.open}
        text={erfolg.text}
        onDismiss={erfolgSchliessen}
      />
    </>
  )
}
