import { createContext, use, useState } from 'react'
import { createPortal } from 'react-dom'

// Der Aktionsinhalt bleibt in den Drawer-Komponenten (er braucht deren
// Mengen-State und den Radix-DrawerTrigger-Kontext) und rendert über
// DockActionSlot per Portal in den Slot des Docks; React-Context bleibt über das
// Portal hinweg erhalten. Der Kontext umschließt Seiteninhalt und Dock, damit
// das Portal aus dem Tab-Inhalt heraus funktioniert.

const DockSlotContext = createContext<HTMLElement | null>(null)

export function DockActionSlot({ children }: { children: React.ReactNode }) {
  const slot = use(DockSlotContext)
  if (slot === null) return null
  return createPortal(children, slot)
}

// Unterer Freiraum der Tab-Inhalte in Dock-Höhe, damit die letzte Zeile über dem
// fixierten Dock endet und antippbar bleibt. Nur unter lg relevant.
export const dockFreiraum = 'pb-[calc(9rem+env(safe-area-inset-bottom,0px))]'

interface ServiceDockProps {
  // Seiteninhalt oberhalb des Docks (Tab-Inhalte). Er muss innerhalb des
  // Kontexts liegen, damit DockActionSlot aus ihm heraus portalen kann.
  children: React.ReactNode
  leiste: React.ReactNode
}

export function ServiceDock({ children, leiste }: ServiceDockProps) {
  const [slot, setSlot] = useState<HTMLElement | null>(null)

  return (
    <DockSlotContext value={slot}>
      {children}
      <div className="fixed inset-x-0 bottom-0 z-40 border-t bg-background px-4 pt-3 pb-[calc(1rem+env(safe-area-inset-bottom,0px))]">
        <div className="mx-auto flex w-full max-w-md flex-col gap-2">
          <div ref={setSlot} className="flex flex-col gap-2" />
          {leiste}
        </div>
      </div>
    </DockSlotContext>
  )
}
