package http

import (
	"context"
	"net/http"
	"time"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/domain/user"
)

type query interface {
	GetAllUsers(ctx context.Context) ([]user.User, error)
}

type QueryHandler struct {
	Query query
}

type benutzer struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type getUsersResponse struct {
	Users []benutzer `json:"users"`
}

func toUser(u user.User) benutzer {
	return benutzer{
		ID:        u.ID,
		Name:      u.Name,
		Username:  u.Username,
		Role:      string(u.Role),
		Status:    string(u.Status),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func toUsers(users []user.User) []benutzer {
	usersResponse := make([]benutzer, 0, len(users))
	for i := range users {
		usersResponse = append(usersResponse, toUser(users[i]))
	}

	return usersResponse
}

func (h *QueryHandler) GetAllUsersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := h.Query.GetAllUsers(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getUsersResponse{Users: toUsers(users)})
	}
}
