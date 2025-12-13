package http

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/admin/product/application"
	"github.com/nicograef/jotti/backend/admin/product/repository"
)

func NewCommandHandler(db *sql.DB) CommandHandler {
	repo := repository.Repository{DB: db}
	command := application.Command{ProductRepo: repo}
	return CommandHandler{Command: command}
}

func NewQueryHandler(db *sql.DB) QueryHandler {
	repo := repository.Repository{DB: db}
	query := application.Query{ProductRepo: repo}
	return QueryHandler{Query: query}
}
