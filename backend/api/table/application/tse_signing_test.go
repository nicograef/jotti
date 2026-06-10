//go:build unit

package application

import (
	"encoding/json"
	"testing"

	e "github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

func TestBuildKassenbelegProcessData_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		positionen  []kasse.Position
		zahlbetrag  int
		expected    string
		expectError bool
	}{
		{
			name:       "regelsteuersatz",
			positionen: []kasse.Position{{Einzelpreis: 1000, Menge: 1, Steuersatz: "regel"}},
			zahlbetrag: 1000,
			expected:   "Beleg^10.00_0.00_0.00_0.00_0.00^10.00:Bar",
		},
		{
			name: "ermaessigt und befreit",
			positionen: []kasse.Position{
				{Einzelpreis: 300, Menge: 1, Steuersatz: "ermaessigt"},
				{Einzelpreis: 200, Menge: 1, Steuersatz: "befreit"},
			},
			zahlbetrag: 500,
			expected:   "Beleg^0.00_3.00_0.00_0.00_2.00^5.00:Bar",
		},
		{
			name:       "kombi wird 70 30 aufgeteilt",
			positionen: []kasse.Position{{Einzelpreis: 1000, Menge: 1, Steuersatz: "kombi"}},
			zahlbetrag: 1000,
			expected:   "Beleg^3.00_7.00_0.00_0.00_0.00^10.00:Bar",
		},
		{
			name:       "ohne tausendertrennzeichen",
			positionen: []kasse.Position{{Einzelpreis: 123456, Menge: 1, Steuersatz: "regel"}},
			zahlbetrag: 123456,
			expected:   "Beleg^1234.56_0.00_0.00_0.00_0.00^1234.56:Bar",
		},
		{
			name:        "ungueltiger steuersatz",
			positionen:  []kasse.Position{{Einzelpreis: 100, Menge: 1, Steuersatz: "foo"}},
			zahlbetrag:  100,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildKassenbelegProcessData(tc.positionen, tc.zahlbetrag)
			if tc.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if got != tc.expected {
				t.Fatalf("unexpected processData\nwant: %q\ngot:  %q", tc.expected, got)
			}
		})
	}
}

func TestBuildKassenbelegProcessDataWithFaktor_NegativBeiStorno(t *testing.T) {
	got, err := buildKassenbelegProcessDataWithFaktor(
		[]kasse.Position{{Einzelpreis: 350, Menge: 2, Steuersatz: "regel"}},
		-700,
		-1,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := "Beleg^-7.00_0.00_0.00_0.00_0.00^-7.00:Bar"
	if got != want {
		t.Fatalf("unexpected processData\nwant: %q\ngot:  %q", want, got)
	}
}

func TestBuildBestellungProcessData(t *testing.T) {
	got, err := buildBestellungProcessData([]kasse.Position{
		{ProduktName: "Ma\u00df Bier", VarianteName: "", Menge: 4},
		{ProduktName: "Wei\u00dfwurst", VarianteName: "normal", Menge: 2},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := "4x Ma\u00df Bier_2x Wei\u00dfwurst normal"
	if got != want {
		t.Fatalf("unexpected processData\nwant: %q\ngot:  %q", want, got)
	}
}

func TestTSETransactionIDForZahlungEvent_Deterministic(t *testing.T) {
	event := e.Event{
		Type:    string(kasse.EventTypeZahlungKassiertV1),
		Subject: "kassensitzung-1/tisch-7",
		Data:    []byte(`{"zahlungId":"11111111-1111-1111-1111-111111111111","positionen":[],"gesamtZahlungCents":100,"kommentar":""}`),
	}

	first, err := tseTransactionIDForZahlungEvent(event)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	second, err := tseTransactionIDForZahlungEvent(event)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if first != second {
		t.Fatalf("expected deterministic tx_id, got %q and %q", first, second)
	}

	var data zahlungKassiertV1Data
	if err := json.Unmarshal(event.Data, &data); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}
	data.ZahlungID = "22222222-2222-2222-2222-222222222222"
	changed, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("expected no marshal error, got %v", err)
	}
	event.Data = changed

	third, err := tseTransactionIDForZahlungEvent(event)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if third == first {
		t.Fatalf("expected tx_id to change when event identity changes, still got %q", third)
	}
}
