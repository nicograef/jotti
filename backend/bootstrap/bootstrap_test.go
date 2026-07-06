//go:build unit

package bootstrap_test

import (
	"bytes"
	"context"
	"regexp"
	"testing"

	"github.com/rs/zerolog"

	"github.com/nicograef/jotti/backend/bootstrap"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/user"
)

var sixDigits = regexp.MustCompile(`^\d{6}$`)

// fakeRepo ist ein reines In-Memory-Repository, das bootstrap.Repository erfüllt —
// ohne echte Datenbank, damit die Entscheidungslogik isoliert getestet werden kann.
type fakeRepo struct {
	users  map[int]user.User
	nextID int
}

func newFakeRepo(users ...user.User) *fakeRepo {
	r := &fakeRepo{users: make(map[int]user.User)}
	for _, u := range users {
		r.nextID++
		u.ID = r.nextID
		r.users[u.ID] = u
	}
	return r
}

func (r *fakeRepo) CountUsers(ctx context.Context) (int, error) {
	return len(r.users), nil
}

func (r *fakeRepo) GetUserByUsername(ctx context.Context, username string) (user.User, error) {
	for _, u := range r.users {
		if u.Username == username {
			return u, nil
		}
	}
	return user.User{}, db.ErrNotFound
}

func (r *fakeRepo) CreateUser(ctx context.Context, u user.User) (int, error) {
	r.nextID++
	u.ID = r.nextID
	r.users[u.ID] = u
	return u.ID, nil
}

func (r *fakeRepo) UpdateUser(ctx context.Context, u user.User) error {
	r.users[u.ID] = u
	return nil
}

// byUsername liefert den (gespeicherten) Benutzer für Assertions.
func (r *fakeRepo) byUsername(t *testing.T, username string) user.User {
	t.Helper()
	for _, u := range r.users {
		if u.Username == username {
			return u
		}
	}
	t.Fatalf("expected user %q to exist in fake repo", username)
	return user.User{}
}

// newAdminWithoutPassword baut einen aktiven admin ohne Passwort, aber mit
// gesetztem Einmalpasswort-Hash (Ausgangszustand des Rotations-/Wiederherstellungsfalls).
func newAdminWithoutPassword(t *testing.T) user.User {
	t.Helper()
	u, _, err := user.NewUser("Administrator", bootstrap.AdminUsername, user.AdminRole)
	if err != nil {
		t.Fatalf("failed to build admin: %v", err)
	}
	u.Activate()
	return u
}

