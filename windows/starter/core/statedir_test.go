package core

import "testing"

func TestStateDirOwnerAllowed(t *testing.T) {
	const user = "S-1-5-21-1-2-3-1001"
	const admin = "S-1-5-21-1-2-3-500"
	tests := []struct {
		name  string
		owner string
		users []string
		want  bool
	}{
		{"SYSTEM", "S-1-5-18", []string{user}, true},
		{"Administratoren", "S-1-5-32-544", []string{user}, true},
		{"angemeldeter Benutzer", user, []string{user, admin}, true},
		{"Prozessbenutzer nach fremder UAC-Abfrage", admin, []string{user, admin}, true},
		{"anderer lokaler Benutzer", "S-1-5-21-1-2-3-1002", []string{user, admin}, false},
		{"Gruppe Benutzer", "S-1-5-32-545", []string{user}, false},
		{"Jeder", "S-1-1-0", []string{user}, false},
		{"leerer Besitzer trifft keinen leeren Benutzer", "", []string{""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StateDirOwnerAllowed(tt.owner, tt.users); got != tt.want {
				t.Errorf("StateDirOwnerAllowed(%q, %q) = %v, want %v", tt.owner, tt.users, got, tt.want)
			}
		})
	}
}
