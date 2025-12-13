package api

import (
	"database/sql"
	"net/http"

	auth "github.com/nicograef/jotti/backend/api/auth/http"
	"github.com/nicograef/jotti/backend/config"
)

func NewAuthApi(cfg config.Config, db *sql.DB) http.Handler {
	r := http.NewServeMux()

	ah := auth.NewCommandHandler(db, cfg.JWTSecret)
	r.HandleFunc("/login", ah.LoginHandler())
	r.HandleFunc("/set-password", ah.SetPasswordHandler())

	return r
}
