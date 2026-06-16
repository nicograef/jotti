package core

import "testing"

func TestIsNewerVersion(t *testing.T) {
	cases := []struct {
		name    string
		current string
		latest  string
		want    bool
	}{
		{"neuer Patch", "v1.2.3", "v1.2.4", true},
		{"neuer Minor", "v1.2.3", "v1.3.0", true},
		{"neuer Major", "v1.2.3", "v2.0.0", true},
		{"gleiche Version", "v1.2.3", "v1.2.3", false},
		{"aeltere Version", "v1.3.0", "v1.2.9", false},
		{"ohne v-Praefix", "1.2.3", "1.2.4", true},
		{"gemischtes Praefix", "1.2.3", "v1.2.4", true},
		{"latest mit Vorabversion", "v1.2.3", "v1.2.4-rc1", true},
		{"Dev-Build wird nie gemeldet", "dev", "v1.2.3", false},
		{"Dev-mit-Sha wird nie gemeldet", "dev-abc1234", "v1.2.3", false},
		{"unlesbares latest meldet nicht", "v1.2.3", "latest", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsNewerVersion(tc.current, tc.latest); got != tc.want {
				t.Fatalf("IsNewerVersion(%q, %q): got %v, want %v",
					tc.current, tc.latest, got, tc.want)
			}
		})
	}
}

func TestParseLatestRelease(t *testing.T) {
	tag, err := ParseLatestRelease([]byte(`{"tag_name":"v1.4.0","name":"Release 1.4.0"}`))
	if err != nil {
		t.Fatalf("unerwarteter Fehler: %v", err)
	}
	if tag != "v1.4.0" {
		t.Fatalf("tag_name falsch geparst: got %q, want %q", tag, "v1.4.0")
	}
}

func TestParseLatestReleaseRejectsEmptyTag(t *testing.T) {
	if _, err := ParseLatestRelease([]byte(`{"name":"ohne tag"}`)); err == nil {
		t.Fatal("fehlender tag_name muss einen Fehler liefern")
	}
}

func TestParseLatestReleaseRejectsInvalidJSON(t *testing.T) {
	if _, err := ParseLatestRelease([]byte(`kein json`)); err == nil {
		t.Fatal("ungueltiges JSON muss einen Fehler liefern")
	}
}
