package zeit

import (
	"testing"
	"time"
)

// Der Kalendertag ist das Wandkalenderdatum in Europe/Berlin — auch kurz nach
// Mitternacht (Sommer- wie Winterzeit), wo UTC noch der Vortag ist.
func TestKalendertag(t *testing.T) {
	cases := []struct {
		name string
		t    time.Time
		want time.Time
	}{
		{"Sommerzeit 00:30 Berlin (22:30 UTC Vortag)", time.Date(2026, 7, 4, 22, 30, 0, 0, time.UTC), time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)},
		{"Winterzeit 00:30 Berlin (23:30 UTC Vortag)", time.Date(2026, 1, 9, 23, 30, 0, 0, time.UTC), time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)},
		{"mittags", time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC), time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Kalendertag(tc.t); !got.Equal(tc.want) || got.Location() != time.UTC {
				t.Errorf("Kalendertag = %v, want %v", got, tc.want)
			}
		})
	}
}
