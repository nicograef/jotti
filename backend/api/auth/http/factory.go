package http

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/api/auth/application"
	"github.com/nicograef/jotti/backend/repository/user_repo"
)

func NewCommandHandler(db *sql.DB, jwtSecret string) CommandHandler {
	repo := user_repo.Repository{DB: db}
	command := application.Command{UserRepo: repo, JWTSecret: jwtSecret}
	return CommandHandler{Command: command}
}
