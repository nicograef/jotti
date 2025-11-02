import { Controller, useForm } from "react-hook-form"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Separator } from "@/components/ui/separator"
import { UserPlus } from "lucide-react"

interface FormData {
  name: string
  username: string
  role: "admin" | "service"
}

function toUsername(name: string) {
  return name
    .toLowerCase()
    .replace(/\s+/g, "")
    .replace(/ä/g, "ae")
    .replace(/ö/g, "oe")
    .replace(/ü/g, "ue")
    .replace(/ß/g, "ss")
    .replace(/[^a-z0-9]/g, "")
}

export function NewUserDialog() {
  const form = useForm<FormData>({
    defaultValues: { name: "", username: "", role: "service" },
  })

  const onSubmit = (data: FormData) => {
    console.log("User:", data)
  }

  return (
    <Dialog>
      <DialogTrigger asChild>
        <div className="fixed bottom-16 right-16 z-50">
          <Button>
            <UserPlus /> Neuer Benutzer
          </Button>
        </div>
      </DialogTrigger>
      <DialogContent
        className="sm:max-w-[425px]"
        aria-description="Dialog, um einen neuen Benutzer anzulegen."
      >
        <DialogHeader>
          <DialogTitle>Neuen Benutzer anlegen</DialogTitle>
        </DialogHeader>
        <Separator />
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
              rules={{
                required: "Name ist ein Pflichtfeld.",
                minLength: {
                  value: 5,
                  message: "Name muss mindestens 5 Zeichen lang sein.",
                },
                maxLength: {
                  value: 50,
                  message: "Name darf maximal 50 Zeichen lang sein.",
                },
              }}
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
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
                    placeholder="Nico Gräf"
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
              rules={{
                required: "Benutzername ist ein Pflichtfeld.",
                minLength: {
                  value: 5,
                  message: "Benutzername muss mindestens 5 Zeichen lang sein.",
                },
                maxLength: {
                  value: 20,
                  message: "Benutzername darf maximal 20 Zeichen lang sein.",
                },
              }}
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
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
                    placeholder="nicograef"
                    autoComplete="off"
                  />
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
            >
              Abbrechen
            </Button>
          </DialogClose>
          <Button type="submit" form="user-form">
            Benutzer anlegen
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
