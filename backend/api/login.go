package api

import (
	"database/sql"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type LoginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Ok    bool   `json:"ok"`
	Token string `json:"token"`
}

type LoginErrorResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

// NewLoginHandler handles user login requests by validating the password hash against the database
// and returns a jwt token if successful.
func NewLoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateMethod(w, r, http.MethodPost) {
			return
		}

		body := LoginBody{}
		if !readJSONRequest(w, r, &body) {
			return
		}

		storedHash, err := getPasswordHashFromDB(db, body.Username)
		if err != nil {
			http.Error(w, "Failed to retrieve user", http.StatusUnauthorized)
			return
		}

		if err := authenticateUser(storedHash, body.Password); err != nil {
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		key := []byte("your-256-bit-secret")
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"iss": "jotti",
			"sub": body.Username,
		})
		stringToken, err := token.SignedString(key)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		sendJSONResponse(w, LoginResponse{
			Ok:    true,
			Token: stringToken,
		})

	}
}

func getPasswordHashFromDB(db *sql.DB, username string) (string, error) {
	var storedHash string
	err := db.QueryRow("SELECT password_hash FROM users WHERE username=$1", username).Scan(&storedHash)
	if err != nil {
		return "", err
	}
	return storedHash, nil
}
