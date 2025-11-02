import { Login } from "@/common/Login"
import { AuthSingleton } from "@/lib/auth"

import { redirect } from "react-router"

export function LoadingPageLoader() {
  console.log("LoadingPageLoader called")
  if (AuthSingleton.isAuthenticated && AuthSingleton.isAdmin) {
    return redirect("/admin")
  } else if (AuthSingleton.isAuthenticated) {
    return redirect("/")
  }
}
export function LoginPage() {
  return <Login />
}
