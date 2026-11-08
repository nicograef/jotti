package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/nicograef/jotti/backend/bootstrap"
	"github.com/nicograef/jotti/backend/config"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/user_repo"
)

// Serve baut die Projektionen neu auf, legt bei Bedarf den Initial-Admin an und bedient die API, bis ctx endet.
// Der Aufrufer besitzt db und schließt sie danach.
func Serve(ctx context.Context, cfg config.Config, db *sql.DB, version string) error {
	// A migration may empty tisch_sessions (03 does); without the rebuild a table would lose its open positions.
	count, err := kassenjournal_repo.NewRepository(db).RebuildAllProjections(ctx)
	if err != nil {
		return fmt.Errorf("rebuild projections: %w", err)
	}
	slog.Info("Projections rebuilt", "subjects", count)

	// Der Klartext-Code des Einmalpassworts landet im Log-Strom.
	res, err := bootstrap.EnsureInitialAdmin(ctx, user_repo.NewRepository(db))
	if err != nil {
		return fmt.Errorf("bootstrap initial admin: %w", err)
	}
	res.Log(slog.Default())

	return NewApp(cfg, db, version).Run(ctx)
}
