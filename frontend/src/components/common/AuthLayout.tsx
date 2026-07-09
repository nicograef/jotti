import type { ReactNode } from 'react'

export function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <div className="flex flex-col min-h-screen max-h-screen items-center justify-start pt-16 sm:justify-center sm:pt-4 p-4 bg-primary/5">
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
