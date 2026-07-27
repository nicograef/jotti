//go:build unit

package application

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/tisch_repo"
	"github.com/rs/zerolog"
)

func TestGetTischState(t *testing.T) {
	positions := []kasse.Position{
		{PositionID: "p1", ProduktName: "Cola", VarianteName: "0,5l", EinzelpreisCents: 350, Menge: 2, BestellerUserID: 5, BestellerName: "Anna"},
	}
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	sitzungMock := kassensitzungen_repo.NewMock(&kasse.Kassensitzung{ZNr: 1, Status: kasse.KassensitzungOffen}, nil)
	subject := kasse.TischSessionSubject(1, 1)
	eventMock.SetTischSession(subject, kasse.TischSession{
		SaldoCents:           700,
		UnbezahltePositionen: positions,
		GesamtZahlungenCents: 0,
	})
	query := Query{
		TischRepo:           tisch_repo.NewMock([]tisch.Tisch{{ID: 1, Name: "Tisch 1", Status: tisch.ActiveStatus}}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: sitzungMock,
	}

	state, err := query.GetTischState(context.Background(), 1, 5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.SaldoCents != 700 {
		t.Errorf("expected saldo 700, got %d", state.SaldoCents)
	}
	if len(state.UnbezahltePositionen) != 1 {
		t.Fatalf("expected 1 unbezahlte position, got %d", len(state.UnbezahltePositionen))
	}
	if state.UnbezahltePositionen[0].Menge != 2 {
		t.Errorf("expected menge 2, got %d", state.UnbezahltePositionen[0].Menge)
	}
	if state.UnbezahltePositionen[0].BestellerName != "Anna" {
		t.Errorf("expected besteller Anna, got %q", state.UnbezahltePositionen[0].BestellerName)
	}
	// Anna hat eigene offene Positionen an diesem Tisch -> nicht erledigt.
	if state.FuerMichErledigt {
		t.Errorf("expected FuerMichErledigt false for besteller with open positions")
	}

	// Eine andere Servicekraft (ohne eigene Positionen) sieht den Tisch als erledigt.
	stateAndere, err := query.GetTischState(context.Background(), 1, 99)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !stateAndere.FuerMichErledigt {
		t.Errorf("expected FuerMichErledigt true for servicekraft without own positions")
	}
}

func TestGetTischState_NoState(t *testing.T) {
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	sitzungMock := kassensitzungen_repo.NewMock(&kasse.Kassensitzung{ZNr: 1, Status: kasse.KassensitzungOffen}, nil)
	query := Query{
		TischRepo:           tisch_repo.NewMock([]tisch.Tisch{{ID: 999, Name: "Tisch 999", Status: tisch.ActiveStatus}}, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: sitzungMock,
	}

	state, err := query.GetTischState(context.Background(), 999, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.SaldoCents != 0 {
		t.Errorf("expected saldo 0, got %d", state.SaldoCents)
	}
	if state.UnbezahltePositionen != nil {
		t.Errorf("expected nil unbezahlt, got %v", state.UnbezahltePositionen)
	}
}

type favoritMock struct {
	ids []int
	err error
}

func (m favoritMock) GetByUser(_ context.Context, _ int) ([]int, error) {
	return m.ids, m.err
}

// GetMeineTischeState liest Name + Session der Favoriten über einen Batch (kein N+1)
// und liefert die Views in Favoriten-Reihenfolge; ein Favorit ohne Session erhält
// eine Null-Session.
func TestGetMeineTischeState_BatchInFavoriteOrder(t *testing.T) {
	positions := []kasse.Position{
		{PositionID: "p1", ProduktName: "Cola", VarianteName: "0,5l", EinzelpreisCents: 350, Menge: 2, BestellerUserID: 5, BestellerName: "Anna"},
	}
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	sitzungMock := kassensitzungen_repo.NewMock(&kasse.Kassensitzung{ZNr: 1, Status: kasse.KassensitzungOffen}, nil)
	eventMock.SetTischName(7, "Tisch 7")
	eventMock.SetTischName(3, "Tisch 3")
	// Nur Tisch 7 hat eine Session; Tisch 3 bleibt session-los.
	eventMock.SetTischSession(kasse.TischSessionSubject(1, 7), kasse.TischSession{
		SaldoCents:           700,
		UnbezahltePositionen: positions,
	})

	query := Query{
		TischRepo:           tisch_repo.NewMock(nil, nil),
		EventRepo:           eventMock,
		FavoritRepo:         favoritMock{ids: []int{7, 3}},
		KassensitzungenRepo: sitzungMock,
	}

	views, err := query.GetMeineTischeState(context.Background(), 99)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(views) != 2 {
		t.Fatalf("expected 2 views, got %d", len(views))
	}

	// Reihenfolge folgt der Favoritenliste [7, 3].
	if views[0].TischID != 7 || views[0].TischName != "Tisch 7" {
		t.Errorf("expected first view tisch 7 'Tisch 7', got %d %q", views[0].TischID, views[0].TischName)
	}
	if views[0].Subject != kasse.TischSessionSubject(1, 7) {
		t.Errorf("unexpected subject %q", views[0].Subject)
	}
	if views[0].SaldoCents != 700 || len(views[0].UnbezahltePositionen) != 1 {
		t.Errorf("expected session data for tisch 7, got saldo %d / %d positionen", views[0].SaldoCents, len(views[0].UnbezahltePositionen))
	}

	// Session-loser Favorit -> Null-Session.
	if views[1].TischID != 3 || views[1].TischName != "Tisch 3" {
		t.Errorf("expected second view tisch 3 'Tisch 3', got %d %q", views[1].TischID, views[1].TischName)
	}
	if views[1].SaldoCents != 0 || views[1].UnbezahltePositionen != nil {
		t.Errorf("expected zero-value session for tisch 3, got saldo %d / %v", views[1].SaldoCents, views[1].UnbezahltePositionen)
	}
}

// Ein Favorit, dessen Tisch gelöscht oder unbekannt ist, wird übersprungen; die
// übrigen Favoriten werden normal geliefert. Ein Abbruch nähme der Servicekraft
// wegen eines einzigen gelöschten Tisches ihre gesamte Tischübersicht.
func TestGetMeineTischeState_SkipsUnresolvableFavorit(t *testing.T) {
	var logbuf bytes.Buffer
	ctx := zerolog.New(&logbuf).WithContext(context.Background())

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	sitzungMock := kassensitzungen_repo.NewMock(&kasse.Kassensitzung{ZNr: 1, Status: kasse.KassensitzungOffen}, nil)
	// Tisch 42 fehlt im Batch (gelöscht); 7 und 3 sind auflösbar.
	eventMock.SetTischName(7, "Tisch 7")
	eventMock.SetTischName(3, "Tisch 3")

	query := Query{
		TischRepo:           tisch_repo.NewMock(nil, nil),
		EventRepo:           eventMock,
		FavoritRepo:         favoritMock{ids: []int{7, 42, 3}},
		KassensitzungenRepo: sitzungMock,
	}

	views, err := query.GetMeineTischeState(ctx, 99)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(views) != 2 {
		t.Fatalf("expected 2 views, got %d", len(views))
	}
	if views[0].TischID != 7 || views[1].TischID != 3 {
		t.Errorf("expected tische 7 and 3, got %d and %d", views[0].TischID, views[1].TischID)
	}

	var warnLines []string
	for _, line := range strings.Split(strings.TrimSpace(logbuf.String()), "\n") {
		if strings.Contains(line, `"level":"warn"`) && strings.Contains(line, "Skipped favorit") {
			warnLines = append(warnLines, line)
		}
	}
	if len(warnLines) != 1 {
		t.Fatalf("expected exactly 1 warn line, got %d: %s", len(warnLines), logbuf.String())
	}
	if !strings.Contains(warnLines[0], `"tisch_id":42`) || !strings.Contains(warnLines[0], `"user_id":99`) {
		t.Errorf("warn line must name tisch and user, got %s", warnLines[0])
	}
}

// Sind alle Favoriten verwaist, ist das Ergebnis eine leere Liste — kein Fehler.
func TestGetMeineTischeState_AllFavoritenUnresolvable(t *testing.T) {
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	sitzungMock := kassensitzungen_repo.NewMock(&kasse.Kassensitzung{ZNr: 1, Status: kasse.KassensitzungOffen}, nil)
	// Kein SetTischName -> beide Tische sind dem Batch unbekannt.

	query := Query{
		TischRepo:           tisch_repo.NewMock(nil, nil),
		EventRepo:           eventMock,
		FavoritRepo:         favoritMock{ids: []int{42, 43}},
		KassensitzungenRepo: sitzungMock,
	}

	views, err := query.GetMeineTischeState(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(views) != 0 {
		t.Fatalf("expected 0 views, got %d", len(views))
	}
}

func TestGetTischHistorie_ReturnsEmptyForTischWithNoEvents(t *testing.T) {
	eventMock := kassenjournal_repo.NewMock(nil, nil)
	sitzungMock := kassensitzungen_repo.NewMock(&kasse.Kassensitzung{ZNr: 1, Status: kasse.KassensitzungOffen}, nil)
	query := Query{
		TischRepo:           tisch_repo.NewMock(nil, nil),
		EventRepo:           eventMock,
		KassensitzungenRepo: sitzungMock,
	}

	historie, err := query.GetTischHistorie(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(historie) != 0 {
		t.Errorf("expected empty historie, got %d entries", len(historie))
	}
}
