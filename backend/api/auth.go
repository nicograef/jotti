package api

import (
	"database/sql"
	"net/http"

	authApp "github.com/nicograef/jotti/backend/api/auth/application"
	authHTTP "github.com/nicograef/jotti/backend/api/auth/http"
	"github.com/nicograef/jotti/backend/config"
	"github.com/nicograef/jotti/backend/repository/user_repo"
)

func NewAuthApi(cfg config.Config, db *sql.DB) http.Handler {
	r := http.NewServeMux()

	userRepo := user_repo.NewRepository(db)
	ah := authHTTP.CommandHandler{}
	ah.Command = authApp.Command{UserRepo: userRepo, JWTSecret: cfg.JWTSecret}
	r.HandleFunc("/login", ah.LoginHandler())
	r.HandleFunc("/set-password", ah.SetPasswordHandler())

	return r
}
