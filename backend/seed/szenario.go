package seed

import (
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/nicograef/jotti/backend/domain/steuer"
)

// demoArgon2idHash is the precomputed Argon2id hash of the demo login password
// "jotti123" (m=65536, t=2, p=2). The password hasher uses a random salt, so a
// fixed constant is required to keep the seed deterministic.
const demoArgon2idHash = "$argon2id$v=19$m=65536,t=2,p=2$OSImYG1ms0Phs26KwwMwkQ$rkoWKOIjsPz7y6ps/W2pVEhn5vTc0N95SyiveQCn404"

// --- Stammdaten-Typen ---

type benutzer struct {
	ID       int
	Name     string
	Username string
	Rolle    string
	Status   string
}

type tisch struct {
	ID     int
	Name   string
	Status string
}

type variante struct {
	ID         int
	Name       string
	PreisCents int
	Status     string
}

type produkt struct {
	ID         int
	Name       string
	Kategorie  product.Kategorie
	Steuersatz steuer.Steuersatz
	Status     string
	Varianten  []variante
}

type betreiber struct {
	Vereinsname  string
	Strasse      string
	Plz          string
	Ort          string
	Steuernummer *string
	UstID        *string
}

// --- Drehbuch-Typen ---

// bestellposten verweist auf eine Produktvariante und die bestellte Menge.
// Die Engine löst daraus die vollständige Position aus den Stammdaten auf.
type bestellposten struct {
	VarianteID int
	Menge      int
}

// tischVerlauf ist das minimale Phase-1-Drehbuch für einen Tisch: eine Bestellung,
// optional die Ausgabe-Bestätigung und optional die Zahlung der gesamten Bestellung.
type tischVerlauf struct {
	TischID       int
	ServiceUserID int
	Posten        []bestellposten
	Ausgegeben    bool
	Bezahlt       bool
}

// kassensitzungDrehbuch beschreibt eine Kassensitzung (Phase 1: die offene Sonntags-Sitzung).
type kassensitzungDrehbuch struct {
	ZNr                 int
	Bezeichnung         string
	EroeffnetVon        int
	AnfangsbestandCents int
	Tische              []tischVerlauf
}

// szenario bündelt die Stammdaten und das Drehbuch.
type szenario struct {
	Benutzer  []benutzer
	Tische    []tisch
	Produkte  []produkt
	Betreiber betreiber
	Sitzung   kassensitzungDrehbuch
}

func strPtr(s string) *string { return &s }

