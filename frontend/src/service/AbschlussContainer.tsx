import { DrawerContent } from '@/components/ui/drawer'

interface AbschlussContainerProps {
  // 'sheet' = Bottom-Sheet-Drawer (Handy), 'spalte' = feste Abschluss-Spalte (ab lg).
  variant: 'sheet' | 'spalte'
  pending: boolean
  children: React.ReactNode
}

// group/drawer-content + data-pending übernehmen in der Spalte das Body-Dimming
// des Drawers während des Submits.
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
