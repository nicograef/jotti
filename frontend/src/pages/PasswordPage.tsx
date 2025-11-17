import { PasswordForm } from '@/components/common/PasswordForm'
import { AuthBackend } from '@/lib/AuthBackend'
import { BackendSingleton } from '@/lib/Backend'

const authBackend = new AuthBackend(BackendSingleton)

export function PasswordPage() {
  return (
    <div className="flex flex-col min-h-screen max-h-screen items-center justify-center p-4 bg-primary/5">
      <PasswordForm backend={authBackend} />
      <footer className="mt-6">
        <p className="text-muted-foreground text-sm ">
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
