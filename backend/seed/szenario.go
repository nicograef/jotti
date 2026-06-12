package seed

import (
	"fmt"
	"sort"
	"time"

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

type favorit struct {
	UserID  int
	TischID int
}

// --- Drehbuch-Typen ---

// bestellposten verweist auf eine Produktvariante und die gewünschte Menge.
// Die Engine löst daraus die vollständige Position aus den Stammdaten auf.
type bestellposten struct {
	VarianteID int
	Menge      int
}

// aktion ist ein Schritt im Drehbuch einer Kassensitzung. Die Engine übersetzt jede Aktion
// über die Domain-Konstruktoren in ein oder zwei Events (umbuchen, kassensturz mit Differenz).
type aktion interface{ istAktion() }

// bestellen nimmt eine Bestellung am Tisch auf.
type bestellen struct {
	Tisch, User int
	Posten      []bestellposten
	Kommentar   string
}

// ausgeben bestätigt die Ausgabe; leere Posten = alle ausstehenden Positionen.
type ausgeben struct {
	Tisch, User int
	Posten      []bestellposten
}

// kassieren kassiert Positionen; leere Posten = alle unbezahlten Positionen.
type kassieren struct {
	Tisch, User int
	Posten      []bestellposten
}

// stornieren storniert die angegebenen Posten (Pflicht); die Engine wählt dafür die
// jüngsten nicht stornierten Positionen der passenden Variante.
type stornieren struct {
	Tisch, User int
	Posten      []bestellposten
	Kommentar   string
}

// auszahlen leistet eine Auszahlung; BetragCents 0 = gesamtes Guthaben (negativer Saldo).
type auszahlen struct {
	Tisch, User int
	BetragCents int
	Kommentar   string
}

// umbuchen verschiebt unbezahlte Positionen als atomares Storno-/Bestellungs-Paar mit den
// Standard-Kommentaren („Umbuchung auf/von Tisch …"); leere Posten = alle unbezahlten.
type umbuchen struct {
	VonTisch, NachTisch, User int
	Posten                    []bestellposten
}

// direktverkauf tätigt einen Direktverkauf; die VerkaufID ist die feste Subject-UUID
// aus dem Drehbuch.
type direktverkauf struct {
	VerkaufID string
	User      int
	Posten    []bestellposten
	Kommentar string
}

// direktverkaufStorno storniert Positionen eines Direktverkaufs; leere Posten = alle
// nicht stornierten Positionen des Verkaufs.
type direktverkaufStorno struct {
	VerkaufID string
	User      int
	Posten    []bestellposten
	Kommentar string
}

// geldtransit bucht eine Einlage oder Entnahme auf Kassensitzungs-Ebene.
type geldtransit struct {
	User        int
	Richtung    string // "einlage" | "entnahme"
	BetragCents int
	Kommentar   string
}

// kassensturz zählt die Kasse: Die Engine berechnet den Soll-Bestand aus den bisherigen
// Events, der Ist-Bestand ergibt sich als Soll − DifferenzCents. Bei Differenz ≠ 0 folgt
// die Differenz-Buchung (Zwei-Event-Muster wie im Produktivbetrieb).
type kassensturz struct {
	User           int
	DifferenzCents int
}

func (bestellen) istAktion()           {}
func (ausgeben) istAktion()            {}
func (kassieren) istAktion()           {}
func (stornieren) istAktion()          {}
func (auszahlen) istAktion()           {}
func (umbuchen) istAktion()            {}
func (direktverkauf) istAktion()       {}
func (direktverkaufStorno) istAktion() {}
func (geldtransit) istAktion()         {}
func (kassensturz) istAktion()         {}

// profilPunkt ist eine Stützstelle der Tagesprofil-Kurve: Nach EventAnteil der Events ist
// ZeitAnteil des Sitzungsfensters vergangen. Damit entstehen Stoßzeiten (viele Events in
// wenig Zeit). Beide Anteile müssen streng monoton steigen; (0,0) und (1,1) sind implizit.
type profilPunkt struct {
	EventAnteil float64
	ZeitAnteil  float64
}

// tseAusfall ist ein TSE-Ausfallfenster relativ zum Sitzungsstart: Fiskalische Events in
// diesem Fenster bleiben unsigniert und erhalten je einen Nachsignier-Auftrag. In
// abgeschlossenen Sitzungen gelten die Aufträge als vom Worker abgearbeitet, in der offenen
// Sitzung bleiben sie offen (der Worker ist ohne TSE-Konfiguration inaktiv).
type tseAusfall struct {
	NachStart time.Duration
	Dauer     time.Duration
	Grund     string // Fehlertext der Signierversuche während der Störung (leer, wenn nie versucht)
}

// kassensitzungDrehbuch beschreibt einen Betriebstag: Zeitfenster relativ zu „jetzt",
// Eröffnung und die chronologische Aktionsfolge. Für abgeschlossene Sitzungen hängt die
// Engine den Tagesabschluss (mit berechneten Summen) automatisch ans Fensterende.
type kassensitzungDrehbuch struct {
	ZNr                 int
	Bezeichnung         string
	EroeffnetVon        int
	AnfangsbestandCents int
	StartVorJetzt       time.Duration
	Dauer               time.Duration
	Abgeschlossen       bool
	Tagesprofil         []profilPunkt
	TSEAusfaelle        []tseAusfall
	Aktionen            []aktion
}

// szenario bündelt die Stammdaten und das Drehbuch.
type szenario struct {
	Benutzer  []benutzer
	Tische    []tisch
	Produkte  []produkt
	Favoriten []favorit
	Betreiber settings.Betreiber
	Sitzungen []kassensitzungDrehbuch
}

func strPtr(s string) *string { return &s }

// --- Drehbuch-Helfer ---

// pos ist die Kurzschreibweise für einen Bestellposten im Drehbuch.
func pos(varianteID, menge int) bestellposten {
	return bestellposten{VarianteID: varianteID, Menge: menge}
}

// posten bündelt Bestellposten zu einer Bestellung.
func posten(p ...bestellposten) []bestellposten { return p }

// runde ist der Standard-Zyklus eines Tischbesuchs: Bestellen, Ausgeben, Kassieren.
func runde(tisch, user int, p []bestellposten) []aktion {
	return []aktion{
		bestellen{Tisch: tisch, User: user, Posten: p},
		ausgeben{Tisch: tisch, User: user},
		kassieren{Tisch: tisch, User: user},
	}
}

// runden erzeugt n Standard-Zyklen für einen Tisch und rotiert dabei durch die Menüfolge.
func runden(tisch, user, n int, menues ...[]bestellposten) []aktion {
	var aktionen []aktion
	for i := range n {
		aktionen = append(aktionen, runde(tisch, user, menues[i%len(menues)])...)
	}
	return aktionen
}

// verflechte verteilt mehrere Aktionsstränge proportional über den Tag: Jeder Strang behält
// seine interne Reihenfolge, seine Aktionen werden aber gleichmäßig über die Gesamtfolge
// gestreut — so laufen die Tische chronologisch verzahnt statt nacheinander.
func verflechte(straenge ...[]aktion) []aktion {
	type eintrag struct {
		anteil float64
		strang int
		index  int
	}
	var eintraege []eintrag
	for s, strang := range straenge {
		for i := range strang {
			eintraege = append(eintraege, eintrag{
				anteil: (float64(i) + 0.5) / float64(len(strang)),
				strang: s,
				index:  i,
			})
		}
	}
	sort.SliceStable(eintraege, func(a, b int) bool { return eintraege[a].anteil < eintraege[b].anteil })

	ergebnis := make([]aktion, 0, len(eintraege))
	for _, e := range eintraege {
		ergebnis = append(ergebnis, straenge[e.strang][e.index])
	}
	return ergebnis
}

// dvID liefert die feste Direktverkauf-Subject-UUID nach erkennbarem Schema:
// erste Gruppe = Festtag (1–3), letzte Gruppe = laufende Nummer.
func dvID(tag, nr int) string {
	return fmt.Sprintf("%08d-0000-4000-8000-%012d", tag, nr)
}

// --- Das Drehbuch: 3-Tage-Sommerfest „TSV Musterstadt e.V." ---

// Benutzer-IDs (sprechende Konstanten für das Drehbuch).
const (
	thomas = 2 // Admin
	felix  = 3 // Serviceleitung
	maria  = 4 // Service
	lisa   = 5 // Service
	jan    = 6 // Service
	sophie = 7 // Serviceleitung
	markus = 8 // Service
	anna   = 9 // Service
)

// demoSzenario liefert die Stammdaten und das vollständige 3-Tage-Drehbuch:
// Freitag (ruhiger Eröffnungsabend) und Samstag (Haupttag) abgeschlossen,
// Sonntag (aktueller Tag) offen.
func demoSzenario() szenario {
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
			{ID: 10, Name: "Paul Fischer", Username: "paul", Rolle: user.ServiceRole, Status: user.InactiveStatus},
			{ID: 11, Name: "Sabine Wolf", Username: "sabine", Rolle: user.ServiceRole, Status: user.DeletedStatus},
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
			{ID: 21, Name: "Reserviert", Status: table.InactiveStatus},
			{ID: 22, Name: "Alter Tisch", Status: table.DeletedStatus},
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
			{ID: 10, Name: "Suppe", Kategorie: product.EssenKategorie, Steuersatz: steuer.ErmaessigtSteuersatz, Status: product.InactiveStatus, Varianten: []variante{
				{ID: 22, Name: "Tagessuppe", PreisCents: 400, Status: product.ActiveStatus},
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
			{ID: 20, Name: "Glühwein", Kategorie: product.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Status: product.InactiveStatus, Varianten: []variante{
				{ID: 50, Name: "Tasse", PreisCents: 350, Status: product.ActiveStatus},
			}},
			{ID: 21, Name: "Festbändchen", Kategorie: product.SonstigesKategorie, Steuersatz: steuer.BefreitSteuersatz, Status: product.ActiveStatus, Varianten: []variante{
				{ID: 51, Name: "Erwachsene", PreisCents: 500, Status: product.ActiveStatus},
				{ID: 52, Name: "Kinder", PreisCents: 300, Status: product.ActiveStatus},
			}},
			{ID: 22, Name: "Langos", Kategorie: product.EssenKategorie, Steuersatz: steuer.ErmaessigtSteuersatz, Status: product.DeletedStatus, Varianten: []variante{
				{ID: 53, Name: "mit Knoblauch", PreisCents: 400, Status: product.DeletedStatus},
				{ID: 54, Name: "mit Käse", PreisCents: 500, Status: product.DeletedStatus},
			}},
		},
		Favoriten: []favorit{
			{UserID: maria, TischID: 1}, {UserID: maria, TischID: 2}, {UserID: maria, TischID: 3},
			{UserID: maria, TischID: 4}, {UserID: maria, TischID: 5},
			{UserID: lisa, TischID: 7}, {UserID: lisa, TischID: 8}, {UserID: lisa, TischID: 9},
			{UserID: lisa, TischID: 10},
			{UserID: jan, TischID: 2}, {UserID: jan, TischID: 16}, {UserID: jan, TischID: 17},
			{UserID: markus, TischID: 11}, {UserID: markus, TischID: 12}, {UserID: markus, TischID: 13},
			{UserID: anna, TischID: 18}, {UserID: anna, TischID: 19}, {UserID: anna, TischID: 20},
			{UserID: felix, TischID: 5},
		},
		Betreiber: settings.Betreiber{
			Vereinsname:  "TSV Musterstadt e.V.",
			Strasse:      "Sportplatzweg 7",
			Plz:          "63776",
			Ort:          "Musterstadt",
			Steuernummer: strPtr("204/123/45678"),
			UstID:        nil,
		},
		Sitzungen: []kassensitzungDrehbuch{
			{
				ZNr:                 1,
				Bezeichnung:         "Sommerfest 26 Freitag",
				EroeffnetVon:        thomas,
				AnfangsbestandCents: 15000,
				StartVorJetzt:       53 * time.Hour,
				Dauer:               5*time.Hour + 30*time.Minute,
				Abgeschlossen:       true,
				// Eröffnungsabend: zum Abend hin voller.
				Tagesprofil: []profilPunkt{{EventAnteil: 0.3, ZeitAnteil: 0.45}, {EventAnteil: 0.8, ZeitAnteil: 0.8}},
				Aktionen:    freitagsAktionen(),
			},
			{
				ZNr:                 2,
				Bezeichnung:         "Sommerfest 26 Samstag",
				EroeffnetVon:        thomas,
				AnfangsbestandCents: 20000,
				StartVorJetzt:       30 * time.Hour,
				Dauer:               12 * time.Hour,
				Abgeschlossen:       true,
				// Haupttag: Mittagsstoßzeit, ruhiger Nachmittag, Abendstoßzeit.
				Tagesprofil: []profilPunkt{
					{EventAnteil: 0.25, ZeitAnteil: 0.18},
					{EventAnteil: 0.35, ZeitAnteil: 0.30},
					{EventAnteil: 0.55, ZeitAnteil: 0.60},
					{EventAnteil: 0.85, ZeitAnteil: 0.80},
				},
				// Cloud-TSE-Störung in der Abendstoßzeit; der Nachsignier-Worker arbeitet
				// die Aufträge nach dem Fensterende ab.
				TSEAusfaelle: []tseAusfall{{
					NachStart: 8*time.Hour + 30*time.Minute,
					Dauer:     time.Hour,
					Grund:     "Cloud-TSE nicht erreichbar: Zeitüberschreitung nach 10 Sekunden",
				}},
				Aktionen: samstagsAktionen(),
			},
			{
				ZNr:                 3,
				Bezeichnung:         "Sommerfest 26 Sonntag",
				EroeffnetVon:        thomas,
				AnfangsbestandCents: 20000,
				StartVorJetzt:       5 * time.Hour,
				Dauer:               5 * time.Hour,
				Abgeschlossen:       false,
				// Kurzer TSE-Aussetzer am laufenden Tag: Die Aufträge bleiben offen und
				// ohne Fehlertext, weil der Worker ohne TSE-Konfiguration nie einen
				// Signierversuch startet.
				TSEAusfaelle: []tseAusfall{{NachStart: 4 * time.Hour, Dauer: 10 * time.Minute}},
				Aktionen:     sonntagsAktionen(),
			},
		},
	}
}

// freitagsAktionen ist der ruhige Eröffnungsabend (~160 Events): neun Tische mit
// Standard-Zyklen plus Direktverkaufsstand, abends der Kassensturz ohne Differenz.
func freitagsAktionen() []aktion {
	stammtisch := runden(1, maria, 8,
		posten(pos(24, 4)),                        // Bier-Runde
		posten(pos(9, 2), pos(27, 2)),             // Tagesgericht Schnitzel + Weizen
		posten(pos(1, 2), pos(20, 2), pos(24, 4)), // Brotzeit
		posten(pos(24, 4), pos(33, 1)),
	)
	ehepaar := runden(3, maria, 3,
		posten(pos(45, 2), pos(16, 2)), // Kaffee und Kuchen
		posten(pos(9, 2), pos(42, 2)),
		posten(pos(37, 2)),
	)
	vorstand := runden(5, felix, 6,
		posten(pos(49, 4)), // Aperol-Runde
		posten(pos(13, 2), pos(43, 2)),
		posten(pos(8, 2), pos(44, 2)),
		posten(pos(46, 4)),
	)
	familie := runden(8, lisa, 6,
		posten(pos(9, 2), pos(4, 2), pos(31, 2), pos(36, 2)),
		posten(pos(17, 2), pos(19, 1)),
		posten(pos(1, 2), pos(5, 1), pos(24, 2)),
	)
	kegelclub := runden(9, lisa, 6,
		posten(pos(24, 6)),
		posten(pos(9, 3), pos(24, 3)),
		posten(pos(25, 2), pos(20, 4)),
	)
	jugend := runden(12, markus, 5,
		posten(pos(4, 4), pos(31, 4)),
		posten(pos(17, 4)),
		posten(pos(20, 4), pos(34, 2)),
	)
	zeltA1 := runden(16, jan, 6,
		posten(pos(1, 4), pos(24, 4)),
		posten(pos(6, 2), pos(7, 1), pos(27, 2)),
		posten(pos(24, 6)),
	)
	bar := runden(18, jan, 7,
		posten(pos(24, 2)),
		posten(pos(48, 2)),
		posten(pos(23, 3)),
		posten(pos(49, 2)),
	)
	eingang := runden(19, anna, 5,
		posten(pos(31, 2), pos(20, 2)),
		posten(pos(24, 2)),
		posten(pos(39, 2)),
	)
	stand := []aktion{
		direktverkauf{VerkaufID: dvID(1, 1), User: sophie, Posten: posten(pos(51, 2), pos(52, 2))},
		direktverkauf{VerkaufID: dvID(1, 2), User: sophie, Posten: posten(pos(51, 4))},
		direktverkauf{VerkaufID: dvID(1, 3), User: sophie, Posten: posten(pos(16, 3))},
		direktverkauf{VerkaufID: dvID(1, 4), User: sophie, Posten: posten(pos(51, 2), pos(16, 2))},
	}

	tag := verflechte(stammtisch, ehepaar, vorstand, familie, kegelclub, jugend, zeltA1, bar, eingang, stand)
	return append(tag, kassensturz{User: thomas})
}

// samstagsAktionen ist der Haupttag (~700 Events): voller Betrieb auf 16 Tischen mit
// Geburtstagsfeier, Stornierungen durch die Serviceleitung, Teil-Ausgabe und -Zahlung,
// Auszahlungen, Direktverkaufsstand mit Storno, Geldtransit-Entnahme und Kassensturz
// mit kleiner Soll/Ist-Differenz.
func samstagsAktionen() []aktion {
	stammtisch := runden(1, maria, 18,
		posten(pos(24, 4)),
		posten(pos(10, 2), pos(27, 2)),
		posten(pos(1, 2), pos(20, 2), pos(24, 4)),
		posten(pos(24, 4), pos(33, 1)),
	)
	// Junge Leute: viele Getränke, eine Fehlbestellung wird von der Serviceleitung storniert.
	jungeLeute := append(runden(2, jan, 19,
		posten(pos(24, 5)),
		posten(pos(28, 2), pos(29, 2)),
		posten(pos(48, 2), pos(49, 2)),
		posten(pos(35, 3), pos(17, 3)),
	),
		bestellen{Tisch: 2, User: jan, Posten: posten(pos(25, 3))},
		stornieren{Tisch: 2, User: felix, Posten: posten(pos(25, 3)), Kommentar: "Falsch bestellt, Gäste wollten 0,5l statt Maß"},
		bestellen{Tisch: 2, User: jan, Posten: posten(pos(24, 3))},
		ausgeben{Tisch: 2, User: jan},
		kassieren{Tisch: 2, User: jan},
	)
	ehepaar := runden(3, maria, 4,
		posten(pos(45, 2), pos(16, 2)),
		posten(pos(10, 2), pos(42, 2)),
	)
	// Geburtstagsfeier: große Bestellungen; am Ende Reklamation einer Grillplatte
	// nach der Zahlung mit anschließender Auszahlung des Guthabens.
	geburtstag := append(runden(4, sophie, 12,
		posten(pos(13, 3), pos(31, 4), pos(24, 4)),
		posten(pos(10, 4), pos(27, 4)),
		posten(pos(16, 6), pos(45, 6)),
		posten(pos(12, 4), pos(39, 4)),
	),
		bestellen{Tisch: 4, User: sophie, Posten: posten(pos(13, 2)), Kommentar: "Nachbestellung Geburtstagsrunde"},
		ausgeben{Tisch: 4, User: sophie},
		kassieren{Tisch: 4, User: sophie},
		stornieren{Tisch: 4, User: sophie, Posten: posten(pos(13, 1)), Kommentar: "Reklamation: Grillplatte kalt serviert"},
		auszahlen{Tisch: 4, User: sophie, Kommentar: "Erstattung nach Reklamation"},
	)
	vorstand := runden(5, felix, 12,
		posten(pos(49, 4)),
		posten(pos(13, 2), pos(43, 2)),
		posten(pos(10, 3), pos(42, 3)),
		posten(pos(48, 2), pos(44, 2)),
	)
	// Kollegenrunde: eine Bestellung wird in zwei Hälften bezahlt (getrennte Kasse).
	kollegen := append(runden(7, lisa, 8,
		posten(pos(10, 2), pos(27, 2)),
		posten(pos(24, 4)),
		posten(pos(6, 2), pos(31, 2)),
	),
		bestellen{Tisch: 7, User: lisa, Posten: posten(pos(10, 4), pos(27, 4))},
		ausgeben{Tisch: 7, User: lisa},
		kassieren{Tisch: 7, User: lisa, Posten: posten(pos(10, 2), pos(27, 2))},
		kassieren{Tisch: 7, User: lisa},
	)
	familie := runden(8, lisa, 20,
		posten(pos(10, 2), pos(4, 2), pos(31, 2), pos(36, 2)),
		posten(pos(17, 2), pos(19, 2)),
		posten(pos(1, 4), pos(24, 2), pos(31, 2)),
		posten(pos(16, 4), pos(45, 2)),
	)
	kegelclub := runden(9, lisa, 12,
		posten(pos(24, 6)),
		posten(pos(10, 3), pos(24, 3)),
		posten(pos(25, 2), pos(21, 4)),
	)
	// Tisch 10: Stornierung nach Bezahlung mit Auszahlung des Guthabens.
	reklamation := append(runden(10, sophie, 2,
		posten(pos(24, 2), pos(31, 2)),
	),
		bestellen{Tisch: 10, User: sophie, Posten: posten(pos(8, 2), pos(44, 2))},
		ausgeben{Tisch: 10, User: sophie},
		kassieren{Tisch: 10, User: sophie},
		stornieren{Tisch: 10, User: sophie, Posten: posten(pos(8, 1)), Kommentar: "Flammkuchen verbrannt, Küche bestätigt"},
		auszahlen{Tisch: 10, User: sophie, Kommentar: "Erstattung Flammkuchen"},
	)
	// Jugendgruppe: große Essensbestellung kommt in zwei Ausgaben aus der Küche.
	jugend := append(runden(12, markus, 10,
		posten(pos(4, 4), pos(31, 4)),
		posten(pos(17, 4), pos(34, 2)),
		posten(pos(20, 4)),
	),
		bestellen{Tisch: 12, User: markus, Posten: posten(pos(12, 6), pos(31, 6))},
		ausgeben{Tisch: 12, User: markus, Posten: posten(pos(12, 3), pos(31, 3))},
		ausgeben{Tisch: 12, User: markus},
		kassieren{Tisch: 12, User: markus},
	)
	musikverein := runden(13, markus, 6,
		posten(pos(27, 6)),
		posten(pos(10, 4), pos(24, 4)),
		posten(pos(1, 4), pos(20, 4)),
	)
	zeltA1 := runden(16, jan, 26,
		posten(pos(1, 4), pos(24, 4)),
		posten(pos(6, 2), pos(7, 2), pos(27, 4)),
		posten(pos(10, 4), pos(24, 6)),
		posten(pos(25, 4)),
	)
	zeltA2 := runden(17, jan, 10,
		posten(pos(2, 2), pos(24, 2)),
		posten(pos(10, 2), pos(39, 2)),
		posten(pos(14, 2), pos(36, 2)),
	)
	bar := runden(18, anna, 28,
		posten(pos(24, 2)),
		posten(pos(48, 2)),
		posten(pos(49, 2)),
		posten(pos(23, 3)),
		posten(pos(30, 2)),
	)
	eingang := runden(19, anna, 20,
		posten(pos(31, 2), pos(20, 2)),
		posten(pos(24, 2)),
		posten(pos(39, 2), pos(16, 1)),
	)
	terrasse := runden(20, anna, 14,
		posten(pos(41, 2), pos(15, 1)),
		posten(pos(24, 3)),
		posten(pos(10, 2), pos(38, 2)),
	)
	// Direktverkaufsstand: Festbändchen und Kuchen den ganzen Tag, ein Verkauf wird
	// von der Serviceleitung storniert.
	stand := []aktion{
		direktverkauf{VerkaufID: dvID(2, 1), User: sophie, Posten: posten(pos(51, 4), pos(52, 2))},
		direktverkauf{VerkaufID: dvID(2, 2), User: sophie, Posten: posten(pos(51, 2))},
		direktverkauf{VerkaufID: dvID(2, 3), User: sophie, Posten: posten(pos(16, 4))},
		direktverkauf{VerkaufID: dvID(2, 4), User: sophie, Posten: posten(pos(51, 6), pos(52, 4))},
		direktverkauf{VerkaufID: dvID(2, 5), User: sophie, Posten: posten(pos(16, 2), pos(52, 2))},
		direktverkauf{VerkaufID: dvID(2, 6), User: sophie, Posten: posten(pos(51, 2))},
		direktverkaufStorno{VerkaufID: dvID(2, 6), User: sophie, Kommentar: "Doppelt kassiert, Bändchen zurückgegeben"},
		direktverkauf{VerkaufID: dvID(2, 7), User: sophie, Posten: posten(pos(51, 3), pos(16, 3))},
		direktverkauf{VerkaufID: dvID(2, 8), User: sophie, Posten: posten(pos(52, 4))},
		direktverkauf{VerkaufID: dvID(2, 9), User: sophie, Posten: posten(pos(51, 2), pos(16, 2))},
		direktverkauf{VerkaufID: dvID(2, 10), User: sophie, Posten: posten(pos(16, 6))},
	}

	tag := verflechte(stammtisch, jungeLeute, ehepaar, geburtstag, vorstand, kollegen, familie,
		kegelclub, reklamation, jugend, musikverein, zeltA1, zeltA2, bar, eingang, terrasse, stand)
	return append(tag,
		geldtransit{User: thomas, Richtung: "entnahme", BetragCents: 50000, Kommentar: "Bargeldabschöpfung, Einzahlung Vereinskonto"},
		kassensturz{User: thomas, DifferenzCents: 350},
	)
}

// sonntagsAktionen ist der offene aktuelle Tag (~130 Events): Tische in allen Zuständen
// (leer, frisch bestellt, teilgeliefert, teilbezahlt, Guthaben, abgeschlossen), eine
// Umbuchung vom Stehtisch Eingang an den freien Tisch 4 und die Wechselgeld-Einlage.
func sonntagsAktionen() []aktion {
	// Frühschoppen am Stammtisch: jede Runde direkt bezahlt, Tisch ist ausgeglichen.
	stammtisch := runden(1, maria, 4,
		posten(pos(26, 4)),
		posten(pos(11, 2), pos(27, 2)),
		posten(pos(20, 4), pos(24, 2)),
	)
	// Tisch 2: teilbezahlt — die erste Bestellung ist kassiert, die zweite noch offen.
	teilbezahlt := []aktion{
		bestellen{Tisch: 2, User: jan, Posten: posten(pos(11, 2), pos(24, 2))},
		ausgeben{Tisch: 2, User: jan},
		bestellen{Tisch: 2, User: jan, Posten: posten(pos(16, 2), pos(45, 2))},
		ausgeben{Tisch: 2, User: jan},
		kassieren{Tisch: 2, User: jan, Posten: posten(pos(11, 2), pos(24, 2))},
	}
	vorstand := runden(5, felix, 2,
		posten(pos(49, 2), pos(48, 2)),
		posten(pos(11, 2), pos(42, 2)),
	)
	// Tisch 6: frische Gäste, bestellt und ausgegeben, noch nicht kassiert.
	frischeGaeste := []aktion{
		bestellen{Tisch: 6, User: lisa, Posten: posten(pos(11, 3), pos(31, 2), pos(5, 2))},
		ausgeben{Tisch: 6, User: lisa},
	}
	// Tisch 7: Stornierung nach Bezahlung — der Tisch wartet mit Guthaben auf die Auszahlung.
	guthaben := []aktion{
		bestellen{Tisch: 7, User: lisa, Posten: posten(pos(11, 2), pos(27, 2))},
		ausgeben{Tisch: 7, User: lisa},
		kassieren{Tisch: 7, User: lisa},
		stornieren{Tisch: 7, User: sophie, Posten: posten(pos(11, 1)), Kommentar: "Essen kam zu spät, Kulanz"},
	}
	// Tisch 8: teilgeliefert — die Getränke sind ausgegeben, das Essen steht noch aus.
	teilgeliefert := []aktion{
		bestellen{Tisch: 8, User: lisa, Posten: posten(pos(13, 1), pos(5, 2), pos(31, 3))},
		ausgeben{Tisch: 8, User: lisa, Posten: posten(pos(31, 3))},
		bestellen{Tisch: 8, User: lisa, Posten: posten(pos(16, 2))},
	}
	// Kegelclub: Stornierung nach Bezahlung, Guthaben bereits ausgezahlt — abgeschlossen.
	kegelclub := []aktion{
		bestellen{Tisch: 9, User: lisa, Posten: posten(pos(24, 6), pos(11, 3))},
		ausgeben{Tisch: 9, User: lisa},
		kassieren{Tisch: 9, User: lisa},
		stornieren{Tisch: 9, User: felix, Posten: posten(pos(11, 1)), Kommentar: "Reklamation Tagesgericht, Kulanz"},
		auszahlen{Tisch: 9, User: lisa, Kommentar: "Erstattung Tagesgericht"},
	}
	dazugekommen := []aktion{
		bestellen{Tisch: 11, User: markus, Posten: posten(pos(24, 2), pos(31, 2))},
		ausgeben{Tisch: 11, User: markus},
		bestellen{Tisch: 11, User: markus, Posten: posten(pos(11, 2))},
	}
	frischBestellt := []aktion{
		bestellen{Tisch: 14, User: anna, Posten: posten(pos(11, 2), pos(14, 1), pos(39, 2))},
	}
	zeltA1 := append(runden(16, jan, 6,
		posten(pos(1, 4), pos(24, 4)),
		posten(pos(11, 4), pos(27, 4)),
		posten(pos(6, 2), pos(31, 2)),
	),
		bestellen{Tisch: 16, User: jan, Posten: posten(pos(24, 6))},
		ausgeben{Tisch: 16, User: jan},
	)
	zeltA2 := append(runden(17, jan, 2,
		posten(pos(2, 2), pos(24, 2)),
		posten(pos(11, 2), pos(39, 2)),
	),
		bestellen{Tisch: 17, User: jan, Posten: posten(pos(16, 4), pos(45, 4))},
	)
	bar := runden(18, anna, 8,
		posten(pos(24, 2)),
		posten(pos(30, 2)),
		posten(pos(48, 2)),
		posten(pos(23, 2)),
	)
	// Stehtisch Eingang: Die Gäste finden einen freien Tisch — die offene Bestellung
	// wird auf Tisch 4 umgebucht und dort ausgegeben.
	umbuchung := append(runden(19, anna, 5,
		posten(pos(31, 2), pos(20, 2)),
		posten(pos(24, 2)),
		posten(pos(39, 2)),
	),
		bestellen{Tisch: 19, User: anna, Posten: posten(pos(11, 2), pos(24, 2))},
		umbuchen{VonTisch: 19, NachTisch: 4, User: anna},
		ausgeben{Tisch: 4, User: anna},
	)
	terrasse := append(runden(20, anna, 4,
		posten(pos(41, 2), pos(16, 2)),
		posten(pos(24, 3)),
	),
		bestellen{Tisch: 20, User: anna, Posten: posten(pos(45, 2), pos(16, 2))},
	)
	stand := []aktion{
		direktverkauf{VerkaufID: dvID(3, 1), User: sophie, Posten: posten(pos(51, 2), pos(52, 1))},
		direktverkauf{VerkaufID: dvID(3, 2), User: sophie, Posten: posten(pos(16, 4))},
		direktverkauf{VerkaufID: dvID(3, 3), User: sophie, Posten: posten(pos(51, 2))},
		direktverkauf{VerkaufID: dvID(3, 4), User: sophie, Posten: posten(pos(16, 2), pos(52, 2))},
	}

	tag := verflechte(stammtisch, teilbezahlt, vorstand, frischeGaeste, guthaben, teilgeliefert,
		kegelclub, dazugekommen, frischBestellt, zeltA1, zeltA2, bar, umbuchung, terrasse, stand)
	// Wechselgeld-Einlage gleich nach der Eröffnung.
	return append([]aktion{
		geldtransit{User: thomas, Richtung: "einlage", BetragCents: 10000, Kommentar: "Wechselgeld-Einlage aus dem Tresor"},
	}, tag...)
}
