import type { ReactNode } from 'react'

// Login-Glows (Handoff Delta B): drei stark geblurte, dekorative Farbkreise
// hinter der Karte. Rein dekorativ (aria-hidden, pointer-events-none) und im
// Druck ausgeblendet. Das Clipping des Überhangs liegt in einer eigenen absolut
// positionierten Ebene (inset-0 overflow-hidden -z-10), damit der äußere
// Container normalen Überlauf behält und hohe Karten (Passwort setzen) auf
// kurzen/Landscape-Viewports scrollbar bleiben. isolate hält die Glows über dem
// Hintergrund und hinter der Karte. Die langsame Drift ist je Kreis in Dauer und
// Delay versetzt (Inline-Override der animate-drift-Dauer), damit sich die Kreise
// unabhängig bewegen — die zentrale Reduced-Motion-Regel stoppt sie.
export function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <div className="relative isolate flex flex-col min-h-screen items-center justify-start pt-16 sm:justify-center sm:pt-4 p-4 bg-primary/5">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 overflow-hidden print:hidden"
      >
        <div
          className="animate-drift absolute -top-20 -left-16 size-80 rounded-full opacity-25 blur-[60px]"
          style={{
            background:
              'radial-gradient(circle, color-mix(in oklab, var(--sp-teal) 55%, transparent), transparent 65%)',
            animationDuration: '16s',
          }}
        />
        <div
          className="animate-drift absolute -right-20 -bottom-24 size-[340px] rounded-full opacity-20 blur-[64px]"
          style={{
            background:
              'radial-gradient(circle, color-mix(in oklab, var(--sp-violet) 45%, transparent), transparent 65%)',
            animationDuration: '20s',
            animationDelay: '-5s',
          }}
        />
        <div
          className="animate-drift absolute top-1/3 -right-12 size-50 rounded-full opacity-[0.18] blur-[56px]"
          style={{
            background:
              'radial-gradient(circle, color-mix(in oklab, var(--sp-orange) 40%, transparent), transparent 65%)',
            animationDuration: '22s',
            animationDelay: '-9s',
          }}
        />
      </div>
      {children}
      <footer className="mt-6">
        <p className="text-muted-foreground text-sm">
          Entwickelt von{' '}
          <a
            href="https://nicograef.de"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:underline"
          >
            Nico Gräf
          </a>
        </p>
      </footer>
    </div>
  )
}
