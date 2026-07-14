'use client'

import * as React from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { Tabs as TabsPrimitive } from 'radix-ui'

import { cn } from '@/lib/utils'

function Tabs({
  className,
  orientation = 'horizontal',
  ...props
}: React.ComponentProps<typeof TabsPrimitive.Root>) {
  return (
    <TabsPrimitive.Root
      data-slot="tabs"
      data-orientation={orientation}
      className={cn(
        'group/tabs flex gap-2 data-horizontal:flex-col',
        className,
      )}
      {...props}
    />
  )
}

const tabsListVariants = cva(
  'group/tabs-list inline-flex w-fit items-center justify-center rounded-lg p-[3px] text-muted-foreground group-data-horizontal/tabs:h-9 group-data-vertical/tabs:h-fit group-data-vertical/tabs:flex-col data-[variant=line]:rounded-none',
  {
    variants: {
      variant: {
        default: 'bg-muted',
        line: 'gap-1 bg-transparent',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  },
)

function TabsList({
  className,
  variant = 'default',
  ...props
}: React.ComponentProps<typeof TabsPrimitive.List> &
  VariantProps<typeof tabsListVariants>) {
  return (
    <TabsPrimitive.List
      data-slot="tabs-list"
      data-variant={variant}
      className={cn(tabsListVariants({ variant }), className)}
      {...props}
    />
  )
}

// Schwelle in Pixeln, ab der eine Seite als „scrollbar" gilt; darunter gelten
// Rundungsreste (subpixel) nicht als verborgene Tabs.
const SCROLL_EPSILON = 1

// ScrollableTabsList umschließt eine TabsList mit horizontalem Scrollen und
// zeigt auf schmalen Viewports eine Rand-Affordance (Fade plus Chevron) auf der
// Seite, auf der noch Tabs verborgen sind. Passen alle Tabs (z. B. am Desktop),
// erscheint keine Affordance und das Layout bleibt unverändert. Ein Klick auf
// ein Chevron scrollt in die jeweilige Richtung.
function ScrollableTabsList({
  className,
  children,
  ...props
}: React.ComponentProps<typeof TabsList>) {
  const scrollRef = React.useRef<HTMLDivElement>(null)
  const [canScrollLeft, setCanScrollLeft] = React.useState(false)
  const [canScrollRight, setCanScrollRight] = React.useState(false)

  const updateAffordance = React.useCallback(() => {
    const el = scrollRef.current
    if (el === null) return
    const maxScroll = el.scrollWidth - el.clientWidth
    setCanScrollLeft(el.scrollLeft > SCROLL_EPSILON)
    setCanScrollRight(el.scrollLeft < maxScroll - SCROLL_EPSILON)
  }, [])

  React.useEffect(() => {
    const el = scrollRef.current
    if (el === null) return
    updateAffordance()
    // ResizeObserver fehlt in der jsdom-Testumgebung; ohne Layout-Engine gibt es
    // dort ohnehin keine messbaren Breiten, daher überspringen wir die Messung.
    if (typeof ResizeObserver === 'undefined') return
    const observer = new ResizeObserver(updateAffordance)
    observer.observe(el)
    for (const child of el.children) observer.observe(child)
    return () => {
      observer.disconnect()
    }
  }, [updateAffordance])

  function scroll(direction: -1 | 1) {
    const el = scrollRef.current
    if (el === null) return
    el.scrollBy({ left: (direction * el.clientWidth) / 2, behavior: 'smooth' })
  }

  return (
    <div data-slot="scrollable-tabs" className="relative">
      <div
        ref={scrollRef}
        onScroll={updateAffordance}
        className="overflow-x-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
      >
        <TabsList className={cn('w-max', className)} {...props}>
          {children}
        </TabsList>
      </div>

      {canScrollLeft && (
        <button
          type="button"
          aria-hidden="true"
          tabIndex={-1}
          data-slot="tabs-scroll-hint"
          data-direction="left"
          onClick={() => {
            scroll(-1)
          }}
          className="absolute inset-y-0 left-0 flex items-center bg-gradient-to-r from-background from-40% to-transparent pr-6 pl-1"
        >
          <ChevronLeft className="size-4 text-muted-foreground" />
        </button>
      )}
      {canScrollRight && (
        <button
          type="button"
          aria-hidden="true"
          tabIndex={-1}
          data-slot="tabs-scroll-hint"
          data-direction="right"
          onClick={() => {
            scroll(1)
          }}
          className="absolute inset-y-0 right-0 flex items-center bg-gradient-to-l from-background from-40% to-transparent pr-1 pl-6"
        >
          <ChevronRight className="size-4 text-muted-foreground" />
        </button>
      )}
    </div>
  )
}

function TabsTrigger({
  className,
  ...props
}: React.ComponentProps<typeof TabsPrimitive.Trigger>) {
  return (
    <TabsPrimitive.Trigger
      data-slot="tabs-trigger"
      className={cn(
        "relative inline-flex h-[calc(100%-1px)] flex-1 items-center justify-center gap-1.5 rounded-md border border-transparent px-2 py-1 text-sm font-medium whitespace-nowrap text-foreground/60 transition-all group-data-vertical/tabs:w-full group-data-vertical/tabs:justify-start hover:text-foreground focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 focus-visible:outline-1 focus-visible:outline-ring disabled:pointer-events-none disabled:opacity-50 has-data-[icon=inline-end]:pr-1.5 has-data-[icon=inline-start]:pl-1.5 dark:text-muted-foreground dark:hover:text-foreground group-data-[variant=default]/tabs-list:data-active:shadow-sm group-data-[variant=line]/tabs-list:data-active:shadow-none [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
        'group-data-[variant=line]/tabs-list:bg-transparent group-data-[variant=line]/tabs-list:data-active:bg-transparent dark:group-data-[variant=line]/tabs-list:data-active:border-transparent dark:group-data-[variant=line]/tabs-list:data-active:bg-transparent',
        'data-active:bg-background data-active:text-foreground dark:data-active:border-input dark:data-active:bg-input/30 dark:data-active:text-foreground',
        'after:absolute after:bg-foreground after:opacity-0 after:transition-opacity group-data-horizontal/tabs:after:inset-x-0 group-data-horizontal/tabs:after:bottom-[-5px] group-data-horizontal/tabs:after:h-0.5 group-data-vertical/tabs:after:inset-y-0 group-data-vertical/tabs:after:-right-1 group-data-vertical/tabs:after:w-0.5 group-data-[variant=line]/tabs-list:data-active:after:opacity-100',
        className,
      )}
      {...props}
    />
  )
}

function TabsContent({
  className,
  ...props
}: React.ComponentProps<typeof TabsPrimitive.Content>) {
  return (
    <TabsPrimitive.Content
      data-slot="tabs-content"
      // Radix hängt den inaktiven Tab-Inhalt aus dem DOM aus; beim Aktivieren
      // mountet er neu, wodurch die fadeUp-Animation (250 ms) jedes Mal neu
      // startet (Motion-Inventar „Tab-/Detail-Wechsel").
      className={cn('flex-1 text-sm outline-none animate-fade-up', className)}
      {...props}
    />
  )
}

export {
  Tabs,
  TabsList,
  ScrollableTabsList,
  TabsTrigger,
  TabsContent,
  tabsListVariants,
}
