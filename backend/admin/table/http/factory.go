package http

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/admin/table/application"
	"github.com/nicograef/jotti/backend/admin/table/repository"
)

func NewCommandHandler(db *sql.DB) CommandHandler {
	repo := repository.Repository{DB: db}
	command := application.Command{TableRepo: repo}
	return CommandHandler{Command: command}
}

func NewQueryHandler(db *sql.DB) QueryHandler {
	repo := repository.Repository{DB: db}
	query := application.Query{TableRepo: repo}
	return QueryHandler{Query: query}
}
