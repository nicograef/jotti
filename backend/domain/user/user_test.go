package user

import (
	"errors"
	"testing"
	"time"
)

// Every mutation stamps UpdatedAt; the fixed past value fails the check when a
// mutation leaves it untouched.
func TestMutationen_SetzenUpdatedAt(t *testing.T) {
	vorher := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name  string
		apply func(u *User, onetimePassword string) error
		check func(t *testing.T, u *User)
	}{
		{
			name: "Activate",
			apply: func(u *User, _ string) error {
				u.Status = InactiveStatus
				u.Activate()
				return nil
			},
			check: func(t *testing.T, u *User) { wantStatus(t, u, ActiveStatus) },
		},
		{
			name:  "Deactivate",
			apply: func(u *User, _ string) error { u.Deactivate(); return nil },
			check: func(t *testing.T, u *User) { wantStatus(t, u, InactiveStatus) },
		},
		{
			name:  "Delete",
			apply: func(u *User, _ string) error { u.Delete(); return nil },
			check: func(t *testing.T, u *User) { wantStatus(t, u, DeletedStatus) },
		},
		{
			name:  "UpdateDetails",
			apply: func(u *User, _ string) error { return u.UpdateDetails("New Name", "newuser", AdminRole) },
			check: func(t *testing.T, u *User) {
				if u.Name != "New Name" {
					t.Errorf("expected name %q, got %q", "New Name", u.Name)
				}
			},
		},
		{
			name: "ResetPassword",
			apply: func(u *User, _ string) error {
				_, err := u.ResetPassword()
				return err
			},
			check: func(*testing.T, *User) {},
		},
		{
			name:  "SetPassword",
			apply: func(u *User, onetimePassword string) error { return u.SetPassword(onetimePassword, "newSecurePass123") },
			check: func(t *testing.T, u *User) {
				if u.OnetimePasswordHash != "" {
					t.Error("SetPassword() did not clear OnetimePasswordHash")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, onetimePassword, err := NewUser("Test User", "testuser", ServiceRole)
			if err != nil {
				t.Fatalf("NewUser: %v", err)
			}
			u.UpdatedAt = vorher

			if err := tt.apply(&u, onetimePassword); err != nil {
				t.Fatalf("%s: %v", tt.name, err)
			}

			if !u.UpdatedAt.After(vorher) {
				t.Errorf("%s() did not advance UpdatedAt", tt.name)
			}
			tt.check(t, &u)
		})
	}
}

func wantStatus(t *testing.T, u *User, want Status) {
	t.Helper()
	if u.Status != want {
		t.Errorf("expected status %s, got %s", want, u.Status)
	}
}

func TestSetPassword_TrimmtEingabe(t *testing.T) {
	u, onetimePassword, err := NewUser("Test User", "testuser", ServiceRole)
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}

	err = u.SetPassword("  "+onetimePassword+" ", "newSecurePass123")
	if err != nil {
		t.Errorf("SetPassword mit getrimmter Eingabe: %v", err)
	}
}

func TestSetPassword_SperrtNachFuenfFehlversuchen(t *testing.T) {
	u, onetimePassword, err := NewUser("Test User", "testuser", ServiceRole)
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}

	for i := 1; i < MaxOnetimePasswordAttempts; i++ {
		err = u.SetPassword("falsch99", "newSecurePass123")
		if !errors.Is(err, ErrInvalidPassword) {
			t.Fatalf("Fehlversuch %d: expected ErrInvalidPassword, got %v", i, err)
		}
		if u.OnetimePasswordAttempts != i {
			t.Fatalf("Fehlversuch %d: expected attempts %d, got %d", i, i, u.OnetimePasswordAttempts)
		}
	}

	err = u.SetPassword("falsch99", "newSecurePass123")
	if !errors.Is(err, ErrOnetimePasswordLocked) {
		t.Fatalf("expected ErrOnetimePasswordLocked, got %v", err)
	}
	if u.OnetimePasswordHash != "" {
		t.Error("expected OnetimePasswordHash to be invalidated after lockout")
	}

	err = u.SetPassword(onetimePassword, "newSecurePass123")
	if !errors.Is(err, ErrNoPassword) {
		t.Errorf("expected ErrNoPassword after lockout, got %v", err)
	}

	neues, err := u.ResetPassword()
	if err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}
	if u.OnetimePasswordAttempts != 0 {
		t.Errorf("expected attempts reset to 0, got %d", u.OnetimePasswordAttempts)
	}
	if err := u.SetPassword(neues, "newSecurePass123"); err != nil {
		t.Errorf("SetPassword nach Reset: %v", err)
	}
}

func TestSetPassword_ErfolgSetztZaehlerZurueck(t *testing.T) {
	u, onetimePassword, err := NewUser("Test User", "testuser", ServiceRole)
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}

	if err := u.SetPassword("falsch99", "newSecurePass123"); !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
	if err := u.SetPassword(onetimePassword, "newSecurePass123"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}
	if u.OnetimePasswordAttempts != 0 {
		t.Errorf("expected attempts reset to 0 after success, got %d", u.OnetimePasswordAttempts)
	}
}
