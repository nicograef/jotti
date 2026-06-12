package seed

import (
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/nicograef/jotti/backend/domain/user"
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
	Rolle    user.Role
	Status   user.Status
}

type tisch struct {
	ID     int
	Name   string
	Status table.Status
}

type variante struct {
	ID         int
	Name       string
	PreisCents int
	Status     product.Status
}

type produkt struct {
	ID         int
	Name       string
	Kategorie  product.Kategorie
	Steuersatz steuer.Steuersatz
	Status     product.Status
	Varianten  []variante
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
	Betreiber settings.Betreiber
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
			{ID: 2, Name: "Thomas Müller", Username: "thomas", Rolle: user.AdminRole, Status: user.ActiveStatus},
			{ID: 3, Name: "Felix Weber", Username: "felix", Rolle: user.ServiceleitungRole, Status: user.ActiveStatus},
			{ID: 4, Name: "Maria Schmidt", Username: "maria", Rolle: user.ServiceRole, Status: user.ActiveStatus},
			{ID: 5, Name: "Lisa Braun", Username: "lisa", Rolle: user.ServiceRole, Status: user.ActiveStatus},
			{ID: 6, Name: "Jan Hoffmann", Username: "jan", Rolle: user.ServiceRole, Status: user.ActiveStatus},
			{ID: 7, Name: "Sophie Becker", Username: "sophie", Rolle: user.ServiceleitungRole, Status: user.ActiveStatus},
			{ID: 8, Name: "Markus Lehmann", Username: "markus", Rolle: user.ServiceRole, Status: user.ActiveStatus},
			{ID: 9, Name: "Anna Krause", Username: "anna", Rolle: user.ServiceRole, Status: user.ActiveStatus},
		},
		Tische: []tisch{
			{ID: 1, Name: "Tisch 1", Status: table.ActiveStatus},
			{ID: 2, Name: "Tisch 2", Status: table.ActiveStatus},
			{ID: 3, Name: "Tisch 3", Status: table.ActiveStatus},
			{ID: 4, Name: "Tisch 4", Status: table.ActiveStatus},
			{ID: 5, Name: "Tisch 5", Status: table.ActiveStatus},
			{ID: 6, Name: "Tisch 6", Status: table.ActiveStatus},
			{ID: 7, Name: "Tisch 7", Status: table.ActiveStatus},
			{ID: 8, Name: "Tisch 8", Status: table.ActiveStatus},
			{ID: 9, Name: "Tisch 9", Status: table.ActiveStatus},
			{ID: 10, Name: "Tisch 10", Status: table.ActiveStatus},
			{ID: 11, Name: "Tisch 11", Status: table.ActiveStatus},
			{ID: 12, Name: "Tisch 12", Status: table.ActiveStatus},
			{ID: 13, Name: "Tisch 13", Status: table.ActiveStatus},
			{ID: 14, Name: "Tisch 14", Status: table.ActiveStatus},
			{ID: 15, Name: "Tisch 15", Status: table.ActiveStatus},
			{ID: 16, Name: "Zelt A1", Status: table.ActiveStatus},
			{ID: 17, Name: "Zelt A2", Status: table.ActiveStatus},
			{ID: 18, Name: "Stehtisch Bar", Status: table.ActiveStatus},
			{ID: 19, Name: "Stehtisch Eingang", Status: table.ActiveStatus},
			{ID: 20, Name: "Stehtisch Terrasse", Status: table.ActiveStatus},
		},
		Produkte: []produkt{
			{ID: 1, Name: "Bratwurst", Kategorie: product.EssenKategorie, Steuersatz: steuer.ErmaessigtSteuersatz, Status: product.ActiveStatus, Varianten: []variante{
				{ID: 1, Name: "Normal", PreisCents: 350, Status: product.ActiveStatus},
				{ID: 2, Name: "XXL", PreisCents: 500, Status: product.ActiveStatus},
				{ID: 3, Name: "Currywurst", PreisCents: 450, Status: product.ActiveStatus},
			}},
			{ID: 2, Name: "Pommes", Kategorie: product.EssenKategorie, Steuersatz: steuer.ErmaessigtSteuersatz, Status: product.ActiveStatus, Varianten: []variante{
				{ID: 4, Name: "Klein", PreisCents: 250, Status: product.ActiveStatus},
				{ID: 5, Name: "Groß", PreisCents: 350, Status: product.ActiveStatus},
			}},
			{ID: 3, Name: "Flammkuchen", Kategorie: product.EssenKategorie, Steuersatz: steuer.ErmaessigtSteuersatz, Status: product.ActiveStatus, Varianten: []variante{
				{ID: 6, Name: "Classic", PreisCents: 600, Status: product.ActiveStatus},
				{ID: 7, Name: "Speck & Zwiebel", PreisCents: 700, Status: product.ActiveStatus},
				{ID: 8, Name: "Mediterran", PreisCents: 750, Status: product.ActiveStatus},
			}},
			{ID: 4, Name: "Tagesgericht", Kategorie: product.EssenKategorie, Steuersatz: steuer.ErmaessigtSteuersatz, Status: product.ActiveStatus, Varianten: []variante{
				{ID: 9, Name: "Fr: Schnitzel mit Pommes", PreisCents: 1250, Status: product.ActiveStatus},
				{ID: 10, Name: "Sa: Gulasch mit Spätzle", PreisCents: 1150, Status: product.ActiveStatus},
				{ID: 11, Name: "So: Hähnchen mit Reis", PreisCents: 1050, Status: product.ActiveStatus},
			}},
			{ID: 5, Name: "Grillplatte", Kategorie: product.EssenKategorie, Steuersatz: steuer.ErmaessigtSteuersatz, Status: product.ActiveStatus, Varianten: []variante{
				{ID: 12, Name: "Klein", PreisCents: 800, Status: product.ActiveStatus},
				{ID: 13, Name: "Groß", PreisCents: 1400, Status: product.ActiveStatus},
			}},
			{ID: 6, Name: "Salat", Kategorie: product.EssenKategorie, Steuersatz: steuer.ErmaessigtSteuersatz, Status: product.ActiveStatus, Varianten: []variante{
				{ID: 14, Name: "Gemischter Salat", PreisCents: 550, Status: product.ActiveStatus},
				{ID: 15, Name: "Caesar Salat", PreisCents: 650, Status: product.ActiveStatus},
			}},
			{ID: 7, Name: "Kuchen", Kategorie: product.EssenKategorie, Steuersatz: steuer.ErmaessigtSteuersatz, Status: product.ActiveStatus, Varianten: []variante{
				{ID: 16, Name: "Stück", PreisCents: 250, Status: product.ActiveStatus},
			}},
			{ID: 8, Name: "Waffeln", Kategorie: product.EssenKategorie, Steuersatz: steuer.ErmaessigtSteuersatz, Status: product.ActiveStatus, Varianten: []variante{
				{ID: 17, Name: "mit Puderzucker", PreisCents: 300, Status: product.ActiveStatus},
				{ID: 18, Name: "mit Sahne", PreisCents: 350, Status: product.ActiveStatus},
				{ID: 19, Name: "mit Nutella", PreisCents: 400, Status: product.ActiveStatus},
			}},
			{ID: 9, Name: "Brezel", Kategorie: product.EssenKategorie, Steuersatz: steuer.ErmaessigtSteuersatz, Status: product.ActiveStatus, Varianten: []variante{
				{ID: 20, Name: "Normal", PreisCents: 200, Status: product.ActiveStatus},
				{ID: 21, Name: "mit Butter", PreisCents: 300, Status: product.ActiveStatus},
			}},
			{ID: 11, Name: "Bier", Kategorie: product.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Status: product.ActiveStatus, Varianten: []variante{
				{ID: 23, Name: "0,3l", PreisCents: 300, Status: product.ActiveStatus},
				{ID: 24, Name: "0,5l", PreisCents: 450, Status: product.ActiveStatus},
				{ID: 25, Name: "Maß 1,0l", PreisCents: 850, Status: product.ActiveStatus},
			}},
			{ID: 12, Name: "Weizen", Kategorie: product.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Status: product.ActiveStatus, Varianten: []variante{
				{ID: 26, Name: "Klein 0,3l", PreisCents: 300, Status: product.ActiveStatus},
				{ID: 27, Name: "Groß 0,5l", PreisCents: 400, Status: product.ActiveStatus},
				{ID: 28, Name: "Colaweizen Klein", PreisCents: 300, Status: product.ActiveStatus},
				{ID: 29, Name: "Colaweizen Groß", PreisCents: 400, Status: product.ActiveStatus},
				{ID: 30, Name: "Russ", PreisCents: 300, Status: product.ActiveStatus},
			}},
			{ID: 13, Name: "Softdrinks", Kategorie: product.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Status: product.ActiveStatus, Varianten: []variante{
				{ID: 31, Name: "Cola", PreisCents: 280, Status: product.ActiveStatus},
				{ID: 32, Name: "Fanta", PreisCents: 280, Status: product.ActiveStatus},
				{ID: 33, Name: "Spezi", PreisCents: 280, Status: product.ActiveStatus},
				{ID: 34, Name: "Sprite", PreisCents: 280, Status: product.ActiveStatus},
				{ID: 35, Name: "Mezzo Mix", PreisCents: 280, Status: product.ActiveStatus},
			}},
			{ID: 14, Name: "Wasser", Kategorie: product.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Status: product.ActiveStatus, Varianten: []variante{
				{ID: 36, Name: "Still 0,5l", PreisCents: 200, Status: product.ActiveStatus},
				{ID: 37, Name: "Medium 0,5l", PreisCents: 200, Status: product.ActiveStatus},
				{ID: 38, Name: "Sprudel 0,5l", PreisCents: 200, Status: product.ActiveStatus},
			}},
			{ID: 15, Name: "Saftschorle", Kategorie: product.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Status: product.ActiveStatus, Varianten: []variante{
				{ID: 39, Name: "Apfelschorle 0,5l", PreisCents: 300, Status: product.ActiveStatus},
				{ID: 40, Name: "Johannisbeerschorle 0,5l", PreisCents: 300, Status: product.ActiveStatus},
				{ID: 41, Name: "Rhabarberschorle 0,5l", PreisCents: 350, Status: product.ActiveStatus},
			}},
			{ID: 16, Name: "Wein", Kategorie: product.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Status: product.ActiveStatus, Varianten: []variante{
				{ID: 42, Name: "Weißwein 0,2l", PreisCents: 400, Status: product.ActiveStatus},
				{ID: 43, Name: "Rotwein 0,2l", PreisCents: 400, Status: product.ActiveStatus},
				{ID: 44, Name: "Rosé 0,2l", PreisCents: 400, Status: product.ActiveStatus},
			}},
			{ID: 17, Name: "Kaffee", Kategorie: product.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Status: product.ActiveStatus, Varianten: []variante{
				{ID: 45, Name: "Tasse", PreisCents: 200, Status: product.ActiveStatus},
				{ID: 46, Name: "Espresso", PreisCents: 180, Status: product.ActiveStatus},
			}},
			{ID: 18, Name: "Tee", Kategorie: product.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Status: product.ActiveStatus, Varianten: []variante{
				{ID: 47, Name: "Verschiedene Sorten", PreisCents: 200, Status: product.ActiveStatus},
			}},
			{ID: 19, Name: "Hugo/Aperol", Kategorie: product.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Status: product.ActiveStatus, Varianten: []variante{
				{ID: 48, Name: "Hugo", PreisCents: 550, Status: product.ActiveStatus},
				{ID: 49, Name: "Aperol Spritz", PreisCents: 550, Status: product.ActiveStatus},
			}},
			{ID: 21, Name: "Festbändchen", Kategorie: product.SonstigesKategorie, Steuersatz: steuer.BefreitSteuersatz, Status: product.ActiveStatus, Varianten: []variante{
				{ID: 51, Name: "Erwachsene", PreisCents: 500, Status: product.ActiveStatus},
				{ID: 52, Name: "Kinder", PreisCents: 300, Status: product.ActiveStatus},
			}},
		},
		Betreiber: settings.Betreiber{
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
