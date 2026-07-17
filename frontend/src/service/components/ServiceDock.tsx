import { createContext, use, useState } from 'react'
import { createPortal } from 'react-dom'

// ServiceDock ist die eine opake Bodenfläche des Service-Bereichs: oben ein
// Aktions-Slot (Button, in Phase 3 zusätzlich die Restbetrag-Zeile), darunter
// die Tab-Leiste in voller Breite. Es ersetzt die zwei früher übereinander
// schwebenden Leisten und gilt auf allen Viewports.
//
// Der Aktionsinhalt bleibt in den Drawer-Komponenten (er braucht deren
// Mengen-State und den Radix-DrawerTrigger-Kontext) und rendert über
// DockActionSlot per Portal in den hier bereitgestellten Slot. React-Context —
// und damit der DrawerTrigger — bleibt über das Portal hinweg erhalten. Der
// Kontext umschließt sowohl den Seiteninhalt (die Aktions-Quelle) als auch das
// Dock (das Slot-Ziel), damit das Portal aus dem Tab-Inhalt heraus funktioniert.

const DockSlotContext = createContext<HTMLElement | null>(null)

// DockActionSlot rendert seine Kinder per Portal in den Aktions-Slot des Docks.
// Solange der Slot noch nicht gemountet ist (erster Render), rendert es nichts.
export function DockActionSlot({ children }: { children: React.ReactNode }) {
  const slot = use(DockSlotContext)
  if (slot === null) return null
  return createPortal(children, slot)
}

// Unterer Freiraum der Tab-Inhalte in Dock-Höhe (Aktionsbutton plus TabsList plus
// Innenabstände), damit die letzte Zeile über dem fixierten Dock endet und
// antippbar bleibt. Nur im Handy-Layout (unter lg) relevant; ab lg trägt die
// Abschluss-Spalte den Button selbst.
export const dockFreiraum = 'pb-[calc(9rem+env(safe-area-inset-bottom,0px))]'

interface ServiceDockProps {
  // Seiteninhalt oberhalb des Docks (Tab-Inhalte). Er muss innerhalb des
  // Kontexts liegen, damit DockActionSlot aus ihm heraus portalen kann.
  children: React.ReactNode
  // Inhalt der unteren Dock-Zeile (die TabsList); steht unter dem Aktions-Slot.
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
