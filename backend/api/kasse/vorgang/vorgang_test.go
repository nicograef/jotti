package vorgang

import (
	"encoding/json"
	"testing"

	"github.com/nicograef/jotti/backend/domain/kasse"
)

func TestToHistorieEintrag_Positionskommentar(t *testing.T) {
	t.Parallel()

	positionen := []kasse.Position{
		{PositionID: "123e4567-e89b-42d3-a456-426614174000", ProduktName: "Bratwurst", Menge: 2, Kommentar: "ohne Senf"},
		{PositionID: "123e4567-e89b-42d3-a456-426614174001", ProduktName: "Bier", Menge: 1},
	}
	raw, err := json.Marshal(ToHistorieEintrag(kasse.VorgangHistorieEintrag{Positionen: positionen, OffenePositionen: positionen}))
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var eintrag struct {
		Positionen       []map[string]any `json:"positionen"`
		OffenePositionen []map[string]any `json:"offenePositionen"`
	}
	if err := json.Unmarshal(raw, &eintrag); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if got := eintrag.Positionen[0]["kommentar"]; got != "ohne Senf" {
		t.Errorf("positionen[0].kommentar = %v, want %q", got, "ohne Senf")
	}
	if got, ok := eintrag.Positionen[1]["kommentar"]; !ok || got != "" {
		t.Errorf("positionen[1].kommentar = %v (present %v), want empty string", got, ok)
	}
	if got := eintrag.Positionen[0]["produktName"]; got != "Bratwurst" {
		t.Errorf("positionen[0].produktName = %v, want Bratwurst", got)
	}
	if _, ok := eintrag.OffenePositionen[0]["kommentar"]; ok {
		t.Error("offenePositionen[0] carries kommentar, want no key")
	}
}
