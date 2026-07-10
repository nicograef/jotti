import * as React from "react"
import { Dialog as DrawerPrimitive } from "radix-ui"

import { cn } from "@/lib/utils"

function Drawer({
  ...props
}: React.ComponentProps<typeof DrawerPrimitive.Root>) {
  return <DrawerPrimitive.Root data-slot="drawer" {...props} />
}

function DrawerTrigger({
  ...props
}: React.ComponentProps<typeof DrawerPrimitive.Trigger>) {
  return <DrawerPrimitive.Trigger data-slot="drawer-trigger" {...props} />
}

function DrawerPortal({
  ...props
}: React.ComponentProps<typeof DrawerPrimitive.Portal>) {
  return <DrawerPrimitive.Portal data-slot="drawer-portal" {...props} />
}

function DrawerClose({
  ...props
}: React.ComponentProps<typeof DrawerPrimitive.Close>) {
  return <DrawerPrimitive.Close data-slot="drawer-close" {...props} />
}

function DrawerOverlay({
  className,
  ...props
}: React.ComponentProps<typeof DrawerPrimitive.Overlay>) {
  return (
    <DrawerPrimitive.Overlay
      data-slot="drawer-overlay"
      className={cn(
        "fixed inset-0 z-50 bg-black/10 duration-100 supports-backdrop-filter:backdrop-blur-xs data-open:animate-in data-open:fade-in-0 data-closed:animate-out data-closed:fade-out-0",
        className
      )}
      {...props}
    />
  )
}

function DrawerContent({
  className,
  children,
  pending = false,
  onEscapeKeyDown,
  onInteractOutside,
  ...props
}: React.ComponentProps<typeof DrawerPrimitive.Content> & {
  // Laufender Submit: Escape und Backdrop-Tap schließen den Drawer nicht,
  // DrawerBody wird gedimmt und nimmt keine Eingaben an.
  pending?: boolean
}) {
  // Radix ruft die Dismiss-Handler aus dem Closure ihrer Registrierung auf;
  // ein dort eingefrorenes `pending` wäre veraltet. Die Ref liefert dem
  // Handler immer den aktuellen Wert.
  const pendingRef = React.useRef(pending)
  React.useEffect(() => {
    pendingRef.current = pending
  }, [pending])

  return (
    <DrawerPortal data-slot="drawer-portal">
      <DrawerOverlay />
      <DrawerPrimitive.Content
        data-slot="drawer-content"
        data-pending={pending || undefined}
        onEscapeKeyDown={(event) => {
          if (pendingRef.current) event.preventDefault()
          onEscapeKeyDown?.(event)
        }}
        onInteractOutside={(event) => {
          if (pendingRef.current) event.preventDefault()
          onInteractOutside?.(event)
        }}
        className={cn(
          "group/drawer-content fixed inset-x-0 bottom-0 z-50 flex max-h-[85dvh] flex-col rounded-t-xl border-t bg-popover pb-[env(safe-area-inset-bottom,0px)] text-sm text-popover-foreground duration-200 data-open:animate-in data-open:fade-in-0 data-open:slide-in-from-bottom-10 data-closed:animate-out data-closed:fade-out-0 data-closed:slide-out-to-bottom-10",
          className
        )}
        {...props}
      >
        {children}
      </DrawerPrimitive.Content>
    </DrawerPortal>
  )
}

function DrawerHeader({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="drawer-header"
      className={cn(
        "flex flex-col gap-0.5 p-4 text-center md:gap-1.5 md:text-left",
        className
      )}
      {...props}
    />
  )
}

// DrawerBody ist der einzige Scrollbereich des Drawers. Header und Footer
// sind direkte Flex-Kinder von DrawerContent und bleiben dadurch immer
// sichtbar — auch bei beliebig langem Inhalt.
function DrawerBody({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="drawer-body"
      className={cn(
        "min-h-0 overflow-y-auto group-data-[pending]/drawer-content:pointer-events-none group-data-[pending]/drawer-content:opacity-50",
        className
      )}
      {...props}
    />
  )
}

function DrawerFooter({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="drawer-footer"
      className={cn("mt-auto flex flex-col gap-2 p-4", className)}
      {...props}
    />
  )
}

function DrawerTitle({
  className,
  ...props
}: React.ComponentProps<typeof DrawerPrimitive.Title>) {
  return (
    <DrawerPrimitive.Title
      data-slot="drawer-title"
      className={cn("font-heading font-medium text-foreground", className)}
      {...props}
    />
  )
}

function DrawerDescription({
  className,
  ...props
}: React.ComponentProps<typeof DrawerPrimitive.Description>) {
  return (
    <DrawerPrimitive.Description
      data-slot="drawer-description"
      className={cn("text-sm text-muted-foreground", className)}
      {...props}
    />
  )
}

export {
  Drawer,
  DrawerPortal,
  DrawerOverlay,
  DrawerTrigger,
  DrawerClose,
  DrawerContent,
  DrawerHeader,
  DrawerBody,
  DrawerFooter,
  DrawerTitle,
  DrawerDescription,
}
