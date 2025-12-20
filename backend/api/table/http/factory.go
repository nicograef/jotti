package http

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/api/table/application"
	"github.com/nicograef/jotti/backend/repository/event_repo"
	"github.com/nicograef/jotti/backend/repository/table_repo"
)

func NewCommandHandler(db *sql.DB) CommandHandler {
	tableRepo := table_repo.Repository{DB: db}
	eventRepo := event_repo.Repository{DB: db}
	command := application.Command{TableRepo: tableRepo, EventRepo: eventRepo}
	return CommandHandler{Command: command}
}

func NewQueryHandler(db *sql.DB) QueryHandler {
	tableRepo := table_repo.Repository{DB: db}
	eventRepo := event_repo.Repository{DB: db}
	query := application.Query{TableRepo: tableRepo, EventRepo: eventRepo}
	return QueryHandler{Query: query}
}
