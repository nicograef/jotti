package api

import (
	"database/sql"
	"net/http"
)

type CreateUserBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateUserResponse struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

type CreateUserErrorResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

// NewCreateUserHandler creates new users in the database.
func NewCreateUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateMethod(w, r, http.MethodPost) {
			return
		}

		body := CreateUserBody{}
		if !readJSONRequest(w, r, &body) {
			return
		}

		// Hash the password securely
		hashedPassword, err := hashPasswordSecure(body.Password)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}

		// Create user in the database
		if err := createUser(db, body.Username, hashedPassword); err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		sendJSONResponse(w, CreateUserResponse{
			Ok:      true,
			Message: "User created successfully",
		})
	}
}

func createUser(db *sql.DB, username, hashedPassword string) error {
	_, err := db.Exec("INSERT INTO users (username, password_hash) VALUES ($1, $2)", username, hashedPassword)
	return err
}
