package http

import (
	"context"
	"net/http"
	"time"

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

type userDTO struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type getUsersResponse = struct {
	Users []userDTO `json:"users"`
}

func toUserDTO(u user.User) userDTO {
	return userDTO{
		ID:        u.ID,
		Name:      u.Name,
		Username:  u.Username,
		Role:      string(u.Role),
		Status:    string(u.Status),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func toUserDTOs(users []user.User) []userDTO {
	userDTOs := make([]userDTO, 0, len(users))
	for _, u := range users {
		userDTOs = append(userDTOs, toUserDTO(u))
	}

	return userDTOs
}

func (h *QueryHandler) GetAllUsersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := application.GetAllUsers(r.Context(), h.UserRepo)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getUsersResponse{Users: toUserDTOs(users)})
	}
}
