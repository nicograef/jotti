import { AuthLayout } from '@/components/common/AuthLayout'
import { PasswordForm } from '@/components/common/PasswordForm'
import { AuthBackend } from '@/lib/AuthBackend'
import { BackendSingleton } from '@/lib/Backend'

const authBackend = new AuthBackend(BackendSingleton)

export function PasswordPage() {
  return (
    <AuthLayout>
      <PasswordForm backend={authBackend} />
    </AuthLayout>
  )
}
