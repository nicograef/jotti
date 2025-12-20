package user_repo

import (
	"context"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/user"
)

func (r Repository) GetUser(ctx context.Context, id int) (user.User, error) {
	row := r.DB.QueryRowContext(ctx, "SELECT id, name, username, role, status, password_hash, onetime_password_hash, created_at FROM users WHERE id = $1 AND status != 'deleted'", id)

	var u dbuser
	err := row.Scan(&u.ID, &u.Name, &u.Username, &u.Role, &u.Status, &u.PasswordHash, &u.OnetimePasswordHash, &u.CreatedAt)

	if err != nil {
		return user.User{}, db.Error(err)
	}

	return u.toDomain(), nil
}

func (r Repository) GetUserByUsername(ctx context.Context, username string) (user.User, error) {
	row := r.DB.QueryRowContext(ctx, "SELECT id, name, username, role, status, password_hash, onetime_password_hash, created_at FROM users WHERE username = $1 AND status != 'deleted'", username)

	var u dbuser
	err := row.Scan(&u.ID, &u.Name, &u.Username, &u.Role, &u.Status, &u.PasswordHash, &u.OnetimePasswordHash, &u.CreatedAt)

	if err != nil {
		return user.User{}, db.Error(err)
	}

	return u.toDomain(), nil
}

func (r Repository) GetAllUsers(ctx context.Context) ([]user.User, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT id, name, username, role, status, created_at FROM users WHERE status != 'deleted' ORDER BY id ASC")
	if err != nil {
		return nil, db.Error(err)
	}
	defer db.Close(rows, "users")

	users := []user.User{}
	for rows.Next() {
		var u dbuser
		err := rows.Scan(&u.ID, &u.Name, &u.Username, &u.Role, &u.Status, &u.CreatedAt)
		if err != nil {
			return nil, db.Error(err)
		}
		users = append(users, u.toDomain())
	}

	if err := rows.Err(); err != nil {
		return nil, db.Error(err)
	}

	return users, nil
}

func (r Repository) CreateUser(ctx context.Context, u user.User) (int, error) {
	var userID int
	err := r.DB.QueryRowContext(ctx,
		"INSERT INTO users (name, username, role, status, password_hash, onetime_password_hash, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		u.Name, u.Username, string(u.Role), string(u.Status), u.PasswordHash, u.OnetimePasswordHash, u.CreatedAt,
	).Scan(&userID)

	if err != nil {
		return 0, db.Error(err)
	}

	return userID, nil
}

func (r Repository) UpdateUser(ctx context.Context, u user.User) error {
	result, err := r.DB.ExecContext(ctx,
		"UPDATE users SET name = $1, username = $2, role = $3, status = $4, password_hash = $5, onetime_password_hash = $6 WHERE id = $7",
		u.Name, u.Username, string(u.Role), string(u.Status), u.PasswordHash, u.OnetimePasswordHash, u.ID,
	)
	if err != nil {
		return db.Error(err)
	}

	return db.ResultError(result)
}
