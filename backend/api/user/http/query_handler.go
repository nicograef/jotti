package http

import (
	"context"
	"net/http"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/user/application"
	"github.com/nicograef/jotti/backend/domain/user"
)

type userReader interface {
	GetAllUsers(ctx context.Context) ([]user.User, error)
}

type QueryHandler struct {
	UserRepo userReader
}

type getUsersResponse = struct {
	Users []user.User `json:"users"`
}

func (h *QueryHandler) GetAllUsersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := application.GetAllUsers(r.Context(), h.UserRepo)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getUsersResponse{Users: users})
	}
}