func TestEnsureInitialAdmin(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) *fakeRepo
		wantAction bootstrap.Action
		check      func(t *testing.T, repo *fakeRepo, res bootstrap.Result)
	}{
		{
			name:       "empty repo creates active admin",
			setup:      func(t *testing.T) *fakeRepo { return newFakeRepo() },
			wantAction: bootstrap.ActionCreate,
			check: func(t *testing.T, repo *fakeRepo, res bootstrap.Result) {
				if !sixDigits.MatchString(res.OnetimePassword) {
					t.Fatalf("expected 6-digit OTP, got %q", res.OnetimePassword)
				}
				admin := repo.byUsername(t, bootstrap.AdminUsername)
				if admin.Role != user.AdminRole {
					t.Fatalf("expected AdminRole, got %q", admin.Role)
				}
				if admin.Status != user.ActiveStatus {
					t.Fatalf("expected active admin, got %q", admin.Status)
				}
				if admin.PasswordHash != "" {
					t.Fatalf("expected empty PasswordHash, got %q", admin.PasswordHash)
				}
				if admin.OnetimePasswordHash == "" {
					t.Fatal("expected non-empty OnetimePasswordHash after create")
				}
			},
		},
		{
			name: "single admin without password rotates OTP",
			setup: func(t *testing.T) *fakeRepo {
				return newFakeRepo(newAdminWithoutPassword(t))
			},
			wantAction: bootstrap.ActionRotate,
			check: func(t *testing.T, repo *fakeRepo, res bootstrap.Result) {
				if !sixDigits.MatchString(res.OnetimePassword) {
					t.Fatalf("expected 6-digit OTP, got %q", res.OnetimePassword)
				}
				admin := repo.byUsername(t, bootstrap.AdminUsername)
				if admin.OnetimePasswordHash == "" {
					t.Fatal("expected fresh non-empty OnetimePasswordHash after rotate")
				}
				if admin.OnetimePasswordAttempts != 0 {
					t.Fatalf("expected attempts reset to 0, got %d", admin.OnetimePasswordAttempts)
				}
				if admin.PasswordHash != "" {
					t.Fatalf("expected PasswordHash to stay empty, got %q", admin.PasswordHash)
				}
				if admin.Status != user.ActiveStatus {
					t.Fatalf("expected status unchanged (active), got %q", admin.Status)
				}
			},
		},
		{
			name: "single locked admin (empty OTP hash) rotates to fresh OTP",
			setup: func(t *testing.T) *fakeRepo {
				admin := newAdminWithoutPassword(t)
				// Aussperrung: das OTP-Hash wurde nach zu vielen Fehlversuchen geleert.
				admin.OnetimePasswordHash = ""
				admin.OnetimePasswordAttempts = 0
				return newFakeRepo(admin)
			},
			wantAction: bootstrap.ActionRotate,
			check: func(t *testing.T, repo *fakeRepo, res bootstrap.Result) {
				if !sixDigits.MatchString(res.OnetimePassword) {
					t.Fatalf("expected 6-digit OTP, got %q", res.OnetimePassword)
				}
				admin := repo.byUsername(t, bootstrap.AdminUsername)
				if admin.OnetimePasswordHash == "" {
					t.Fatal("expected fresh non-empty OnetimePasswordHash after self-healing rotate")
				}
				if admin.OnetimePasswordAttempts != 0 {
					t.Fatalf("expected attempts 0, got %d", admin.OnetimePasswordAttempts)
				}
			},
		},
		{
			name: "single admin with password is skipped",
			setup: func(t *testing.T) *fakeRepo {
				admin := newAdminWithoutPassword(t)
				admin.PasswordHash = "existing-password-hash"
				admin.OnetimePasswordHash = ""
				return newFakeRepo(admin)
			},
			wantAction: bootstrap.ActionSkip,
			check: func(t *testing.T, repo *fakeRepo, res bootstrap.Result) {
				if res.OnetimePassword != "" {
					t.Fatalf("expected empty OTP on skip, got %q", res.OnetimePassword)
				}
				admin := repo.byUsername(t, bootstrap.AdminUsername)
				if admin.PasswordHash != "existing-password-hash" {
					t.Fatalf("expected PasswordHash unchanged, got %q", admin.PasswordHash)
				}
				if admin.OnetimePasswordHash != "" {
					t.Fatalf("expected OnetimePasswordHash untouched (empty), got %q", admin.OnetimePasswordHash)
				}
			},
		},
		{
			name: "multiple users leaves open service OTP untouched",
			setup: func(t *testing.T) *fakeRepo {
				admin := newAdminWithoutPassword(t)
				service, _, err := user.NewUser("Service Kraft", "service1", user.ServiceRole)
				if err != nil {
					t.Fatalf("failed to build service user: %v", err)
				}
				return newFakeRepo(admin, service)
			},
			wantAction: bootstrap.ActionSkip,
			check: func(t *testing.T, repo *fakeRepo, res bootstrap.Result) {
				if res.OnetimePassword != "" {
					t.Fatalf("expected empty OTP on skip, got %q", res.OnetimePassword)
				}
				service := repo.byUsername(t, "service1")
				if service.OnetimePasswordHash == "" {
					t.Fatal("expected the service user's open OTP hash to stay intact")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setup(t)

			// Alten OTP-Hash des Admins vor der Entscheidung merken (für Rotations-Assertion).
			var oldAdminOTPHash string
			if before, err := repo.GetUserByUsername(context.Background(), bootstrap.AdminUsername); err == nil {
				oldAdminOTPHash = before.OnetimePasswordHash
			}

			res, err := bootstrap.EnsureInitialAdmin(context.Background(), repo)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if res.Action != tt.wantAction {
				t.Fatalf("expected action %q, got %q", tt.wantAction, res.Action)
			}

			if tt.wantAction == bootstrap.ActionRotate {
				admin := repo.byUsername(t, bootstrap.AdminUsername)
				if admin.OnetimePasswordHash == oldAdminOTPHash {
					t.Fatal("expected the OTP hash to change on rotate")
				}
			}

			tt.check(t, repo, res)
		})
	}
}

// TestResultLog_MarkerSurvivesConsoleWriter belegt, dass der grep-stabile Präfix und
// der Klartext-Code die zerolog-ConsoleWriter-Formatierung als Literal überstehen.
func TestResultLog_MarkerSurvivesConsoleWriter(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(zerolog.ConsoleWriter{Out: &buf, NoColor: true})

	bootstrap.Result{Action: bootstrap.ActionCreate, OnetimePassword: "123456"}.Log(logger)

	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("ADMIN-EINMALPASSWORT")) {
		t.Fatalf("expected grep-stable marker prefix in log output, got: %s", out)
	}
	if !bytes.Contains([]byte(out), []byte("123456")) {
		t.Fatalf("expected plaintext OTP in log output, got: %s", out)
	}
}

// TestResultLog_SkipWritesNothing: bei ActionSkip darf keine Zeile entstehen.
func TestResultLog_SkipWritesNothing(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(zerolog.ConsoleWriter{Out: &buf, NoColor: true})

	bootstrap.Result{Action: bootstrap.ActionSkip}.Log(logger)

	if buf.Len() != 0 {
		t.Fatalf("expected no log output on skip, got: %s", buf.String())
	}
}
