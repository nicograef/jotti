import { redirect } from "react-router"

import { PasswordForm } from "@/common/PasswordForm"
import { AuthSingleton } from "@/lib/auth"

export function LoadingPageLoader() {
  if (AuthSingleton.isAuthenticated && AuthSingleton.isAdmin) {
    return redirect("/admin")
  } else if (AuthSingleton.isAuthenticated) {
    return redirect("/")
  }
}
export function PasswordPage() {
  return (
    <div className="flex flex-col min-h-screen max-h-screen items-center justify-center p-4 bg-primary/5">
      <PasswordForm />
      <footer className="mt-6">
        <p className="text-muted-foreground text-sm ">
          Entwickelt von{" "}
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
