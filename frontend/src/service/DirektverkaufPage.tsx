import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { BackendSingleton } from '@/lib/Backend'

import { Direktverkauf } from './components/direktverkauf/Direktverkauf'
import { DirektverkaufHistorie } from './components/direktverkauf/DirektverkaufHistorie'
import { DirektverkaufBackend } from './direktverkauf/DirektverkaufBackend'
import { useDirektverkaufHistorie } from './direktverkauf/hooks'
import { useAktiveProdukte } from './product/hooks'

const direktverkaufBackend = new DirektverkaufBackend(BackendSingleton)

export function DirektverkaufPage() {
  const { produkte, isPending } = useAktiveProdukte()
  const {
    historie,
    isPending: historieLoading,
    refetch: reloadHistorie,
  } = useDirektverkaufHistorie()

  return (
    <div className="space-y-4 py-2">
      <h1 className="text-2xl font-bold">Direktverkauf</h1>
      <Tabs defaultValue="verkaufen">
        <TabsList>
          <TabsTrigger value="verkaufen" className="p-4">
            Verkaufen
          </TabsTrigger>
          <TabsTrigger value="historie" className="p-4">
            Historie
          </TabsTrigger>
        </TabsList>
        <TabsContent value="verkaufen">
          <Direktverkauf
            backend={direktverkaufBackend}
            products={produkte}
            productsLoading={isPending}
            onVerkauft={() => {
              void reloadHistorie()
            }}
          />
        </TabsContent>
        <TabsContent value="historie">
          <DirektverkaufHistorie
            historie={historie}
            historieLoading={historieLoading}
            backend={direktverkaufBackend}
            onStorniert={() => {
              void reloadHistorie()
            }}
          />
        </TabsContent>
      </Tabs>
    </div>
  )
}
