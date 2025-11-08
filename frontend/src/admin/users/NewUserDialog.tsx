import { Controller, useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { UserPlus } from "lucide-react"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { toUsername, UserRole, UserSchema, type User } from "@/lib/user"
import { useState } from "react"
import { BackendSingleton } from "@/lib/backend"
import { z } from "zod"

const FormDataSchema = UserSchema.pick({
  name: true,
  username: true,
  role: true,
})
type FormData = z.infer<typeof FormDataSchema>

interface NewUserDialogProps {
  created: (user: User, onetimePassword: string) => void
}

export function NewUserDialog({ created }: NewUserDialogProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const form = useForm<FormData>({
    defaultValues: { name: "", username: "", role: UserRole.SERVICE },
    resolver: zodResolver(FormDataSchema),
    mode: "onBlur",
  })

  const onSubmit = async (data: FormData) => {
    setLoading(true)

    try {
      const response = await BackendSingleton.createUser(
        data.name,
        data.username,
        data.role as UserRole,
      )
      form.reset()
      setOpen(false)
      created(response.user, response.onetimePassword)
    } catch (error: unknown) {
      console.error(error)
    }

    setLoading(false)
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <div className="fixed bottom-16 right-16 z-50">
          <Button className="cursor-pointer hover:shadow-sm">
            <UserPlus /> Neuer Benutzer
          </Button>
        </div>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader className="mb-4">
          <DialogTitle>Neuen Benutzer anlegen</DialogTitle>
          <DialogDescription>
            Das Passwort kann der Benutzer später selbst festlegen.
          </DialogDescription>
        </DialogHeader>
        <form
          id="user-form"
          onSubmit={(e) => {
            e.preventDefault()
            void form.handleSubmit(onSubmit)()
            return false
          }}
        >
          <FieldGroup>
            <Controller
              name="name"
              control={form.control}
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid} className="gap-1">
                  <FieldLabel htmlFor="user-form-name">Name</FieldLabel>
                  <Input
                    {...field}
                    onBlur={() => {
                      if (form.getValues("username").length === 0) {
                        const username = toUsername(field.value)
                        form.setValue("username", username, {
                          shouldDirty: true,
                          shouldValidate: true,
                        })
                      }
                      field.onBlur()
                    }}
                    id="user-form-name"
                    aria-invalid={fieldState.invalid}
                    placeholder="Vor- und Nachname eingeben"
                    autoComplete="off"
                  />
                  {fieldState.invalid && (
                    <FieldError errors={[fieldState.error]} />
                  )}
                </Field>
              )}
            />
            <Controller
              name="username"
              control={form.control}
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid} className="gap-1">
                  <FieldLabel htmlFor="user-form-username">
                    Benutzername
                  </FieldLabel>
                  <Input
                    {...field}
                    onChange={(e) => {
                      const username = toUsername(e.target.value)
                      field.onChange(username)
                    }}
                    id="user-form-username"
                    aria-invalid={fieldState.invalid}
                    placeholder=""
                    autoComplete="off"
                  />
                  {fieldState.invalid && (
                    <FieldError errors={[fieldState.error]} />
                  )}
                </Field>
              )}
            />
            <Controller
              name="role"
              control={form.control}
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid} className="gap-1">
                  <FieldLabel htmlFor="user-form-role">Rolle</FieldLabel>
                  {field.value === "admin" && (
                    <FieldDescription>
                      Administratoren können alle Funktionen nutzen.
                    </FieldDescription>
                  )}
                  {field.value === "service" && (
                    <FieldDescription>
                      Service kann Bestellungen und Bezahlungen verwalten.
                    </FieldDescription>
                  )}
                  <Select
                    name={field.name}
                    value={field.value}
                    onValueChange={field.onChange}
                  >
                    <SelectTrigger
                      id="user-form-role"
                      aria-invalid={fieldState.invalid}
                    >
                      <SelectValue placeholder="Auswählen" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="admin">Administrator</SelectItem>
                      <SelectItem value="service">Service</SelectItem>
                    </SelectContent>
                  </Select>
                  {fieldState.invalid && (
                    <FieldError errors={[fieldState.error]} />
                  )}
                </Field>
              )}
            />
          </FieldGroup>
        </form>
        <DialogFooter className="mt-4">
          <DialogClose asChild>
            <Button
              variant="outline"
              onClick={() => {
                form.reset()
              }}
              disabled={loading}
            >
              Abbrechen
            </Button>
          </DialogClose>
          <Button type="submit" form="user-form" disabled={loading}>
            Benutzer anlegen
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