// phase1Szenario liefert die aktiven Stammdaten des „3-Tage-Sommerfest TSV Musterstadt e.V."
// sowie das minimale Drehbuch für Phase 1: eine offene Sonntags-Kassensitzung mit wenigen
// Bestell-/Ausgabe-/Zahlungs-Zyklen. Inaktiv-/Gelöscht-Beispiele, Favoriten und das volle
// 3-Tage-Drehbuch folgen in späteren Phasen.
func phase1Szenario() szenario {
	return szenario{
		Benutzer: []benutzer{
			{ID: 2, Name: "Thomas Müller", Username: "thomas", Rolle: "admin", Status: "active"},
			{ID: 3, Name: "Felix Weber", Username: "felix", Rolle: "serviceleitung", Status: "active"},
			{ID: 4, Name: "Maria Schmidt", Username: "maria", Rolle: "service", Status: "active"},
			{ID: 5, Name: "Lisa Braun", Username: "lisa", Rolle: "service", Status: "active"},
			{ID: 6, Name: "Jan Hoffmann", Username: "jan", Rolle: "service", Status: "active"},
			{ID: 7, Name: "Sophie Becker", Username: "sophie", Rolle: "serviceleitung", Status: "active"},
			{ID: 8, Name: "Markus Lehmann", Username: "markus", Rolle: "service", Status: "active"},
			{ID: 9, Name: "Anna Krause", Username: "anna", Rolle: "service", Status: "active"},
		},
		Tische: []tisch{
			{ID: 1, Name: "Tisch 1", Status: "active"},
			{ID: 2, Name: "Tisch 2", Status: "active"},
			{ID: 3, Name: "Tisch 3", Status: "active"},
			{ID: 4, Name: "Tisch 4", Status: "active"},
			{ID: 5, Name: "Tisch 5", Status: "active"},
			{ID: 6, Name: "Tisch 6", Status: "active"},
			{ID: 7, Name: "Tisch 7", Status: "active"},
			{ID: 8, Name: "Tisch 8", Status: "active"},
			{ID: 9, Name: "Tisch 9", Status: "active"},
			{ID: 10, Name: "Tisch 10", Status: "active"},
			{ID: 11, Name: "Tisch 11", Status: "active"},
			{ID: 12, Name: "Tisch 12", Status: "active"},
			{ID: 13, Name: "Tisch 13", Status: "active"},
			{ID: 14, Name: "Tisch 14", Status: "active"},
			{ID: 15, Name: "Tisch 15", Status: "active"},
			{ID: 16, Name: "Zelt A1", Status: "active"},
			{ID: 17, Name: "Zelt A2", Status: "active"},
			{ID: 18, Name: "Stehtisch Bar", Status: "active"},
			{ID: 19, Name: "Stehtisch Eingang", Status: "active"},
			{ID: 20, Name: "Stehtisch Terrasse", Status: "active"},
		},
		Produkte: []produkt{
			{ID: 1, Name: "Bratwurst", Kategorie: product.EssenKategorie, Steuersatz: steuer.ErmaessigtSteuersatz, Status: "active", Varianten: []variante{
				{ID: 1, Name: "Normal", PreisCents: 350, Status: "active"},
				{ID: 2, Name: "XXL", PreisCents: 500, Status: "active"},
				{ID: 3, Name: "Currywurst", PreisCents: 450, Status: "active"},
			}},
			{ID: 2, Name: "Pommes", Kategorie: product.EssenKategorie, Steuersatz: steuer.ErmaessigtSteuersatz, Status: "active", Varianten: []variante{
				{ID: 4, Name: "Klein", PreisCents: 250, Status: "active"},
				{ID: 5, Name: "Groß", PreisCents: 350, Status: "active"},
			}},
			{ID: 3, Name: "Flammkuchen", Kategorie: product.EssenKategorie, Steuersatz: steuer.ErmaessigtSteuersatz, Status: "active", Varianten: []variante{
				{ID: 6, Name: "Classic", PreisCents: 600, Status: "active"},
				{ID: 7, Name: "Speck & Zwiebel", PreisCents: 700, Status: "active"},
				{ID: 8, Name: "Mediterran", PreisCents: 750, Status: "active"},
			}},
			{ID: 4, Name: "Tagesgericht", Kategorie: product.EssenKategorie, Steuersatz: steuer.ErmaessigtSteuersatz, Status: "active", Varianten: []variante{
				{ID: 9, Name: "Fr: Schnitzel mit Pommes", PreisCents: 1250, Status: "active"},
				{ID: 10, Name: "Sa: Gulasch mit Spätzle", PreisCents: 1150, Status: "active"},
				{ID: 11, Name: "So: Hähnchen mit Reis", PreisCents: 1050, Status: "active"},
			}},
			{ID: 5, Name: "Grillplatte", Kategorie: product.EssenKategorie, Steuersatz: steuer.ErmaessigtSteuersatz, Status: "active", Varianten: []variante{
				{ID: 12, Name: "Klein", PreisCents: 800, Status: "active"},
				{ID: 13, Name: "Groß", PreisCents: 1400, Status: "active"},
			}},
			{ID: 6, Name: "Salat", Kategorie: product.EssenKategorie, Steuersatz: steuer.ErmaessigtSteuersatz, Status: "active", Varianten: []variante{
				{ID: 14, Name: "Gemischter Salat", PreisCents: 550, Status: "active"},
				{ID: 15, Name: "Caesar Salat", PreisCents: 650, Status: "active"},
			}},
			{ID: 7, Name: "Kuchen", Kategorie: product.EssenKategorie, Steuersatz: steuer.ErmaessigtSteuersatz, Status: "active", Varianten: []variante{
				{ID: 16, Name: "Stück", PreisCents: 250, Status: "active"},
			}},
			{ID: 8, Name: "Waffeln", Kategorie: product.EssenKategorie, Steuersatz: steuer.ErmaessigtSteuersatz, Status: "active", Varianten: []variante{
				{ID: 17, Name: "mit Puderzucker", PreisCents: 300, Status: "active"},
				{ID: 18, Name: "mit Sahne", PreisCents: 350, Status: "active"},
				{ID: 19, Name: "mit Nutella", PreisCents: 400, Status: "active"},
			}},
			{ID: 9, Name: "Brezel", Kategorie: product.EssenKategorie, Steuersatz: steuer.ErmaessigtSteuersatz, Status: "active", Varianten: []variante{
				{ID: 20, Name: "Normal", PreisCents: 200, Status: "active"},
				{ID: 21, Name: "mit Butter", PreisCents: 300, Status: "active"},
			}},
			{ID: 11, Name: "Bier", Kategorie: product.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Status: "active", Varianten: []variante{
				{ID: 23, Name: "0,3l", PreisCents: 300, Status: "active"},
				{ID: 24, Name: "0,5l", PreisCents: 450, Status: "active"},
				{ID: 25, Name: "Maß 1,0l", PreisCents: 850, Status: "active"},
			}},
			{ID: 12, Name: "Weizen", Kategorie: product.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Status: "active", Varianten: []variante{
				{ID: 26, Name: "Klein 0,3l", PreisCents: 300, Status: "active"},
				{ID: 27, Name: "Groß 0,5l", PreisCents: 400, Status: "active"},
				{ID: 28, Name: "Colaweizen Klein", PreisCents: 300, Status: "active"},
				{ID: 29, Name: "Colaweizen Groß", PreisCents: 400, Status: "active"},
				{ID: 30, Name: "Russ", PreisCents: 300, Status: "active"},
			}},
			{ID: 13, Name: "Softdrinks", Kategorie: product.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Status: "active", Varianten: []variante{
				{ID: 31, Name: "Cola", PreisCents: 280, Status: "active"},
				{ID: 32, Name: "Fanta", PreisCents: 280, Status: "active"},
				{ID: 33, Name: "Spezi", PreisCents: 280, Status: "active"},
				{ID: 34, Name: "Sprite", PreisCents: 280, Status: "active"},
				{ID: 35, Name: "Mezzo Mix", PreisCents: 280, Status: "active"},
			}},
			{ID: 14, Name: "Wasser", Kategorie: product.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Status: "active", Varianten: []variante{
				{ID: 36, Name: "Still 0,5l", PreisCents: 200, Status: "active"},
				{ID: 37, Name: "Medium 0,5l", PreisCents: 200, Status: "active"},
				{ID: 38, Name: "Sprudel 0,5l", PreisCents: 200, Status: "active"},
			}},
			{ID: 15, Name: "Saftschorle", Kategorie: product.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Status: "active", Varianten: []variante{
				{ID: 39, Name: "Apfelschorle 0,5l", PreisCents: 300, Status: "active"},
				{ID: 40, Name: "Johannisbeerschorle 0,5l", PreisCents: 300, Status: "active"},
				{ID: 41, Name: "Rhabarberschorle 0,5l", PreisCents: 350, Status: "active"},
			}},
			{ID: 16, Name: "Wein", Kategorie: product.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Status: "active", Varianten: []variante{
				{ID: 42, Name: "Weißwein 0,2l", PreisCents: 400, Status: "active"},
				{ID: 43, Name: "Rotwein 0,2l", PreisCents: 400, Status: "active"},
				{ID: 44, Name: "Rosé 0,2l", PreisCents: 400, Status: "active"},
			}},
			{ID: 17, Name: "Kaffee", Kategorie: product.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Status: "active", Varianten: []variante{
				{ID: 45, Name: "Tasse", PreisCents: 200, Status: "active"},
				{ID: 46, Name: "Espresso", PreisCents: 180, Status: "active"},
			}},
			{ID: 18, Name: "Tee", Kategorie: product.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Status: "active", Varianten: []variante{
				{ID: 47, Name: "Verschiedene Sorten", PreisCents: 200, Status: "active"},
			}},
			{ID: 19, Name: "Hugo/Aperol", Kategorie: product.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Status: "active", Varianten: []variante{
				{ID: 48, Name: "Hugo", PreisCents: 550, Status: "active"},
				{ID: 49, Name: "Aperol Spritz", PreisCents: 550, Status: "active"},
			}},
			{ID: 21, Name: "Festbändchen", Kategorie: product.SonstigesKategorie, Steuersatz: steuer.BefreitSteuersatz, Status: "active", Varianten: []variante{
				{ID: 51, Name: "Erwachsene", PreisCents: 500, Status: "active"},
				{ID: 52, Name: "Kinder", PreisCents: 300, Status: "active"},
			}},
		},
		Betreiber: betreiber{
			Vereinsname:  "TSV Musterstadt e.V.",
			Strasse:      "Sportplatzweg 7",
			Plz:          "63776",
			Ort:          "Musterstadt",
			Steuernummer: strPtr("204/123/45678"),
			UstID:        nil,
		},
		Sitzung: kassensitzungDrehbuch{
			ZNr:                 3,
			Bezeichnung:         "Sommerfest 26 Sonntag",
			EroeffnetVon:        2, // Thomas Müller (admin)
			AnfangsbestandCents: 20000,
			Tische: []tischVerlauf{
				{
					// Tisch 1: bestellt, ausgegeben und bezahlt (abgeschlossen, Saldo 0).
					TischID:       1,
					ServiceUserID: 4, // Maria
					Posten: []bestellposten{
						{VarianteID: 1, Menge: 2},  // Bratwurst Normal
						{VarianteID: 24, Menge: 2}, // Bier 0,5l
					},
					Ausgegeben: true,
					Bezahlt:    true,
				},
				{
					// Tisch 2: bestellt und ausgegeben, aber noch offen (unbezahlt).
					TischID:       2,
					ServiceUserID: 6, // Jan
					Posten: []bestellposten{
						{VarianteID: 5, Menge: 1},  // Pommes Groß
						{VarianteID: 31, Menge: 2}, // Cola
					},
					Ausgegeben: true,
					Bezahlt:    false,
				},
				{
					// Tisch 3: frisch bestellt, noch nicht ausgegeben.
					TischID:       3,
					ServiceUserID: 5, // Lisa
					Posten: []bestellposten{
						{VarianteID: 6, Menge: 1}, // Flammkuchen Classic
					},
					Ausgegeben: false,
					Bezahlt:    false,
				},
			},
		},
	}
}
