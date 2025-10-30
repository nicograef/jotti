import React from "react"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Spinner } from "@/components/ui/spinner"
import { backend } from "@/lib/backend"

export default function Login() {
  const [loading, setLoading] = React.useState(false)

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setLoading(true)

    const form = new FormData(e.currentTarget)
    const username = form.get("username") ?? ""
    const password = form.get("password") ?? ""

    if (typeof username !== "string" || typeof password !== "string") {
      setLoading(false)
      throw new Error("Invalid form data")
    }

    backend
      .login(username, password)
      .then((token) => {
        console.log("Logged in with token:", token)
        localStorage.setItem("JOTTI_TOKEN", token)
      })
      .catch((error: unknown) => {
        console.error("Login failed:", error)
      })
      .finally(() => {
        setLoading(false)
      })
  }

  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <form onSubmit={handleSubmit} className="w-full max-w-sm space-y-4">
        <div className="space-y-2">
          <Label htmlFor="username">Benutzername</Label>
          <Input
            id="username"
            name="username"
            autoComplete="username"
            required
            disabled={loading}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="password">Passwort</Label>
          <Input
            id="password"
            name="password"
            type="password"
            autoComplete="current-password"
            required
            disabled={loading}
          />
        </div>
        <Button type="submit" className="w-full" disabled={loading}>
          {loading ? <Spinner /> : <></>} Anmelden
        </Button>
      </form>
    </div>
  )
}
