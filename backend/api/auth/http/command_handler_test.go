//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/api/auth/application"
	"github.com/nicograef/jotti/backend/domain/user"
)

type mockAuthCommand struct {
	token string
	err   error
}

func (m mockAuthCommand) GenerateJWTToken(ctx context.Context, username, password string) (string, error) {
	return m.token, m.err
}

func (m mockAuthCommand) SetNewPassword(ctx context.Context, username, password, onetimePassword string) error {
	return m.err
}

func TestLoginHandler_Success(t *testing.T) {
	command := mockAuthCommand{token: "test-token", err: nil}
	handler := CommandHandler{Command: command}

	body := `{"username":"testuser","password":"Test123!"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.LoginHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	command := mockAuthCommand{token: "", err: application.ErrInvalidPassword}
	handler := CommandHandler{Command: command}

	body := `{"username":"testuser","password":"wrongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.LoginHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestLoginHandler_Throttled(t *testing.T) {
	command := mockAuthCommand{token: "", err: application.ErrLoginThrottled}
	handler := CommandHandler{Command: command}

	body := `{"username":"testuser","password":"Test123!"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.LoginHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429 for throttled login, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "login_throttled") {
		t.Errorf("expected body to carry the login_throttled code, got %s", rec.Body.String())
	}
}

func TestLoginHandler_InactiveUser(t *testing.T) {
	command := mockAuthCommand{token: "", err: application.ErrNotActive}
	handler := CommandHandler{Command: command}

	body := `{"username":"testuser","password":"Test123!"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.LoginHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestLoginHandler_InvalidInput(t *testing.T) {
	command := mockAuthCommand{token: "test-token", err: nil}
	handler := CommandHandler{Command: command}

	body := `{"username":"","password":""}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.LoginHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestSetPasswordHandler_InvalidInput(t *testing.T) {
	command := mockAuthCommand{err: nil}
	handler := CommandHandler{Command: command}

	body := `{"username":"INVALID","password":"123","onetimePassword":""}`
	req := httptest.NewRequest(http.MethodPost, "/set-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.SetPasswordHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

// Ein Einmalpasswort, das nicht genau 6 Ziffern ist, wird vom Schema abgelehnt,
// bevor der Command (und damit der Hashvergleich) überhaupt erreicht wird.
func TestSetPasswordHandler_RejectsNon6DigitOTP(t *testing.T) {
	cases := []struct {
		name string
		otp  string
	}{
		{"empty", ""},
		{"too short", "12345"},
		{"too long", "1234567"},
		{"letters", "abcdef"},
		{"mixed", "12ab56"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Der Command darf gar nicht erst aufgerufen werden — schlägt er an, ist
			// die Formatvalidierung durchgerutscht.
			command := mockAuthCommand{err: nil}
			handler := CommandHandler{Command: command}

			body := `{"username":"testuser","password":"newSecurePass123","onetimePassword":"` + tc.otp + `"}`
			req := httptest.NewRequest(http.MethodPost, "/set-password", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.SetPasswordHandler().ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status 400 for OTP %q, got %d", tc.otp, rec.Code)
			}
		})
	}
}

// Ein wohlgeformtes 6-Ziffern-Einmalpasswort passiert die Formatvalidierung und
// erreicht den Command.
func TestSetPasswordHandler_AcceptsValid6DigitOTP(t *testing.T) {
	command := mockAuthCommand{err: nil}
	handler := CommandHandler{Command: command}

	body := `{"username":"testuser","password":"newSecurePass123","onetimePassword":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/set-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.SetPasswordHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 for a valid 6-digit OTP, got %d", rec.Code)
	}
}

// passwordVerifyingCommand prüft das übergebene Passwort gegen einen echten
// Benutzerdatensatz. So misst der Test die Normalisierung des Login-Schemas am
// tatsächlich gesetzten Passwort statt an einem festen Rückgabewert.
type passwordVerifyingCommand struct {
	stored user.User
}

func (c passwordVerifyingCommand) GenerateJWTToken(ctx context.Context, username, password string) (string, error) {
	if err := c.stored.VerifyPassword(password); err != nil {
		return "", application.ErrInvalidPassword
	}
	return "test-token", nil
}

func (c passwordVerifyingCommand) SetNewPassword(ctx context.Context, username, password, onetimePassword string) error {
	return nil
}

// Das Setzen trimmt das Passwort (PasswordSchema), also muss der Login dieselbe
// Normalisierung anwenden: beide Schreibweisen führen auf dasselbe Geheimnis.
func TestLoginHandler_TrimsSurroundingWhitespace(t *testing.T) {
	stored := user.User{Username: "testuser", Status: user.ActiveStatus}
	onetimePassword, err := stored.ResetPassword()
	if err != nil {
		t.Fatalf("failed to reset password: %v", err)
	}
	if err := stored.SetPassword(onetimePassword, "  geheim123  "); err != nil {
		t.Fatalf("failed to set password: %v", err)
	}

	cases := []struct {
		name     string
		password string
	}{
		{"trimmed", "geheim123"},
		{"padded", "  geheim123  "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := CommandHandler{Command: passwordVerifyingCommand{stored: stored}}

			body := `{"username":"testuser","password":"` + tc.password + `"}`
			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.LoginHandler().ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status 200 for password %q, got %d", tc.password, rec.Code)
			}
		})
	}
}
