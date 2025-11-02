package api

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

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

// Predefined error variable for better error handling and comparison
var ErrorUserNotFound = errors.New("user not found")
var ErrorInvalidPassword = errors.New("invalid password")
var ErrorPasswordHashing = errors.New("password hashing error")
var ErrorDatabase = errors.New("database error")

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

		err := createOrValidatePassword(db, body.Username, body.Password)
		if err != nil {
			if errors.Is(err, ErrorUserNotFound) || errors.Is(err, ErrorInvalidPassword) {
				sendJSONResponse(w, LoginErrorResponse{
					Ok:    false,
					Error: "Invalid username or password",
				})
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		user, err := getUserFromUsername(db, body.Username)
		if err != nil {
			log.Printf("ERROR Failed to retrieve user ID for user %s: %v", body.Username, err)
			http.Error(w, "Failed to retrieve user ID", http.StatusInternalServerError)
			return
		}

		key := []byte("your-256-bit-secret")
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"iss":  "jotti",
			"aud":  "jotti-users",
			"iat":  jwt.NewNumericDate(time.Now()),
			"exp":  jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hours
			"sub":  strconv.Itoa(user.ID),
			"role": user.Role,
		})
		stringToken, err := token.SignedString(key)
		if err != nil {
			log.Printf("ERROR Failed to generate token for user %s: %v", body.Username, err)
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		sendJSONResponse(w, LoginResponse{
			Ok:    true,
			Token: stringToken,
		})

	}
}

func createOrValidatePassword(db *sql.DB, username, password string) error {
	storedHash, err := getPasswordHashFromDB(db, username)
	if err != nil {
		log.Printf("ERROR Failed to retrieve password hash for user %s: %v", username, err)
		return ErrorUserNotFound
	}

	log.Printf("DEBUG Retrieved password hash for user %s: %v", username, storedHash)

	if storedHash == nil {
		log.Printf("INFO Setting password for first time login for user %s", username)
		hashedPassword, err := hashPasswordSecure(password)
		if err != nil {
			log.Printf("ERROR Failed to hash password for user %s: %v", username, err)
			return ErrorPasswordHashing
		}
		if err := setPasswordHashInDB(db, username, hashedPassword); err != nil {
			log.Printf("ERROR Failed to set password hash in DB for user %s: %v", username, err)
			return ErrorDatabase
		}
		storedHash = &hashedPassword
		log.Printf("INFO Password set successfully for user %s", username)
	}

	if err := validatePasswordAgainstHash(*storedHash, password); err != nil {
		log.Printf("ERROR Password validation failed for user %s: %v", username, err)
		return ErrorInvalidPassword
	}

	return nil
}

func getPasswordHashFromDB(db *sql.DB, username string) (*string, error) {
	var storedHash *string
	err := db.QueryRow("SELECT password_hash FROM users WHERE username=$1", username).Scan(&storedHash)
	if err != nil {
		return nil, err
	}
	return storedHash, nil
}

func setPasswordHashInDB(db *sql.DB, username, hashedPassword string) error {
	_, err := db.Exec("UPDATE users SET password_hash=$1 WHERE username=$2", hashedPassword, username)
	return err
}

type User struct {
	ID       int
	Username string
	Role     string
}

func getUserFromUsername(db *sql.DB, username string) (User, error) {
	var user User
	err := db.QueryRow("SELECT id, username, role FROM users WHERE username=$1", username).Scan(&user.ID, &user.Username, &user.Role)
	if err != nil {
		return User{}, err
	}
	return user, nil
}
