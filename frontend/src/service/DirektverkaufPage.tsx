import { BackendSingleton } from '@/lib/Backend'

import { Direktverkauf } from './components/direktverkauf/Direktverkauf'
import { DirektverkaufBackend } from './direktverkauf/DirektverkaufBackend'
import { useAktiveProdukte } from './product/hooks'

const direktverkaufBackend = new DirektverkaufBackend(BackendSingleton)

export function DirektverkaufPage() {
  const { produkte, isPending } = useAktiveProdukte()

  return (
    <div className="space-y-4 py-2">
      <h1 className="text-2xl font-bold">Direktverkauf</h1>
      <Direktverkauf
        backend={direktverkaufBackend}
        products={produkte}
        productsLoading={isPending}
      />
    </div>
  )
}
