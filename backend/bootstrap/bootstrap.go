// Package bootstrap entscheidet beim Backend-Start, ob der Initial-Admin angelegt
// oder dessen Einmalpasswort rotiert werden muss, solange die Ersteinrichtung noch
// offen ist. Es enthält ausschließlich die Entscheidungslogik plus das Marker-Logging
// und komponiert dafür die bestehenden Domain-Pfade (user.NewUser/Activate für create,
// user.ResetPassword für rotate) — es führt keine neuen Domain-Operationen ein.
package bootstrap

import (
	"context"
	"errors"

	"github.com/rs/zerolog"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/user"
)

// AdminUsername ist der Benutzername des generierten Initial-Admins.
const AdminUsername = "admin"

// MarkerPrefix ist das feste ASCII-Literal am Anfang der maschinen-greifbaren
// Log-Zeile. Phase 3 (Windows-Starter, prod-init.sh) grept exakt diesen String —
// er darf sich nicht ändern.
const MarkerPrefix = "ADMIN-EINMALPASSWORT"

// Action ist die genau eine Aktion, die der Bootstrap aus dem DB-Zustand ableitet.
type Action string

const (
	// ActionCreate: das Backend hat einen aktiven admin mit frischem Einmalpasswort angelegt.
	ActionCreate Action = "create"
	// ActionRotate: das Einmalpasswort des bestehenden, passwortlosen admins wurde neu erzeugt.
	ActionRotate Action = "rotate"
	// ActionSkip: kein Eingriff — der Zustand erfordert keine Änderung.
	ActionSkip Action = "skip"
)

// Result ist das Ergebnis der Bootstrap-Entscheidung. OnetimePassword ist der
// Klartext-Code (leer bei ActionSkip) und wird nur in den Log-Strom geschrieben,
// nie über das Netz übertragen.
type Result struct {
	Action          Action
	OnetimePassword string
}

// Repository ist der Ausschnitt des Benutzer-Repositorys, den der Bootstrap braucht.
// Ein eigenes Interface hält das Package isoliert testbar (Fake-Repo ohne echte DB);
// user_repo.Repository erfüllt es mit der neuen CountUsers-Methode.
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
			// Der einzige Benutzer ist nicht der admin — nicht anfassen.
			return Result{Action: ActionSkip}, nil
		}
		if err != nil {
			return Result{}, err
		}

		if admin.Role == user.AdminRole && admin.PasswordHash == "" {
			// Rotations-/Wiederherstellungsfall: ResetPassword erzeugt ein frisches
			// OTP, setzt den Fehlversuchszähler zurück und heilt so auch die durch
			// eine Aussperrung geleerte OTP-Sackgasse. PasswordHash bleibt leer.
			otp, err := admin.ResetPassword()
			if err != nil {
				return Result{}, err
			}
			if err := repo.UpdateUser(ctx, admin); err != nil {
				return Result{}, err
			}
			return Result{Action: ActionRotate, OnetimePassword: otp}, nil
		}

		// Admin hat bereits ein Passwort, oder der einzige Benutzer ist kein admin.
		return Result{Action: ActionSkip}, nil
	}

	// Mehr als ein Benutzer: Einrichtung ist über den Initialzustand hinaus — nie
	// eingreifen (ein offenes Service-OTP muss unangetastet bleiben).
	return Result{Action: ActionSkip}, nil
}

// Log schreibt das Ergebnis in den Log-Strom. Bei ActionSkip passiert nichts; bei
// create/rotate werden zwei Zeilen geschrieben: eine maschinen-greifbare Markerzeile
// (fester Präfix + Klartext-Code, als MESSAGE damit sie unabhängig von der Feld-
// formatierung verbatim in der ConsoleWriter-Ausgabe steht) und eine menschenlesbare
// Bannerzeile. Der Klartext-Code geht ausschließlich in den Log-Strom.
func (r Result) Log(logger zerolog.Logger) {
	if r.Action == ActionSkip {
		return
	}

	logger.Info().Msg(MarkerPrefix + " benutzer=" + AdminUsername + " code=" + r.OnetimePassword)
	logger.Warn().Msg("Ersteinrichtung: Melde dich als Benutzer »" + AdminUsername + "« mit dem Einmalpasswort " + r.OnetimePassword + " an und lege ein eigenes Passwort fest.")
}
