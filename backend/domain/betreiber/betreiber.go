package betreiber

import (
	"fmt"
	"time"

	z "github.com/Oudwins/zog"
)

type Betreiber struct {
	Vereinsname  string
	Strasse      string
	Plz          string
	Ort          string
	Steuernummer *string
	UstID        *string
	// ElsterGemeldetAm ist das Datum der ELSTER-Kassenmeldung (§ 146a Abs. 4 AO)
	// oder nil, solange die Kasse noch nicht gemeldet wurde. Es wird nicht über
	// den Konstruktor gesetzt, sondern über die dedizierten Meldungs-Befehle.
	ElsterGemeldetAm *time.Time
	UpdatedAt        time.Time
}

// Die Obergrenzen sind die amtlichen Maximallängen der DSFinV-K-Stammdaten
// (2.4, index.xml: NAME 60, STRASSE 60, PLZ 10, ORT 62, STNR 20, USTID 15) — der
// Export schreibt jedes Feld in eine dieser Spalten. Jedes Schema trimmt; danach
// fängt Min(1) einen Wert aus reinen Leerzeichen, Required das leere Feld (zog
// prüft Required vor den Transformationen). Die vier Pflichtfelder sind per
// Definition required — Aufrufstellen nutzen sie direkt und rufen `.Required()`
// nie erneut auf (zog mutiert den Empfänger in place).
var VereinsnameSchema = z.String().Trim().
	Min(1, z.Message("Vereinsname ist erforderlich")).
	Max(60, z.Message("Vereinsname zu lang")).
	Required(z.Message("Vereinsname ist erforderlich"))

var StrasseSchema = z.String().Trim().
	Min(1, z.Message("Straße ist erforderlich")).
	Max(60, z.Message("Straße zu lang")).
	Required(z.Message("Straße ist erforderlich"))

var PlzSchema = z.String().Trim().
	Min(1, z.Message("PLZ ist erforderlich")).
	Max(10, z.Message("PLZ zu lang")).
	Required(z.Message("PLZ ist erforderlich"))

var OrtSchema = z.String().Trim().
	Min(1, z.Message("Ort ist erforderlich")).
	Max(62, z.Message("Ort zu lang")).
	Required(z.Message("Ort ist erforderlich"))

// Steuernummer und USt-IdNr. sind optional: Ein Verein ohne Steuernummer lässt
// das Feld leer, deshalb tragen die beiden Schemas nur die Obergrenze.
var SteuernummerSchema = z.String().Trim().Max(20, z.Message("Steuernummer zu lang"))

var UstIDSchema = z.String().Trim().Max(15, z.Message("USt-IdNr. zu lang"))

// betreiberSchema prüft ausschließlich, ob die Stammdaten gefüllt sind: Validate
// läuft über Daten aus der Datenbank und ist das Gate vor dem Eröffnen einer
// Kassensitzung. Darum trägt es keine Obergrenzen — die Spalten sind TEXT, ein
// Bestandswert kann länger sein als die amtliche Maximallänge, und eine Grenze
// hier sperrte die Kasse. Die Grenzen gelten auf dem Schreibweg (NewBetreiber);
// der DSFinV-K-Export kürzt, was Bestandsdaten mitbringen.
var betreiberSchema = z.Struct(z.Shape{
	"Vereinsname":  z.String().Min(1, z.Message("Vereinsname ist erforderlich")).Required(),
	"Strasse":      z.String().Min(1, z.Message("Straße ist erforderlich")).Required(),
	"Plz":          z.String().Min(1, z.Message("PLZ ist erforderlich")).Required(),
	"Ort":          z.String().Min(1, z.Message("Ort ist erforderlich")).Required(),
	"Steuernummer": z.Ptr(z.String()),
	"UstID":        z.Ptr(z.String()),
	"UpdatedAt":    z.Time().Required(),
})

func (b Betreiber) Validate() error {
	if errs := betreiberSchema.Validate(&b); errs != nil {
		issues := z.Issues.FlattenAndCollect(errs)
		return fmt.Errorf("invalid betreiber: %v", issues)
	}
	return nil
}

func NewBetreiber(vereinsname, strasse, plz, ort string, steuernummer, ustId *string) (Betreiber, error) {
	// Jedes Feld wird einzeln gegen sein Schema geprüft: Das trimmt den Wert in
	// der lokalen Variablen, bevor er in die Struktur geht (zog schreibt die
	// Transformation über den Zeiger zurück). Die optionalen Felder werden über
	// eine Kopie geprüft, damit der Wert des Aufrufers unverändert bleibt.
	if issue := VereinsnameSchema.Validate(&vereinsname); issue != nil {
		return Betreiber{}, fmt.Errorf("invalid vereinsname")
	}

	if issue := StrasseSchema.Validate(&strasse); issue != nil {
		return Betreiber{}, fmt.Errorf("invalid strasse")
	}

	if issue := PlzSchema.Validate(&plz); issue != nil {
		return Betreiber{}, fmt.Errorf("invalid plz")
	}

	if issue := OrtSchema.Validate(&ort); issue != nil {
		return Betreiber{}, fmt.Errorf("invalid ort")
	}

	if steuernummer != nil {
		wert := *steuernummer
		if issue := SteuernummerSchema.Validate(&wert); issue != nil {
			return Betreiber{}, fmt.Errorf("invalid steuernummer")
		}
		steuernummer = &wert
	}

	if ustId != nil {
		wert := *ustId
		if issue := UstIDSchema.Validate(&wert); issue != nil {
			return Betreiber{}, fmt.Errorf("invalid ustId")
		}
		ustId = &wert
	}

	return Betreiber{
		Vereinsname:  vereinsname,
		Strasse:      strasse,
		Plz:          plz,
		Ort:          ort,
		Steuernummer: steuernummer,
		UstID:        ustId,
		UpdatedAt:    time.Now().UTC(),
	}, nil
}
