package application

import (
	"context"
	"errors"
	"testing"

	"github.com/nicograef/jotti/backend/api/auth/throttle"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/user"
	"github.com/nicograef/jotti/backend/repository/repotest"
)

// testUserHash ist der Argon2id-Hash zu "testpassword"; Testbenutzer mit diesem
// Hash durchlaufen echte erfolgreiche und fehlgeschlagene Logins.
const testUserHash = "$argon2id$v=19$m=64,t=2,p=4$QzFPUlMxVUd2Wm51a09BNA$WC7jqeO84JjhcPYJKIN6Ep71DLRc0wog7vjIwYq+EEk"

func TestGenerateJWTToken_NotFound(t *testing.T) {
	repo := repotest.NewUserRepo([]user.User{}, db.ErrNotFound)
	command := Command{UserRepo: repo, JWTSecret: "test-secret", Throttle: throttle.NewLoginThrottle()}

	_, err := command.GenerateJWTToken(context.Background(), "nonexistent", "password")

	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected user not found error, got %v", err)
	}
}

func TestGenerateJWTToken_Success(t *testing.T) {
	repo := repotest.NewUserRepo([]user.User{{ID: 1, Username: "testuser", Status: user.ActiveStatus, PasswordHash: testUserHash}}, nil)
	command := Command{UserRepo: repo, JWTSecret: "test-secret", Throttle: throttle.NewLoginThrottle()}

	token, err := command.GenerateJWTToken(context.Background(), "testuser", "testpassword")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Errorf("expected a token, got empty string")
	}
}

func TestGenerateJWTToken_InvalidPassword(t *testing.T) {
	repo := repotest.NewUserRepo([]user.User{{ID: 1, Username: "testuser", Status: user.ActiveStatus, PasswordHash: testUserHash}}, nil)
	command := Command{UserRepo: repo, JWTSecret: "test-secret", Throttle: throttle.NewLoginThrottle()}

	_, err := command.GenerateJWTToken(context.Background(), "testuser", "wrongpassword")

	if !errors.Is(err, ErrInvalidPassword) {
		t.Errorf("expected invalid password error, got %v", err)
	}
}

func TestGenerateJWTToken_HashError(t *testing.T) {
	repo := repotest.NewUserRepo([]user.User{{ID: 1, Username: "testuser", Status: user.ActiveStatus, PasswordHash: "invalidhashformat"}}, nil)
	command := Command{UserRepo: repo, JWTSecret: "test-secret", Throttle: throttle.NewLoginThrottle()}

	_, err := command.GenerateJWTToken(context.Background(), "testuser", "somepassword")

	if !errors.Is(err, ErrTokenGeneration) {
		t.Errorf("expected token generation error, got %v", err)
	}
}

func TestGenerateJWTToken_UserInactive(t *testing.T) {
	repo := repotest.NewUserRepo([]user.User{{ID: 1, Username: "testuser", Status: user.InactiveStatus, PasswordHash: testUserHash}}, nil)
	command := Command{UserRepo: repo, JWTSecret: "test-secret", Throttle: throttle.NewLoginThrottle()}

	_, err := command.GenerateJWTToken(context.Background(), "testuser", "testpassword")

	if !errors.Is(err, ErrNotActive) {
		t.Errorf("expected user not active error, got %v", err)
	}
}

func TestGenerateJWTToken_ThrottledAfterRepeatedFailures(t *testing.T) {
	repo := repotest.NewUserRepo([]user.User{{ID: 1, Username: "testuser", Status: user.ActiveStatus, PasswordHash: testUserHash}}, nil)
	command := Command{UserRepo: repo, JWTSecret: "test-secret", Throttle: throttle.NewLoginThrottle()}

	// Die Standardschwelle ist 5: fünf Fehlversuche liefern noch invalid_password ...
	for i := range 5 {
		if _, err := command.GenerateJWTToken(context.Background(), "testuser", "wrongpassword"); !errors.Is(err, ErrInvalidPassword) {
			t.Fatalf("Versuch %d: expected ErrInvalidPassword, got %v", i+1, err)
		}
	}

	// ... der nächste Versuch ist gedrosselt (nicht mehr invalid_password).
	if _, err := command.GenerateJWTToken(context.Background(), "testuser", "wrongpassword"); !errors.Is(err, ErrLoginThrottled) {
		t.Errorf("expected ErrLoginThrottled after threshold, got %v", err)
	}
}

func TestGenerateJWTToken_SuccessResetsThrottle(t *testing.T) {
	repo := repotest.NewUserRepo([]user.User{{ID: 1, Username: "testuser", Status: user.ActiveStatus, PasswordHash: testUserHash}}, nil)
	command := Command{UserRepo: repo, JWTSecret: "test-secret", Throttle: throttle.NewLoginThrottle()}

	for range 4 {
		_, _ = command.GenerateJWTToken(context.Background(), "testuser", "wrongpassword")
	}

	// Erfolgreicher Login setzt den Zähler zurück.
	if _, err := command.GenerateJWTToken(context.Background(), "testuser", "testpassword"); err != nil {
		t.Fatalf("expected successful login, got %v", err)
	}

	// Vier weitere Fehlversuche dürfen dank Reset noch nicht drosseln.
	for i := range 4 {
		if _, err := command.GenerateJWTToken(context.Background(), "testuser", "wrongpassword"); !errors.Is(err, ErrInvalidPassword) {
			t.Fatalf("nach Reset, Versuch %d: expected ErrInvalidPassword, got %v", i+1, err)
		}
	}
}

