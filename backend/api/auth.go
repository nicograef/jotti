package api

import (
	"net/http"

	authApp "github.com/nicograef/jotti/backend/api/auth/application"
	authHTTP "github.com/nicograef/jotti/backend/api/auth/http"
	"github.com/nicograef/jotti/backend/api/auth/throttle"
	"github.com/nicograef/jotti/backend/config"
)

func NewAuthApi(cfg config.Config, deps Deps) http.Handler {
	r := http.NewServeMux()

	ah := authHTTP.CommandHandler{}
	ah.Command = authApp.Command{UserRepo: deps.UserRepo, JWTSecret: cfg.JWTSecret, Throttle: throttle.NewLoginThrottle()}
	r.HandleFunc("/login", ah.LoginHandler())
	r.HandleFunc("/set-password", ah.SetPasswordHandler())

	return r
}
