//go:build unit

package user

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestActivate_SetsUpdatedAt(t *testing.T) {
	u, _, err := NewUser("Test User", "testuser", ServiceRole)
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}
	u.Status = InactiveStatus
	before := u.UpdatedAt.Add(-time.Millisecond)

	u.Activate()

	if !u.UpdatedAt.After(before) {
		t.Error("Activate() did not advance UpdatedAt")
	}
	if u.Status != ActiveStatus {
		t.Errorf("expected status %s, got %s", ActiveStatus, u.Status)
	}
}

func TestDeactivate_SetsUpdatedAt(t *testing.T) {
	u, _, err := NewUser("Test User", "testuser", ServiceRole)
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}
	before := u.UpdatedAt.Add(-time.Millisecond)

	u.Deactivate()

	if !u.UpdatedAt.After(before) {
		t.Error("Deactivate() did not advance UpdatedAt")
	}
	if u.Status != InactiveStatus {
		t.Errorf("expected status %s, got %s", InactiveStatus, u.Status)
	}
}

func TestDelete_SetsUpdatedAt(t *testing.T) {
	u, _, err := NewUser("Test User", "testuser", ServiceRole)
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}
	before := u.UpdatedAt.Add(-time.Millisecond)

	u.Delete()

	if !u.UpdatedAt.After(before) {
		t.Error("Delete() did not advance UpdatedAt")
	}
	if u.Status != DeletedStatus {
		t.Errorf("expected status %s, got %s", DeletedStatus, u.Status)
	}
}

func TestUpdateDetails_SetsUpdatedAt(t *testing.T) {
	u, _, err := NewUser("Test User", "testuser", ServiceRole)
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}
	before := u.UpdatedAt.Add(-time.Millisecond)

	err = u.UpdateDetails("New Name", "newuser", AdminRole)
	if err != nil {
		t.Fatalf("UpdateDetails: %v", err)
	}

	if !u.UpdatedAt.After(before) {
		t.Error("UpdateDetails() did not advance UpdatedAt")
	}
	if u.Name != "New Name" {
		t.Errorf("expected name %q, got %q", "New Name", u.Name)
	}
}

func TestResetPassword_SetsUpdatedAt(t *testing.T) {
	u, _, err := NewUser("Test User", "testuser", ServiceRole)
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}
	before := u.UpdatedAt.Add(-time.Millisecond)

	_, err = u.ResetPassword()
	if err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}

	if !u.UpdatedAt.After(before) {
		t.Error("ResetPassword() did not advance UpdatedAt")
	}
}

func TestSetPassword_SetsUpdatedAt(t *testing.T) {
	u, onetimePassword, err := NewUser("Test User", "testuser", ServiceRole)
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}
	before := u.UpdatedAt.Add(-time.Millisecond)

	err = u.SetPassword(onetimePassword, "newSecurePass123")
	if err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	if !u.UpdatedAt.After(before) {
		t.Error("SetPassword() did not advance UpdatedAt")
	}
	if u.OnetimePasswordHash != "" {
		t.Error("SetPassword() did not clear OnetimePasswordHash")
	}
}

func TestSetPassword_NormalisiertEingabe(t *testing.T) {
	u, onetimePassword, err := NewUser("Test User", "testuser", ServiceRole)
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}

	// Großschreibung und umgebende Leerzeichen werden toleriert.
	err = u.SetPassword("  "+strings.ToUpper(onetimePassword)+" ", "newSecurePass123")
	if err != nil {
		t.Fatalf("SetPassword mit normalisierbarer Eingabe: %v", err)
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

	// Der fünfte Fehlversuch sperrt: Einmalpasswort wird ungültig.
	err = u.SetPassword("falsch99", "newSecurePass123")
	if !errors.Is(err, ErrOnetimePasswordLocked) {
		t.Fatalf("expected ErrOnetimePasswordLocked, got %v", err)
	}
	if u.OnetimePasswordHash != "" {
		t.Error("expected OnetimePasswordHash to be invalidated after lockout")
	}

	// Auch das korrekte Einmalpasswort funktioniert danach nicht mehr.
	err = u.SetPassword(onetimePassword, "newSecurePass123")
	if !errors.Is(err, ErrNoPassword) {
		t.Fatalf("expected ErrNoPassword after lockout, got %v", err)
	}

	// Admin-Reset erzeugt ein frisches Einmalpasswort und setzt den Zähler zurück.
	neues, err := u.ResetPassword()
	if err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}
	if u.OnetimePasswordAttempts != 0 {
		t.Errorf("expected attempts reset to 0, got %d", u.OnetimePasswordAttempts)
	}
	if err := u.SetPassword(neues, "newSecurePass123"); err != nil {
		t.Fatalf("SetPassword nach Reset: %v", err)
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
