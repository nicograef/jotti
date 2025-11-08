import { Controller, useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { toUsername, UserSchema, type User } from "@/lib/user"
import { useState } from "react"
import { BackendSingleton } from "@/lib/backend"
import { z } from "zod"
import { DialogDescription } from "@radix-ui/react-dialog"
import { Switch } from "@/components/ui/switch"

const FormDataSchema = UserSchema.pick({
  name: true,
  username: true,
  role: true,
  locked: true,
})
type FormData = z.infer<typeof FormDataSchema>

interface NewUserDialogProps {
  open: boolean
  user: User
  updated: (user: User) => void
  close: () => void
}

export function EditUserDialog(props: Readonly<NewUserDialogProps>) {
  const [loading, setLoading] = useState(false)
  const form = useForm<FormData>({
    defaultValues: props.user,
    resolver: zodResolver(FormDataSchema),
    mode: "onBlur",
  })

  const onOpenChange = (isOpen: boolean) => {
    if (!isOpen) {
      form.reset()
      props.close()
    }
  }

  const onSubmit = async (data: FormData) => {
    setLoading(true)

    try {
      const updatedUser = await BackendSingleton.updateUser({
        id: props.user.id,
        ...data,
      })
      form.reset()
      props.updated(updatedUser)
      props.close()
    } catch (error: unknown) {
      console.error(error)
    }

    setLoading(false)
  }

  return (
    <Dialog open={props.open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader className="mb-4">
          <DialogTitle>{props.user.name}</DialogTitle>
          <DialogDescription>
            Du kannst Name, Benutzername, Rolle und Status des Benutzers ändern.
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
            <Controller
              name="locked"
              control={form.control}
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid} className="gap-1">
                  <FieldLabel htmlFor="user-form-locked">Sperren?</FieldLabel>
                  <FieldContent className="flex flex-row items-center">
                    <Switch
                      id="user-form-locked"
                      aria-invalid={fieldState.invalid}
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                    {field.value && (
                      <FieldDescription className="ml-4">
                        Wenn du diesen Benutzer sperrst, kann er sich nicht mehr
                        anmelden.
                      </FieldDescription>
                    )}
                  </FieldContent>
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
            Speichern
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
