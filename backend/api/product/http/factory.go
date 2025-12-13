package http

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/api/product/application"
	"github.com/nicograef/jotti/backend/repository/product_repo"
)

func NewCommandHandler(db *sql.DB) CommandHandler {
	repo := product_repo.Repository{DB: db}
	command := application.Command{ProductRepo: repo}
	return CommandHandler{Command: command}
}

func NewQueryHandler(db *sql.DB) QueryHandler {
	repo := product_repo.Repository{DB: db}
	query := application.Query{ProductRepo: repo}
	return QueryHandler{Query: query}
}
