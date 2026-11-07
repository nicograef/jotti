package application

import (
	"context"
	"errors"
	"testing"

	"github.com/nicograef/jotti/backend/domain/user"
	"github.com/nicograef/jotti/backend/repository/repotest"
)

const adminID = 42

func commandMitAdmin() Command {
	admin := user.User{ID: adminID, Name: "Anna Admin", Username: "anna", Role: user.AdminRole, Status: user.ActiveStatus}
	andere := user.User{ID: 7, Name: "Sam Service", Username: "sam", Role: user.ServiceRole, Status: user.ActiveStatus}
	return Command{UserRepo: repotest.NewUserRepo([]user.User{admin, andere}, nil)}
}

// An admin can neither demote, deactivate nor delete the own account: each locks the
// instance out once no other admin is left.
func TestCommand_SchuetztEigenesKonto(t *testing.T) {
	tests := []struct {
		name string
		call func(Command) error
		want error
	}{
		{"demote", func(c Command) error {
			return c.UpdateUser(context.Background(), adminID, adminID, "Anna Admin", "anna", user.ServiceRole)
		}, ErrCannotDemoteSelf},
		{"deactivate", func(c Command) error { return c.DeactivateUser(context.Background(), adminID, adminID) }, ErrCannotDeactivateSelf},
		{"delete", func(c Command) error { return c.DeleteUser(context.Background(), adminID, adminID) }, ErrCannotDeleteSelf},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := commandMitAdmin()
			if err := tt.call(c); !errors.Is(err, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, err)
			}
			u, err := c.UserRepo.GetUser(context.Background(), adminID)
			if err != nil {
				t.Fatalf("GetUser: %v", err)
			}
			if u.Role != user.AdminRole || u.Status != user.ActiveStatus {
				t.Errorf("own account changed: %+v", u)
			}
		})
	}
}

// The rule targets only the own account: the admin keeps editing its own name and manages every other user.
func TestCommand_ErlaubtEigeneDetailsUndAndereKonten(t *testing.T) {
	c := commandMitAdmin()
	ctx := context.Background()

	if err := c.UpdateUser(ctx, adminID, adminID, "Anna Neu", "annaneu", user.AdminRole); err != nil {
		t.Fatalf("UpdateUser own details: %v", err)
	}
	if err := c.DeactivateUser(ctx, adminID, 7); err != nil {
		t.Fatalf("DeactivateUser other: %v", err)
	}
	if u, _ := c.UserRepo.GetUser(ctx, 7); u.Status != user.InactiveStatus {
		t.Errorf("expected the other user inactive, got %s", u.Status)
	}
	if err := c.DeleteUser(ctx, adminID, 7); err != nil {
		t.Fatalf("DeleteUser other: %v", err)
	}
}
