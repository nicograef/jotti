import { AuthLayout } from '@/components/common/AuthLayout'
import { LoginForm } from '@/components/common/LoginForm'
import { AuthBackend } from '@/lib/AuthBackend'
import { BackendSingleton } from '@/lib/Backend'

const authBackend = new AuthBackend(BackendSingleton)

export function LoginPage() {
  return (
    <AuthLayout>
      <LoginForm backend={authBackend} />
    </AuthLayout>
  )
}
