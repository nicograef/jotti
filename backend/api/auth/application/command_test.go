//go:build unit

package application

import (
	"context"
	"testing"

	"github.com/nicograef/jotti/backend/api/auth/throttle"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/user"
	"github.com/nicograef/jotti/backend/repository/user_repo"
)

func TestGenerateJWTToken_NotFound(t *testing.T) {
	repo := user_repo.NewMock([]user.User{}, db.ErrNotFound)
	command := Command{UserRepo: repo, JWTSecret: "test-secret", Throttle: throttle.NewLoginThrottle()}

	_, err := command.GenerateJWTToken(context.Background(), "nonexistent", "password")

	if err != ErrUserNotFound {
		t.Fatalf("expected user not found error, got %v", err)
	}
}

func TestGenerateJWTToken_Success(t *testing.T) {
	repo := user_repo.NewMock([]user.User{{ID: 1, Username: "testuser", Status: user.ActiveStatus, PasswordHash: "$argon2id$v=19$m=64,t=2,p=4$QzFPUlMxVUd2Wm51a09BNA$WC7jqeO84JjhcPYJKIN6Ep71DLRc0wog7vjIwYq+EEk"}}, nil)
	command := Command{UserRepo: repo, JWTSecret: "test-secret", Throttle: throttle.NewLoginThrottle()}

	token, err := command.GenerateJWTToken(context.Background(), "testuser", "testpassword")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatalf("expected a token, got empty string")
	}
}

func TestGenerateJWTToken_InvalidPassword(t *testing.T) {
	repo := user_repo.NewMock([]user.User{{ID: 1, Username: "testuser", Status: user.ActiveStatus, PasswordHash: "$argon2id$v=19$m=64,t=2,p=4$QzFPUlMxVUd2Wm51a09BNA$WC7jqeO84JjhcPYJKIN6Ep71DLRc0wog7vjIwYq+EEk"}}, nil)
	command := Command{UserRepo: repo, JWTSecret: "test-secret", Throttle: throttle.NewLoginThrottle()}

	_, err := command.GenerateJWTToken(context.Background(), "testuser", "wrongpassword")

	if err != ErrInvalidPassword {
		t.Fatalf("expected invalid password error, got %v", err)
	}
}

func TestGenerateJWTToken_HashError(t *testing.T) {
	repo := user_repo.NewMock([]user.User{{ID: 1, Username: "testuser", Status: user.ActiveStatus, PasswordHash: "invalidhashformat"}}, nil)
	command := Command{UserRepo: repo, JWTSecret: "test-secret", Throttle: throttle.NewLoginThrottle()}

	_, err := command.GenerateJWTToken(context.Background(), "testuser", "somepassword")

	if err != ErrTokenGeneration {
		t.Fatalf("expected token generation error, got %v", err)
	}
}

func TestGenerateJWTToken_UserInactive(t *testing.T) {
	repo := user_repo.NewMock([]user.User{{ID: 1, Username: "testuser", Status: user.InactiveStatus, PasswordHash: "$argon2id$v=19$m=64,t=2,p=4$QzFPUlMxVUd2Wm51a09BNA$WC7jqeO84JjhcPYJKIN6Ep71DLRc0wog7vjIwYq+EEk"}}, nil)
	command := Command{UserRepo: repo, JWTSecret: "test-secret", Throttle: throttle.NewLoginThrottle()}

	_, err := command.GenerateJWTToken(context.Background(), "testuser", "testpassword")

	if err != ErrNotActive {
		t.Fatalf("expected user not active error, got %v", err)
	}
}

// activeUserHash ist der Argon2id-Hash zu "testpassword", damit die Throttle-Tests
// echte erfolgreiche und fehlgeschlagene Logins durchspielen können.
const activeUserHash = "$argon2id$v=19$m=64,t=2,p=4$QzFPUlMxVUd2Wm51a09BNA$WC7jqeO84JjhcPYJKIN6Ep71DLRc0wog7vjIwYq+EEk"

func TestGenerateJWTToken_ThrottledAfterRepeatedFailures(t *testing.T) {
	repo := user_repo.NewMock([]user.User{{ID: 1, Username: "testuser", Status: user.ActiveStatus, PasswordHash: activeUserHash}}, nil)
	command := Command{UserRepo: repo, JWTSecret: "test-secret", Throttle: throttle.NewLoginThrottle()}

	// Die Standardschwelle ist 5: fünf Fehlversuche liefern noch invalid_password ...
	for i := 0; i < 5; i++ {
		if _, err := command.GenerateJWTToken(context.Background(), "testuser", "wrongpassword"); err != ErrInvalidPassword {
			t.Fatalf("Versuch %d: expected ErrInvalidPassword, got %v", i+1, err)
		}
	}

	// ... der nächste Versuch ist gedrosselt (nicht mehr invalid_password).
	if _, err := command.GenerateJWTToken(context.Background(), "testuser", "wrongpassword"); err != ErrLoginThrottled {
		t.Fatalf("expected ErrLoginThrottled after threshold, got %v", err)
	}
}

func TestGenerateJWTToken_SuccessResetsThrottle(t *testing.T) {
	repo := user_repo.NewMock([]user.User{{ID: 1, Username: "testuser", Status: user.ActiveStatus, PasswordHash: activeUserHash}}, nil)
	command := Command{UserRepo: repo, JWTSecret: "test-secret", Throttle: throttle.NewLoginThrottle()}

	for i := 0; i < 4; i++ {
		_, _ = command.GenerateJWTToken(context.Background(), "testuser", "wrongpassword")
	}

	// Erfolgreicher Login setzt den Zähler zurück.
	if _, err := command.GenerateJWTToken(context.Background(), "testuser", "testpassword"); err != nil {
		t.Fatalf("expected successful login, got %v", err)
	}

	// Vier weitere Fehlversuche dürfen dank Reset noch nicht drosseln.
	for i := 0; i < 4; i++ {
		if _, err := command.GenerateJWTToken(context.Background(), "testuser", "wrongpassword"); err != ErrInvalidPassword {
			t.Fatalf("nach Reset, Versuch %d: expected ErrInvalidPassword, got %v", i+1, err)
		}
	}
}

func TestGenerateJWTToken_ThrottleIsPerAccount(t *testing.T) {
	repo := user_repo.NewMock([]user.User{
		{ID: 1, Username: "opfer", Status: user.ActiveStatus, PasswordHash: activeUserHash},
		{ID: 2, Username: "unbeteiligt", Status: user.ActiveStatus, PasswordHash: activeUserHash},
	}, nil)
	command := Command{UserRepo: repo, JWTSecret: "test-secret", Throttle: throttle.NewLoginThrottle()}

	for i := 0; i < 5; i++ {
		_, _ = command.GenerateJWTToken(context.Background(), "opfer", "wrongpassword")
	}
	if _, err := command.GenerateJWTToken(context.Background(), "opfer", "wrongpassword"); err != ErrLoginThrottled {
		t.Fatalf("expected 'opfer' to be throttled, got %v", err)
	}

	// Ein anderes Konto darf nie betroffen sein.
	if _, err := command.GenerateJWTToken(context.Background(), "unbeteiligt", "testpassword"); err != nil {
		t.Fatalf("expected 'unbeteiligt' login to succeed, got %v", err)
	}
}
