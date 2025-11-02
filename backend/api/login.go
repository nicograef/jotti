package api

import (
	"database/sql"
	"log"
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
// If this is the first time the user logs in (no password hash set), it sets the provided password as the new password.
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

		if storedHash == "" {
			log.Printf("INFO Setting password for first time login for user %s", body.Username)
			if err := setPasswordHashInDB(db, body.Username, body.Password); err != nil {
				http.Error(w, "Failed to set password", http.StatusInternalServerError)
				return
			}
		}

		if err := validatePasswordAgainstHash(storedHash, body.Password); err != nil {
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}

		userId, err := getUserIdFromUsername(db, body.Username)
		if err != nil {
			http.Error(w, "Failed to retrieve user ID", http.StatusInternalServerError)
			return
		}

		key := []byte("your-256-bit-secret")
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"iss": "jotti",
			"sub": userId,
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

func setPasswordHashInDB(db *sql.DB, username, hashedPassword string) error {
	_, err := db.Exec("UPDATE users SET password_hash=$1 WHERE username=$2", hashedPassword, username)
	return err
}

func getUserIdFromUsername(db *sql.DB, username string) (int, error) {
	var userId int
	err := db.QueryRow("SELECT id FROM users WHERE username=$1", username).Scan(&userId)
	if err != nil {
		return 0, err
	}
	return userId, nil
}
