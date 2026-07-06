//go:build unit

package seed

import "testing"

func TestAllowedByEnv(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  bool
	}{
		{"exact 1 allows", "1", true},
		{"unset denies", "", false},
		{"zero denies", "0", false},
		{"true denies", "true", false},
		{"yes denies", "yes", false},
		{"padded denies", " 1", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			getenv := func(key string) string {
				if key != AllowSeedEnv {
					t.Fatalf("unexpected env lookup: %s", key)
				}
				return tc.value
			}
			if got := AllowedByEnv(getenv); got != tc.want {
				t.Errorf("AllowedByEnv(%q) = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}
