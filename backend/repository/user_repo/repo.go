package user_repo

import (
	"context"
	"database/sql"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/user"
	"github.com/nicograef/jotti/backend/sqlc/dbgen"
)

func (r Repository) GetUser(ctx context.Context, id int) (user.User, error) {
	row, err := r.q.GetUser(ctx, id)
	if err != nil {
		return user.User{}, db.Error(err)
	}

	return userRowToDomain(row), nil
}

func (r Repository) GetUserByUsername(ctx context.Context, username string) (user.User, error) {
	row, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		return user.User{}, db.Error(err)
	}

	return userByUsernameRowToDomain(row), nil
}

func (r Repository) GetAllUsers(ctx context.Context) ([]user.User, error) {
	rows, err := r.q.GetAllUsers(ctx)
	if err != nil {
		return nil, db.Error(err)
	}

	users := make([]user.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, allUsersRowToDomain(row))
	}

	return users, nil
}

// CountUsers zählt ALLE Benutzerzeilen inklusive soft-gelöschter (status = 'deleted'):
// Benutzernamen werden nie recycelt, deshalb blockiert ein soft-gelöschter "admin"
// die Neuanlage. Der Bootstrap-Entscheider verlässt sich auf diese Vollzählung.
func (r Repository) CountUsers(ctx context.Context) (int, error) {
	count, err := r.q.CountUsers(ctx)
	if err != nil {
		return 0, db.Error(err)
	}

	return int(count), nil
}

func (r Repository) CreateUser(ctx context.Context, u user.User) (int, error) {
	userID, err := r.q.CreateUser(ctx, dbgen.CreateUserParams{
		Name:                u.Name,
		Username:            u.Username,
		Role:                dbgen.Userrole(u.Role),
		Status:              dbgen.Entitystatus(u.Status),
		PasswordHash:        sql.NullString{String: u.PasswordHash, Valid: u.PasswordHash != ""},
		OnetimePasswordHash: sql.NullString{String: u.OnetimePasswordHash, Valid: u.OnetimePasswordHash != ""},
		CreatedAt:           u.CreatedAt,
		UpdatedAt:           u.UpdatedAt,
	})
	if err != nil {
		return 0, db.Error(err)
	}

	return userID, nil
}

func (r Repository) UpdateUser(ctx context.Context, u user.User) error {
	result, err := r.q.UpdateUser(ctx, updateUserParams(u))
	if err != nil {
		return db.Error(err)
	}

	return db.ResultError(result)
}

func updateUserParams(u user.User) dbgen.UpdateUserParams {
	return dbgen.UpdateUserParams{
		Name:                    u.Name,
		Username:                u.Username,
		Role:                    dbgen.Userrole(u.Role),
		Status:                  dbgen.Entitystatus(u.Status),
		PasswordHash:            sql.NullString{String: u.PasswordHash, Valid: u.PasswordHash != ""},
		OnetimePasswordHash:     sql.NullString{String: u.OnetimePasswordHash, Valid: u.OnetimePasswordHash != ""},
		OnetimePasswordAttempts: u.OnetimePasswordAttempts,
		UpdatedAt:               u.UpdatedAt,
		ID:                      u.ID,
	}
}

// SetPasswordTx lädt den Benutzer mit Zeilensperre (FOR UPDATE), führt apply aus
// und persistiert das Ergebnis in EINER Transaktion — nach dem Callback-in-TX-
// Muster von kassenjournal_repo.EroeffneKassensitzung. Damit werden konkurrierende
// Set-Password-Versuche für denselben Benutzer serialisiert: der zweite Versuch
// wartet auf der Zeilensperre, bis der erste committet hat, und liest dann den
// bereits erhöhten Fehlversuchszähler — der Zähler kann nicht unterzählen.
//
// apply mutiert den Benutzer (SetPassword zählt den Fehlversuchszähler hoch bzw.
// setzt das neue Passwort). Der von apply gelieferte Fachfehler (falsches
// Einmalpasswort, Sperre, ...) wird NICHT als Transaktionsabbruch gewertet: der
// mutierte Zustand wird trotzdem im selben Commit persistiert und der Fehler
// danach an den Aufrufer zurückgegeben. Nur echte DB-Fehler brechen ab (Rollback).
func (r Repository) SetPasswordTx(ctx context.Context, username string, apply func(*user.User) error) error {
	var applyErr error
	txErr := r.withTx(ctx, func(qtx *dbgen.Queries) error {
		row, err := qtx.GetUserByUsernameForUpdate(ctx, username)
		if err != nil {
			return db.Error(err)
		}

		u := userByUsernameForUpdateRowToDomain(row)
		applyErr = apply(&u)

		if _, err := qtx.UpdateUser(ctx, updateUserParams(u)); err != nil {
			return db.Error(err)
		}
		return nil
	})
	if txErr != nil {
		return txErr
	}
	return applyErr
}

// withTx runs fn within a single transaction: it begins the tx, rolls back on any
// error (a rollback after commit is a no-op), and commits otherwise. fn receives
// the transaction-bound queries; only begin/commit failures are normalized here.
func (r Repository) withTx(ctx context.Context, fn func(*dbgen.Queries) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return db.Error(err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	if err := fn(r.q.WithTx(tx)); err != nil {
		return err
	}

	return db.Error(tx.Commit())
}
