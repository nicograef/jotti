import { useEffect, useRef, useState } from 'react'
import { Menu, X } from 'lucide-react'

// Mobile Navigation als Burger-Menü unter 860px (Sichtbarkeit per CSS: der
// Wrapper trägt `nav:hidden`, ist also nur unterhalb des Breakpoints da).
// Interaktionsmuster wie das frühere public/mobile-nav.js: `aria-expanded`,
// Escape schließt und gibt den Fokus zurück, Klick auf einen Link schließt.

interface NavLink {
  href: string
  label: string
}

export default function MobileNav({ links }: { links: NavLink[] }) {
  const [open, setOpen] = useState(false)
  const buttonRef = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    if (!open) return
    function onKeydown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        setOpen(false)
        buttonRef.current?.focus()
      }
    }
    document.addEventListener('keydown', onKeydown)
    return () => document.removeEventListener('keydown', onKeydown)
  }, [open])

  return (
    <div className="nav:hidden">
      <button
        ref={buttonRef}
        type="button"
        aria-expanded={open}
        aria-controls="mobile-nav-panel"
        aria-label={open ? 'Menü schließen' : 'Menü öffnen'}
        onClick={() => setOpen((value) => !value)}
        className="flex h-10 w-10 items-center justify-center text-foreground"
      >
        {open ? (
          <X size={26} aria-hidden="true" />
        ) : (
          <Menu size={26} aria-hidden="true" />
        )}
      </button>

      {open && (
        <nav
          id="mobile-nav-panel"
          aria-label="Mobile Navigation"
          className="absolute inset-x-0 top-full flex flex-col gap-1 border-t border-card-border bg-background px-6 py-4 shadow-lg"
        >
          {links.map((link) => (
            <a
              key={link.href}
              href={link.href}
              onClick={() => setOpen(false)}
              className="rounded-lg px-3 py-2.5 font-semibold hover:bg-surface-alt"
            >
              {link.label}
            </a>
          ))}
        </nav>
      )}
    </div>
  )
}