func TestGenerateJWTToken_ThrottleIsPerAccount(t *testing.T) {
	repo := repotest.NewUserRepo([]user.User{
		{ID: 1, Username: "opfer", Status: user.ActiveStatus, PasswordHash: testUserHash},
		{ID: 2, Username: "unbeteiligt", Status: user.ActiveStatus, PasswordHash: testUserHash},
	}, nil)
	command := Command{UserRepo: repo, JWTSecret: "test-secret", Throttle: throttle.NewLoginThrottle()}

	for range 5 {
		_, _ = command.GenerateJWTToken(context.Background(), "opfer", "wrongpassword")
	}
	if _, err := command.GenerateJWTToken(context.Background(), "opfer", "wrongpassword"); !errors.Is(err, ErrLoginThrottled) {
		t.Errorf("expected 'opfer' to be throttled, got %v", err)
	}

	// Ein anderes Konto darf nie betroffen sein.
	if _, err := command.GenerateJWTToken(context.Background(), "unbeteiligt", "testpassword"); err != nil {
		t.Errorf("expected 'unbeteiligt' login to succeed, got %v", err)
	}
}

// falschesEinmalpasswort liefert ein wohlgeformtes Einmalpasswort, das sicher nicht das richtige ist.
func falschesEinmalpasswort(richtig string) string {
	if richtig == "000000" {
		return "111111"
	}
	return "000000"
}

// SetNewPassword bildet jeden Fehler von Repo und Domäne auf genau einen Sentinel ab;
// der HTTP-Handler mappt diese Sentinels auf die Fehlercodes der API.
func TestSetNewPassword_Fehlerabbildung(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		mitEinmalpw   bool
		fehlversuche  int
		falschesOTP   bool
		neuesPasswort string
		repoErr       error
		want          error
	}{
		{"richtiges Einmalpasswort", "anna", true, 0, false, "neuesPasswort1", nil, nil},
		{"falsches Einmalpasswort", "anna", true, 0, true, "neuesPasswort1", nil, ErrInvalidPassword},
		{"fünfter Fehlversuch sperrt", "anna", true, 4, true, "neuesPasswort1", nil, ErrOnetimePasswordLocked},
		{"nach der Sperre kein Einmalpasswort mehr", "anna", true, 5, false, "neuesPasswort1", nil, ErrNoOnetimePassword},
		{"kein Einmalpasswort gesetzt", "anna", false, 0, false, "neuesPasswort1", nil, ErrNoOnetimePassword},
		{"neues Passwort zu schwach", "anna", true, 0, false, "abc", nil, ErrPasswordTooWeak},
		{"unbekannter Benutzer", "niemand", true, 0, false, "neuesPasswort1", nil, ErrUserNotFound},
		{"Datenbankfehler", "anna", true, 0, false, "neuesPasswort1", errors.New("verbindung verloren"), ErrDatabase},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anna := user.User{ID: 1, Username: "anna", Status: user.ActiveStatus, PasswordHash: testUserHash}
			einmalpw := ""
			if tt.mitEinmalpw {
				var err error
				if einmalpw, err = anna.ResetPassword(); err != nil {
					t.Fatalf("Einmalpasswort erzeugen: %v", err)
				}
			}
			repo := repotest.NewUserRepo([]user.User{anna}, tt.repoErr)
			command := Command{UserRepo: repo, JWTSecret: "test-secret", Throttle: throttle.NewLoginThrottle()}
			ctx := context.Background()

			for i := range tt.fehlversuche {
				if err := command.SetNewPassword(ctx, "anna", "neuesPasswort1", falschesEinmalpasswort(einmalpw)); err == nil {
					t.Fatalf("Fehlversuch %d: expected an error, got nil", i+1)
				}
			}

			otp := einmalpw
			if tt.falschesOTP || !tt.mitEinmalpw {
				otp = falschesEinmalpasswort(einmalpw)
			}
			err := command.SetNewPassword(ctx, tt.username, tt.neuesPasswort, otp)

			if !errors.Is(err, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, err)
			}
			if tt.want != nil {
				return
			}
			if _, err := command.GenerateJWTToken(ctx, "anna", "neuesPasswort1"); err != nil {
				t.Errorf("expected login with the new password, got %v", err)
			}
			if err := command.SetNewPassword(ctx, "anna", "nochEinsNeu1", einmalpw); !errors.Is(err, ErrNoOnetimePassword) {
				t.Errorf("expected the used one-time password to be gone, got %v", err)
			}
		})
	}
}
