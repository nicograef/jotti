package http

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/api/user/application"
	"github.com/nicograef/jotti/backend/repository/user_repo"
)

func NewCommandHandler(db *sql.DB) CommandHandler {
	repo := user_repo.Repository{DB: db}
	command := application.Command{UserRepo: repo}
	return CommandHandler{Command: command}
}

func NewQueryHandler(db *sql.DB) QueryHandler {
	repo := user_repo.Repository{DB: db}
	query := application.Query{UserRepo: repo}
	return QueryHandler{Query: query}
}
