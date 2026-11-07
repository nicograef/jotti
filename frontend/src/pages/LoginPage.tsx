import { AuthLayout } from '@/components/common/AuthLayout'
import { LoginForm } from '@/components/common/LoginForm'

export function LoginPage() {
  return (
    <AuthLayout>
      <LoginForm />
    </AuthLayout>
  )
}
