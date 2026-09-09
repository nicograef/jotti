import { DrawerContent } from '@/components/ui/drawer'

interface AbschlussContainerProps {
  // 'sheet' rendert den Bottom-Sheet-Drawer-Inhalt (Handy), 'spalte' die feste
  // Abschluss-Spalte (ab lg).
  variant: 'sheet' | 'spalte'
  pending: boolean
  children: React.ReactNode
}

// Umschließender Container eines Abschluss-Inhalts (Bestellung, Zahlung,
// Direktverkauf). Die feste Spalte nutzt dieselben Header/Body/Footer-Primitive
// wie das Sheet, nur in einem eigenen, unabhängig scrollenden Container.
// group/drawer-content + data-pending übernehmen dort das Body-Dimming des
// Drawers während des Submits.
export function AbschlussContainer({
  variant,
  pending,
  children,
}: AbschlussContainerProps) {
  if (variant === 'sheet') {
    return <DrawerContent pending={pending}>{children}</DrawerContent>
  }

  return (
    <aside
      data-pending={pending || undefined}
      className="group/drawer-content flex min-h-0 flex-col overflow-hidden rounded-xl border bg-popover text-sm text-popover-foreground"
    >
      {children}
    </aside>
  )
}
