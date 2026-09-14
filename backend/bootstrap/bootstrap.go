// Package bootstrap entscheidet beim Backend-Start, ob der Initial-Admin angelegt
// oder dessen Einmalpasswort rotiert werden muss.
package bootstrap

import (
	"context"
	"errors"

	"github.com/rs/zerolog"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/user"
)

const AdminUsername = "admin"

// MarkerPrefix ist das feste ASCII-Literal am Anfang der maschinen-greifbaren
// Log-Zeile. windows/starter/core/adminmarker.go und scripts/prod-init.sh greifen
// exakt diesen String — er darf sich nicht ändern.
const MarkerPrefix = "ADMIN-EINMALPASSWORT"

type Action string

const (
	ActionCreate Action = "create"
	ActionRotate Action = "rotate"
	ActionSkip   Action = "skip"
)

// Result trägt die Bootstrap-Entscheidung. OnetimePassword ist der Klartext-Code
// (leer bei ActionSkip) und geht nur in den Log-Strom, nie über das Netz.
type Result struct {
	Action          Action
	OnetimePassword string
}

type Repository interface {
	CountUsers(ctx context.Context) (int, error)
	GetUserByUsername(ctx context.Context, username string) (user.User, error)
	CreateUser(ctx context.Context, u user.User) (int, error)
	UpdateUser(ctx context.Context, u user.User) error
}

// EnsureInitialAdmin entscheidet aus dem DB-Zustand genau eine Aktion:
//   - leeres Repo → create: aktiver admin mit frischem 6-Ziffern-OTP, kein Passwort.
//   - genau ein Benutzer, dieser ist admin ohne Passwort → rotate: neues OTP, Zähler 0.
//   - jeder andere Zustand → skip: keine Änderung (offenes Service-OTP nie antasten).
func EnsureInitialAdmin(ctx context.Context, repo Repository) (Result, error) {
	count, err := repo.CountUsers(ctx)
	if err != nil {
		return Result{}, err
	}

	if count == 0 {
		u, otp, err := user.NewUser("Administrator", AdminUsername, user.AdminRole)
		if err != nil {
			return Result{}, err
		}
		u.Activate()
		if _, err := repo.CreateUser(ctx, u); err != nil {
			return Result{}, err
		}
		return Result{Action: ActionCreate, OnetimePassword: otp}, nil
	}

	if count == 1 {
		admin, err := repo.GetUserByUsername(ctx, AdminUsername)
		if errors.Is(err, db.ErrNotFound) {
			return Result{Action: ActionSkip}, nil
		}
		if err != nil {
			return Result{}, err
		}

		if admin.Role == user.AdminRole && admin.PasswordHash == "" {
			// ResetPassword heilt auch die durch eine Aussperrung geleerte
			// OTP-Sackgasse (frisches OTP, Fehlversuchszähler zurück).
			otp, err := admin.ResetPassword()
			if err != nil {
				return Result{}, err
			}
			if err := repo.UpdateUser(ctx, admin); err != nil {
				return Result{}, err
			}
			return Result{Action: ActionRotate, OnetimePassword: otp}, nil
		}

		return Result{Action: ActionSkip}, nil
	}

	return Result{Action: ActionSkip}, nil
}

// Log schreibt bei create/rotate die maschinen-greifbare Markerzeile als MESSAGE
// — nur so steht sie unabhängig von der Feldformatierung verbatim in der
// ConsoleWriter-Ausgabe — plus eine menschenlesbare Bannerzeile.
func (r Result) Log(logger zerolog.Logger) {
	if r.Action == ActionSkip {
		return
	}

	logger.Info().Msg(MarkerPrefix + " benutzer=" + AdminUsername + " code=" + r.OnetimePassword)
	logger.Warn().Msg("Ersteinrichtung: Melde dich als Benutzer »" + AdminUsername + "« mit dem Einmalpasswort " + r.OnetimePassword + " an und lege ein eigenes Passwort fest.")
}
