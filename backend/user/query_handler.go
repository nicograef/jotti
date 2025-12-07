package user

import (
	"context"
	"net/http"

	"github.com/nicograef/jotti/backend/api"
)

type query interface {
	GetAllUsers(ctx context.Context) ([]User, error)
}

type QueryHandler struct {
	Query query
}

type getUsersResponse = struct {
	Users []User `json:"users"`
}

// GetAllUsersHandler handles requests to retrieve all users.
func (h *QueryHandler) GetAllUsersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		users, err := h.Query.GetAllUsers(ctx)
		if err != nil {
			api.SendServerError(w)
			return
		}

		api.SendResponse(w, getUsersResponse{Users: users})
	}
}
