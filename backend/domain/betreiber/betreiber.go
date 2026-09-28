package betreiber

import (
	"fmt"
	"time"
	"unicode/utf8"

	z "github.com/Oudwins/zog"
)

type Betreiber struct {
	Vereinsname  string
	Strasse      string
	Plz          string
	Ort          string
	Steuernummer *string
	UstID        *string
	// ElsterGemeldetAm ist das Datum der ELSTER-Kassenmeldung (§ 146a Abs. 4 AO),
	// nil solange ungemeldet; gesetzt nur über die Meldungs-Befehle, nie im Konstruktor.
	ElsterGemeldetAm *time.Time
	UpdatedAt        time.Time
}

// Amtliche Maximallängen der Betreiber-Felder in den DSFinV-K-Stammdaten (2.4,
// index.xml: NAME 60, STRASSE 60, PLZ 10, ORT 62, STNR 20, USTID 15). Sie zählen
// Zeichen, nicht Bytes; api/fiskal/dsfinvk misst mit denselben Konstanten.
const (
	MaxLengthVereinsname  = 60
	MaxLengthStrasse      = 60
	MaxLengthPlz          = 10
	MaxLengthOrt          = 62
	MaxLengthSteuernummer = 20
	MaxLengthUstID        = 15
)

// Die Pflichtmeldungen der vier Adressfelder. Sie stehen einmal, weil das
// Feld-Schema (Schreibweg) und betreiberSchema (Lesepfad) dieselbe Lücke melden.
const (
	vereinsnameErforderlich = "Vereinsname ist erforderlich"
	strasseErforderlich     = "Straße ist erforderlich"
	plzErforderlich         = "PLZ ist erforderlich"
	ortErforderlich         = "Ort ist erforderlich"
)

// maxRunes baut die Prüfung „höchstens n Zeichen" für ein Zeichenketten-Schema.
// zogs Max zählt Bytes; ein Vereinsname mit Umlauten hat mehr Bytes als Zeichen
// und liefe damit gegen eine engere Grenze als die amtliche.
func maxRunes(n int) z.BoolTFunc[*string] {
	return func(wert *string, _ z.Ctx) bool {
		return utf8.RuneCountInString(*wert) <= n
	}
}

// Min(1) catches whitespace-only input because zog checks Required before Trim.
// Call sites never call .Required() again: zog mutates the receiver in place.
var VereinsnameSchema = z.String().Trim().
	Min(1, z.Message(vereinsnameErforderlich)).
	TestFunc(maxRunes(MaxLengthVereinsname), z.Message("Vereinsname zu lang")).
	Required(z.Message(vereinsnameErforderlich))

var StrasseSchema = z.String().Trim().
	Min(1, z.Message(strasseErforderlich)).
	TestFunc(maxRunes(MaxLengthStrasse), z.Message("Straße zu lang")).
	Required(z.Message(strasseErforderlich))

var PlzSchema = z.String().Trim().
	Min(1, z.Message(plzErforderlich)).
	TestFunc(maxRunes(MaxLengthPlz), z.Message("PLZ zu lang")).
	Required(z.Message(plzErforderlich))

var OrtSchema = z.String().Trim().
	Min(1, z.Message(ortErforderlich)).
	TestFunc(maxRunes(MaxLengthOrt), z.Message("Ort zu lang")).
	Required(z.Message(ortErforderlich))

// Optional: a Verein without Steuernummer leaves the field empty.
var SteuernummerSchema = z.String().Trim().
	TestFunc(maxRunes(MaxLengthSteuernummer), z.Message("Steuernummer zu lang"))

var UstIDSchema = z.String().Trim().
	TestFunc(maxRunes(MaxLengthUstID), z.Message("USt-IdNr. zu lang"))

// betreiberSchema gates opening a Kassensitzung on stored values, so it checks presence only.
// A max length here would lock the Kasse on a long stored value; NewBetreiber enforces it and the DSFinV-K export truncates.
var betreiberSchema = z.Struct(z.Shape{
	"Vereinsname":  z.String().Min(1, z.Message(vereinsnameErforderlich)).Required(),
	"Strasse":      z.String().Min(1, z.Message(strasseErforderlich)).Required(),
	"Plz":          z.String().Min(1, z.Message(plzErforderlich)).Required(),
	"Ort":          z.String().Min(1, z.Message(ortErforderlich)).Required(),
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
	// zog schreibt die Trim-Transformation über den Zeiger zurück; die optionalen
	// Felder werden darum über eine Kopie geprüft, damit der Wert des Aufrufers
	// unverändert bleibt.
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
