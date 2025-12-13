package api

import (
	"database/sql"
	"net/http"

	"github.com/nicograef/jotti/backend/auth/auth"
	"github.com/nicograef/jotti/backend/config"
)

func NewApi(cfg config.Config, db *sql.DB) http.Handler {
	r := http.NewServeMux()

	ah := auth.AuthHandler{JWTSecret: cfg.JWTSecret, Command: &auth.Command{Persistence: &auth.Persistence{DB: db}}}
	r.HandleFunc("/login", ah.LoginHandler())
	r.HandleFunc("/set-password", ah.SetPasswordHandler())

	return r
}
