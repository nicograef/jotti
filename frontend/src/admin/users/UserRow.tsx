import { KeyRound, MoreHorizontal, Pen, Trash2 } from 'lucide-react'
import { useState } from 'react'

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Switch } from '@/components/ui/switch'

import { type User, UserStatus } from './User'
import { RolleBadge } from './UserRolle'
import { BENUTZER_SPALTEN } from './UsersSpalten'

interface UserRowProps {
  loading: boolean
  user: User
  isSelf: boolean
  onEdit: (userId: number) => void
  onActivate: (userId: number) => Promise<void>
  onDeactivate: (userId: number) => Promise<void>
  onResetPassword: (userId: number) => Promise<void>
  onDelete: (userId: number) => Promise<void>
}

// Eine Zeile der Benutzertabelle (Design-Handoff 1e): Name mit Login (und beim
// eigenen Konto der „das bist du“-Badge), Rollen-Badge, Status-Switch und die
// Aktionen (Bearbeiten plus „···“-Menü mit Passwort-Zurücksetzen und Löschen).
// Am eigenen Konto wird kein Löschen angeboten; der Backend-Schutz
// `cannot_delete_self` bleibt die zweite Verteidigung.
export function UserRow(props: UserRowProps) {
  const [deleteOpen, setDeleteOpen] = useState(false)
  const isActive = props.user.status === UserStatus.ACTIVE

  return (
    <div
      className={`${BENUTZER_SPALTEN} gap-y-1 border-t px-4 py-3 text-sm first:border-t-0`}
    >
      <span className="font-medium">
        {props.user.name}{' '}
        <span className="font-mono text-xs text-muted-foreground">
          {props.user.username}
        </span>
        {props.isSelf && (
          <Badge variant="secondary" className="ml-1.5 align-middle">
            das bist du
          </Badge>
        )}
      </span>

      <span>
        <RolleBadge role={props.user.role} />
      </span>

      <span className="flex items-center gap-2 text-xs text-muted-foreground">
        <Switch
          className="cursor-pointer"
          aria-label={isActive ? 'Helfer deaktivieren' : 'Helfer aktivieren'}
          disabled={props.loading}
          checked={isActive}
          onCheckedChange={(checked) => {
            if (checked) {
              void props.onActivate(props.user.id)
            } else {
              void props.onDeactivate(props.user.id)
            }
          }}
        />
        {isActive ? 'aktiv' : 'deaktiviert'}
      </span>

      <span className="flex items-center justify-end gap-1">
        <Button
          size="icon-sm"
          variant="ghost"
          className="cursor-pointer rounded-full"
          aria-label="Helfer bearbeiten"
          onClick={() => {
            props.onEdit(props.user.id)
          }}
        >
          <Pen />
        </Button>

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              size="icon-sm"
              variant="ghost"
              className="cursor-pointer rounded-full"
              aria-label="Weitere Aktionen"
            >
              <MoreHorizontal />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem
              disabled={props.loading}
              onSelect={() => void props.onResetPassword(props.user.id)}
            >
              <KeyRound /> Passwort zurücksetzen
            </DropdownMenuItem>
            {!props.isSelf && (
              <>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  variant="destructive"
                  onSelect={() => {
                    setDeleteOpen(true)
                  }}
                >
                  <Trash2 /> Löschen…
                </DropdownMenuItem>
              </>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </span>

      <AlertDialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Helfer löschen?</AlertDialogTitle>
            <AlertDialogDescription>
              Der Helfer &ldquo;{props.user.name}&rdquo; wird unwiderruflich
              gelöscht.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Abbrechen</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive-solid"
              onClick={(e) => {
                e.preventDefault()
                void props.onDelete(props.user.id).then(() => {
                  setDeleteOpen(false)
                })
              }}
              disabled={props.loading}
            >
              Löschen
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
