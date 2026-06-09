//go:build unit

package steuer

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestSteuersatzProzent(t *testing.T) {
	tests := []struct {
		name     string
		satz     Steuersatz
		expected int
	}{
		{name: "regel", satz: RegelSteuersatz, expected: 19},
		{name: "ermaessigt", satz: ErmaessigtSteuersatz, expected: 7},
		{name: "befreit", satz: BefreitSteuersatz, expected: 0},
		{name: "kombi has no single percent", satz: KombiSteuersatz, expected: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.satz.Prozent(); got != tt.expected {
				t.Fatalf("Prozent() = %d, expected %d", got, tt.expected)
			}
		})
	}
}

func TestAufteilen(t *testing.T) {
	tests := []struct {
		name     string
		brutto   int
		satz     Steuersatz
		expected []Aufteilung
	}{
		{
			name:   "regel 500ct",
			brutto: 500,
			satz:   RegelSteuersatz,
			expected: []Aufteilung{{
				Satz:   RegelSteuersatz,
				Brutto: 500,
				Netto:  420,
				Steuer: 80,
			}},
		},
		{
			name:   "ermaessigt 500ct",
			brutto: 500,
			satz:   ErmaessigtSteuersatz,
			expected: []Aufteilung{{
				Satz:   ErmaessigtSteuersatz,
				Brutto: 500,
				Netto:  467,
				Steuer: 33,
			}},
		},
		{
			name:   "befreit 500ct",
			brutto: 500,
			satz:   BefreitSteuersatz,
			expected: []Aufteilung{{
				Satz:   BefreitSteuersatz,
				Brutto: 500,
				Netto:  500,
				Steuer: 0,
			}},
		},
		{
			name:   "regel 0ct",
			brutto: 0,
			satz:   RegelSteuersatz,
			expected: []Aufteilung{{
				Satz:   RegelSteuersatz,
				Brutto: 0,
				Netto:  0,
				Steuer: 0,
			}},
		},
		{
			name:   "ermaessigt 1ct",
			brutto: 1,
			satz:   ErmaessigtSteuersatz,
			expected: []Aufteilung{{
				Satz:   ErmaessigtSteuersatz,
				Brutto: 1,
				Netto:  1,
				Steuer: 0,
			}},
		},
		{
			name:   "kombi 1ct",
			brutto: 1,
			satz:   KombiSteuersatz,
			expected: []Aufteilung{
				{Satz: ErmaessigtSteuersatz, Brutto: 1, Netto: 1, Steuer: 0},
				{Satz: RegelSteuersatz, Brutto: 0, Netto: 0, Steuer: 0},
			},
		},
		{
			name:   "kombi odd amount",
			brutto: 501,
			satz:   KombiSteuersatz,
			expected: []Aufteilung{
				{Satz: ErmaessigtSteuersatz, Brutto: 351, Netto: 328, Steuer: 23},
				{Satz: RegelSteuersatz, Brutto: 150, Netto: 126, Steuer: 24},
			},
		},
		{
			name:   "kombi half-up split boundary",
			brutto: 5,
			satz:   KombiSteuersatz,
			expected: []Aufteilung{
				{Satz: ErmaessigtSteuersatz, Brutto: 4, Netto: 4, Steuer: 0},
				{Satz: RegelSteuersatz, Brutto: 1, Netto: 1, Steuer: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Aufteilen(tt.brutto, tt.satz)

			if !reflect.DeepEqual(got, tt.expected) {
				t.Fatalf("Aufteilen(%d, %s) = %+v, expected %+v", tt.brutto, tt.satz, got, tt.expected)
			}

			for _, aufteilung := range got {
				if aufteilung.Netto+aufteilung.Steuer != aufteilung.Brutto {
					t.Fatalf("invariant broken for %+v", aufteilung)
				}
			}

			if tt.satz == KombiSteuersatz {
				sumBrutto := 0
				for _, aufteilung := range got {
					sumBrutto += aufteilung.Brutto
				}
				if sumBrutto != tt.brutto {
					t.Fatalf("kombi brutto split mismatch: got %d expected %d", sumBrutto, tt.brutto)
				}
			}
		})
	}
}

func TestSteuersatzSchema(t *testing.T) {
	valid := []Steuersatz{RegelSteuersatz, ErmaessigtSteuersatz, BefreitSteuersatz, KombiSteuersatz}
	for _, satz := range valid {
		t.Run("valid_"+string(satz), func(t *testing.T) {
			if issue := SteuersatzSchema.Validate(&satz); issue != nil {
				t.Fatalf("expected valid steuersatz %q, got issue %v", satz, issue)
			}
		})
	}

	t.Run("invalid message is german", func(t *testing.T) {
		satz := Steuersatz("falsch")
		issue := SteuersatzSchema.Validate(&satz)
		if issue == nil {
			t.Fatal("expected validation issue for invalid steuersatz")
		}

		issueText := fmt.Sprintf("%v", issue)
		if !strings.Contains(issueText, "Ungueltiger Steuersatz") {
			t.Fatalf("expected german validation message, got %q", issueText)
		}
	})
}
