package http

import (
	"context"
	"net/http"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/domain/user"
)

type query interface {
	GetAllUsers(ctx context.Context) ([]user.User, error)
}

type QueryHandler struct {
	Query query
}

type getUsersResponse = struct {
	Users []user.User `json:"users"`
}

func (h *QueryHandler) GetAllUsersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := h.Query.GetAllUsers(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getUsersResponse{Users: users})
	}
}
