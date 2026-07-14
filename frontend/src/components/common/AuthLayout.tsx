import type { ReactNode } from 'react'

// Login-Glows (Handoff Delta B): drei stark geblurte, dekorative Farbkreise
// hinter der Karte. Rein dekorativ (aria-hidden, pointer-events-none) und im
// Druck ausgeblendet; overflow-hidden am Wrapper clippt den Überhang, isolate
// hält sie über dem Hintergrund und hinter der Karte. Die langsame Drift ist je
// Kreis in Dauer und Delay versetzt (Inline-Override der animate-drift-Dauer),
// damit sich die Kreise unabhängig bewegen — die zentrale Reduced-Motion-Regel
// stoppt sie.
export function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <div className="relative isolate flex flex-col min-h-screen max-h-screen items-center justify-start overflow-hidden pt-16 sm:justify-center sm:pt-4 p-4 bg-primary/5">
      <div
        aria-hidden
        className="animate-drift pointer-events-none absolute -top-20 -left-16 -z-10 size-80 rounded-full opacity-25 blur-[60px] print:hidden"
        style={{
          background:
            'radial-gradient(circle, color-mix(in oklab, var(--sp-teal) 55%, transparent), transparent 65%)',
          animationDuration: '16s',
        }}
      />
      <div
        aria-hidden
        className="animate-drift pointer-events-none absolute -right-20 -bottom-24 -z-10 size-[340px] rounded-full opacity-20 blur-[64px] print:hidden"
        style={{
          background:
            'radial-gradient(circle, color-mix(in oklab, var(--sp-violet) 45%, transparent), transparent 65%)',
          animationDuration: '20s',
          animationDelay: '-5s',
        }}
      />
      <div
        aria-hidden
        className="animate-drift pointer-events-none absolute top-1/3 -right-12 -z-10 size-50 rounded-full opacity-[0.18] blur-[56px] print:hidden"
        style={{
          background:
            'radial-gradient(circle, color-mix(in oklab, var(--sp-orange) 40%, transparent), transparent 65%)',
          animationDuration: '22s',
          animationDelay: '-9s',
        }}
      />
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
