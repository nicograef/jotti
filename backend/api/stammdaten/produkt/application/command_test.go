package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/repository/repotest"
)

var testProdukt = produkt.Produkt{
	ID:         1,
	Name:       "Cola",
	Kategorie:  produkt.GetraenkKategorie,
	Steuersatz: steuer.RegelSteuersatz,
	Status:     produkt.ActiveStatus,
	Varianten:  []produkt.Variante{},
	CreatedAt:  time.Now().UTC(),
	UpdatedAt:  time.Now().UTC(),
}

func TestCreateProdukt(t *testing.T) {
	repo := repotest.NewProduktRepo(nil, nil)
	cmd := Command{ProduktRepo: repo}

	id, err := cmd.CreateProdukt(context.Background(), "Bier", produkt.GetraenkKategorie, steuer.RegelSteuersatz, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 1 {
		t.Errorf("expected produkt ID 1, got %d", id)
	}
}

func TestCreateProdukt_AlreadyExists(t *testing.T) {
	repo := repotest.NewProduktRepo(nil, db.ErrAlreadyExists)
	cmd := Command{ProduktRepo: repo}

	_, err := cmd.CreateProdukt(context.Background(), "Bier", produkt.GetraenkKategorie, steuer.RegelSteuersatz, false)
	if !errors.Is(err, ErrProduktAlreadyExists) {
		t.Errorf("expected ErrProduktAlreadyExists, got %v", err)
	}
}

func TestUpdateProdukt(t *testing.T) {
	repo := repotest.NewProduktRepo([]produkt.Produkt{testProdukt}, nil)
	cmd := Command{ProduktRepo: repo}

	err := cmd.UpdateProdukt(context.Background(), 1, "Fanta", produkt.GetraenkKategorie, steuer.RegelSteuersatz, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updated, err := repo.GetProdukt(context.Background(), 1)
	if err != nil {
		t.Fatalf("failed to get updated produkt: %v", err)
	}
	if updated.Name != "Fanta" {
		t.Errorf("expected produkt name 'Fanta', got '%s'", updated.Name)
	}
}

func TestUpdateProdukt_NotFound(t *testing.T) {
	repo := repotest.NewProduktRepo(nil, db.ErrNotFound)
	cmd := Command{ProduktRepo: repo}

	err := cmd.UpdateProdukt(context.Background(), 999, "Fanta", produkt.GetraenkKategorie, steuer.RegelSteuersatz, false)
	if !errors.Is(err, ErrProduktNotFound) {
		t.Errorf("expected ErrProduktNotFound, got %v", err)
	}
}

func TestUpdateProdukt_AlreadyExists(t *testing.T) {
	repo := repotest.NewProduktRepo([]produkt.Produkt{testProdukt}, nil)
	repo.SetUpdateProduktError(db.ErrAlreadyExists)
	cmd := Command{ProduktRepo: repo}

	err := cmd.UpdateProdukt(context.Background(), 1, "Fanta", produkt.GetraenkKategorie, steuer.RegelSteuersatz, false)
	if !errors.Is(err, ErrProduktAlreadyExists) {
		t.Errorf("expected ErrProduktAlreadyExists, got %v", err)
	}
}

func TestVerschiebeProdukt_NotFound(t *testing.T) {
	repo := repotest.NewProduktRepo(nil, db.ErrNotFound)
	cmd := Command{ProduktRepo: repo}

	err := cmd.VerschiebeProdukt(context.Background(), 999, produkt.RichtungHoch)
	if !errors.Is(err, ErrProduktNotFound) {
		t.Errorf("expected ErrProduktNotFound, got %v", err)
	}
}

func TestVerschiebeVariante_NotFound(t *testing.T) {
	repo := repotest.NewProduktRepo(nil, db.ErrNotFound)
	cmd := Command{ProduktRepo: repo}

	err := cmd.VerschiebeVariante(context.Background(), 999, produkt.RichtungRunter)
	if !errors.Is(err, ErrVarianteNotFound) {
		t.Errorf("expected ErrVarianteNotFound, got %v", err)
	}
}

// testVarianteVon baut ein Produkt mit genau einer Variante — so, wie GetProdukt
// es liefert: die Variantenliste des Produkts und die Variante selbst.
func testVarianteVon(id int) (produkt.Produkt, produkt.Variante) {
	variante := produkt.Variante{
		ID:         1,
		Name:       "0,5l",
		PreisCents: 350,
		Status:     produkt.ActiveStatus,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	p := testProdukt
	p.ID = id
	p.Varianten = []produkt.Variante{variante}

	return p, variante
}

func TestDeleteVariante(t *testing.T) {
	eigenes, variante := testVarianteVon(1)
	repo := repotest.NewProduktRepo([]produkt.Produkt{eigenes}, nil)
	repo.AddVariante(eigenes.ID, variante)
	cmd := Command{ProduktRepo: repo}

	if err := cmd.DeleteVariante(context.Background(), eigenes.ID, variante.ID); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	geloescht, err := repo.GetVariante(context.Background(), variante.ID)
	if err != nil {
		t.Fatalf("failed to get variante after delete: %v", err)
	}
	if geloescht.Status != produkt.DeletedStatus {
		t.Errorf("expected status deleted, got %q", geloescht.Status)
	}
}

// Eine Variante gehört genau einem Produkt. Nennt der Aufruf ein fremdes
// Produkt, wird nichts gelöscht.
func TestDeleteVariante_FremdeVariante(t *testing.T) {
	eigenes, variante := testVarianteVon(1)
	fremdes := testProdukt
	fremdes.ID = 2
	fremdes.Name = "Wasser"
	fremdes.Varianten = []produkt.Variante{}

	repo := repotest.NewProduktRepo([]produkt.Produkt{eigenes, fremdes}, nil)
	repo.AddVariante(eigenes.ID, variante)
	cmd := Command{ProduktRepo: repo}

	err := cmd.DeleteVariante(context.Background(), fremdes.ID, variante.ID)
	if !errors.Is(err, ErrVarianteNotFound) {
		t.Errorf("expected ErrVarianteNotFound, got %v", err)
	}

	unberuehrt, err := repo.GetVariante(context.Background(), variante.ID)
	if err != nil {
		t.Fatalf("failed to get variante after rejected delete: %v", err)
	}
	if unberuehrt.Status != produkt.ActiveStatus {
		t.Errorf("expected status active, got %q", unberuehrt.Status)
	}
}

func TestDeleteProdukt(t *testing.T) {
	repo := repotest.NewProduktRepo([]produkt.Produkt{testProdukt}, nil)
	cmd := Command{ProduktRepo: repo}

	err := cmd.DeleteProdukt(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	deleted, err := repo.GetProdukt(context.Background(), 1)
	if err != nil {
		t.Fatalf("failed to get produkt after delete: %v", err)
	}
	if deleted.Status != produkt.DeletedStatus {
		t.Errorf("expected status deleted, got %q", deleted.Status)
	}
}

// Only the active Varianten change; an inactive or deleted one keeps its status.
func TestDeactivateAlleVarianten(t *testing.T) {
	varianten := []produkt.Variante{
		{ID: 1, Name: "0,3 l", PreisCents: 250, Status: produkt.ActiveStatus},
		{ID: 2, Name: "0,5 l", PreisCents: 350, Status: produkt.ActiveStatus},
		{ID: 3, Name: "1,0 l", PreisCents: 600, Status: produkt.DeletedStatus},
	}
	mitVarianten := testProdukt
	mitVarianten.Varianten = varianten
	repo := repotest.NewProduktRepo([]produkt.Produkt{mitVarianten}, nil)
	for _, v := range varianten {
		repo.AddVariante(1, v)
	}

	if err := (Command{ProduktRepo: repo}).DeactivateAlleVarianten(context.Background(), 1); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for id, want := range map[int]produkt.Status{1: produkt.InactiveStatus, 2: produkt.InactiveStatus, 3: produkt.DeletedStatus} {
		v, err := repo.GetVariante(context.Background(), id)
		if err != nil {
			t.Fatalf("failed to get variante %d: %v", id, err)
		}
		if v.Status != want {
			t.Errorf("variante %d: expected status %q, got %q", id, want, v.Status)
		}
	}
}

// schreibZaehler counts the Varianten writes, so a rejected Pfand-Verknüpfung can prove it wrote nothing.
type schreibZaehler struct {
	*repotest.ProduktRepo
	schreibvorgaenge int
}

func (s *schreibZaehler) CreateVariante(ctx context.Context, produktID int, v produkt.Variante) (int, error) {
	s.schreibvorgaenge++
	return s.ProduktRepo.CreateVariante(ctx, produktID, v)
}

func (s *schreibZaehler) UpdateVariante(ctx context.Context, v produkt.Variante) error {
	s.schreibvorgaenge++
	return s.ProduktRepo.UpdateVariante(ctx, v)
}

func TestVariantePfandVerknuepfung(t *testing.T) {
	bierVariante := produkt.Variante{ID: 10, Name: "0,5 l", PreisCents: 400, Status: produkt.ActiveStatus}
	glas := produkt.Variante{ID: 20, Name: "Glas", PreisCents: 200, Status: produkt.ActiveStatus}
	teller := produkt.Variante{ID: 30, Name: "Teller", PreisCents: 300, Status: produkt.ActiveStatus}
	bier := produkt.Produkt{ID: 1, Name: "Bier", Steuersatz: steuer.RegelSteuersatz, Status: produkt.ActiveStatus, Varianten: []produkt.Variante{bierVariante}}
	glasPfand := produkt.Produkt{ID: 2, Name: "Glaspfand", Steuersatz: steuer.RegelSteuersatz, Pfand: true, Status: produkt.ActiveStatus, Varianten: []produkt.Variante{glas}}
	tellerPfand := produkt.Produkt{ID: 3, Name: "Tellerpfand", Steuersatz: steuer.ErmaessigtSteuersatz, Pfand: true, Status: produkt.ActiveStatus, Varianten: []produkt.Variante{teller}}

	ptr := func(id int) *int { return &id }
	cases := []struct {
		name    string
		aendern bool
		pfand   int
		wantErr error
	}{
		{"anlegen mit gültigem Pfand", false, glas.ID, nil},
		{"anlegen mit Steuersatzbruch", false, teller.ID, ErrPfandSteuersatzAbweichend},
		{"ändern mit gültigem Pfand", true, glas.ID, nil},
		{"ändern mit Steuersatzbruch", true, teller.ID, ErrPfandSteuersatzAbweichend},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			repo := &schreibZaehler{ProduktRepo: repotest.NewProduktRepo([]produkt.Produkt{bier, glasPfand, tellerPfand}, nil)}
			for _, p := range []produkt.Produkt{bier, glasPfand, tellerPfand} {
				repo.AddVariante(p.ID, p.Varianten[0])
			}
			cmd := Command{ProduktRepo: repo}

			varianteID := bierVariante.ID
			var err error
			if tc.aendern {
				err = cmd.UpdateVariante(ctx, varianteID, "0,5 l", 400, ptr(tc.pfand))
			} else {
				varianteID, err = cmd.CreateVariante(ctx, bier.ID, "0,3 l", 300, ptr(tc.pfand))
			}

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
			if tc.wantErr != nil {
				if repo.schreibvorgaenge != 0 {
					t.Errorf("expected no write, got %d", repo.schreibvorgaenge)
				}
				return
			}
			gespeichert, err := repo.GetVariante(ctx, varianteID)
			if err != nil {
				t.Fatalf("failed to get variante %d: %v", varianteID, err)
			}
			if gespeichert.PfandVarianteID == nil || *gespeichert.PfandVarianteID != tc.pfand {
				t.Errorf("expected PfandVarianteID %d, got %v", tc.pfand, gespeichert.PfandVarianteID)
			}
		})
	}
}

func (s *schreibZaehler) UpdateProdukt(ctx context.Context, p produkt.Produkt) error {
	s.schreibvorgaenge++
	return s.ProduktRepo.UpdateProdukt(ctx, p)
}

func TestUpdateProdukt_PfandVerknuepfung(t *testing.T) {
	ptr := func(id int) *int { return &id }
	glas := produkt.Variante{ID: 20, Name: "Glas", PreisCents: 200, Status: produkt.ActiveStatus}
	glasPfand := produkt.Produkt{ID: 2, Name: "Glaspfand", Kategorie: produkt.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Pfand: true, Status: produkt.ActiveStatus, Varianten: []produkt.Variante{glas}}
	bierMit := func(pfandVarianteID *int) produkt.Produkt {
		return produkt.Produkt{ID: 1, Name: "Bier", Kategorie: produkt.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Status: produkt.ActiveStatus,
			Varianten: []produkt.Variante{{ID: 10, Name: "0,5 l", PreisCents: 400, Status: produkt.ActiveStatus, PfandVarianteID: pfandVarianteID}}}
	}

	cases := []struct {
		name       string
		verknuepft bool
		produkt    produkt.Produkt
		steuersatz steuer.Steuersatz
		pfand      bool
		wantErr    error
	}{
		{"Steuersatzwechsel der Quelle", true, bierMit(nil), steuer.ErmaessigtSteuersatz, false, ErrPfandSteuersatzAbweichend},
		{"Steuersatzwechsel des Pfands", true, glasPfand, steuer.ErmaessigtSteuersatz, true, ErrPfandSteuersatzAbweichend},
		{"Kennzeichen aus", true, glasPfand, steuer.RegelSteuersatz, false, ErrPfandNochVerknuepft},
		{"Kennzeichen an", true, bierMit(nil), steuer.RegelSteuersatz, true, ErrPfandVerkettet},
		{"Steuersatzwechsel ohne Verknüpfung", false, bierMit(nil), steuer.ErmaessigtSteuersatz, false, nil},
		{"Kennzeichen aus ohne Verknüpfung", false, glasPfand, steuer.RegelSteuersatz, false, nil},
		{"Kennzeichen an ohne Verknüpfung", false, bierMit(nil), steuer.RegelSteuersatz, true, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			bier := bierMit(nil)
			if tc.verknuepft {
				bier = bierMit(ptr(glas.ID))
			}
			repo := &schreibZaehler{ProduktRepo: repotest.NewProduktRepo([]produkt.Produkt{bier, glasPfand}, nil)}
			cmd := Command{ProduktRepo: repo}

			err := cmd.UpdateProdukt(ctx, tc.produkt.ID, tc.produkt.Name, tc.produkt.Kategorie, tc.steuersatz, tc.pfand)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
			gespeichert, err := repo.GetProdukt(ctx, tc.produkt.ID)
			if err != nil {
				t.Fatalf("failed to get produkt %d: %v", tc.produkt.ID, err)
			}
			if tc.wantErr != nil {
				if repo.schreibvorgaenge != 0 {
					t.Errorf("expected no write, got %d", repo.schreibvorgaenge)
				}
				return
			}
			if gespeichert.Steuersatz != tc.steuersatz || gespeichert.Pfand != tc.pfand {
				t.Errorf("expected Steuersatz %q and Pfand %v, got %q and %v", tc.steuersatz, tc.pfand, gespeichert.Steuersatz, gespeichert.Pfand)
			}
		})
	}
}
