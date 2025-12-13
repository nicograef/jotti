package http

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/api/table/application"
	"github.com/nicograef/jotti/backend/repository/table_repo"
)

func NewCommandHandler(db *sql.DB) CommandHandler {
	repo := table_repo.Repository{DB: db}
	command := application.Command{TableRepo: repo}
	return CommandHandler{Command: command}
}

func NewQueryHandler(db *sql.DB) QueryHandler {
	repo := table_repo.Repository{DB: db}
	query := application.Query{TableRepo: repo}
	return QueryHandler{Query: query}
}
