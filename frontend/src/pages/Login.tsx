import React from "react"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"

export default function Login() {
  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()

    const form = new FormData(e.currentTarget)
    const username = form.get("username") ?? ""
    const password = form.get("password") ?? ""

    if (typeof username !== "string" || typeof password !== "string") {
      throw new Error("Invalid form data")
    }

    console.log({ username, password })
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
          />
        </div>
        <Button type="submit" className="w-full">
          Anmelden
        </Button>
      </form>
    </div>
  )
}
