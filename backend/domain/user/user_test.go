//go:build unit

package user

import (
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
