import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { BackendSingleton } from '@/lib/Backend'

import { Direktverkauf } from './components/direktverkauf/Direktverkauf'
import { DirektverkaufHistorie } from './components/direktverkauf/DirektverkaufHistorie'
import { ServiceDock } from './components/ServiceDock'
import { DirektverkaufBackend } from './direktverkauf/DirektverkaufBackend'
import { useDirektverkaufHistorie } from './direktverkauf/hooks'
import { useAktiveProdukte } from './product/hooks'

const direktverkaufBackend = new DirektverkaufBackend(BackendSingleton)

// Unterer Freiraum der Tab-Inhalte in Dock-Höhe (Aktionsbutton plus TabsList
// plus Innenabstände), damit die letzte Zeile über dem fixierten ServiceDock
// endet und antippbar bleibt.
const dockFreiraum = 'pb-[calc(9rem+env(safe-area-inset-bottom,0px))]'

export function DirektverkaufPage() {
  const { produkte, isPending } = useAktiveProdukte()
  const {
    historie,
    isPending: historieLoading,
    refetch: reloadHistorie,
  } = useDirektverkaufHistorie()

  return (
    <Tabs defaultValue="verkaufen">
      <ServiceDock
        leiste={
          <TabsList className="h-10 w-full">
            <TabsTrigger value="verkaufen" className="flex-1">
              Verkaufen
            </TabsTrigger>
            <TabsTrigger value="historie" className="flex-1">
              Historie
            </TabsTrigger>
          </TabsList>
        }
      >
        <TabsContent value="verkaufen" className={dockFreiraum}>
          <Direktverkauf
            backend={direktverkaufBackend}
            products={produkte}
            productsLoading={isPending}
            onVerkauft={() => {
              void reloadHistorie()
            }}
          />
        </TabsContent>
        <TabsContent value="historie" className={dockFreiraum}>
          <DirektverkaufHistorie
            historie={historie}
            historieLoading={historieLoading}
            backend={direktverkaufBackend}
            onStorniert={() => {
              void reloadHistorie()
            }}
          />
        </TabsContent>
      </ServiceDock>
    </Tabs>
  )
}
