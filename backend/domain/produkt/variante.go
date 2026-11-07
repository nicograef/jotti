package produkt

import (
	"errors"
	"fmt"
	"slices"
	"time"

	z "github.com/Oudwins/zog"
)

// ErrPreisUngueltig means a Variante's price violates PreisCentsSchema.
var ErrPreisUngueltig = errors.New("invalid price")

// ErrPfandZielUngueltig means the linked Variante is missing, deleted or belongs to no Pfand-Produkt.
var ErrPfandZielUngueltig = errors.New("pfand target is no pfand variante")

// ErrPfandVerkettet means a Variante of a Pfand-Produkt would link a Pfand itself.
var ErrPfandVerkettet = errors.New("pfand produkt links a pfand")

// ErrPfandSteuersatzAbweichend means the Pfand and the linking Variante differ in Steuersatz.
var ErrPfandSteuersatzAbweichend = errors.New("pfand steuersatz differs")

// ErrPfandNochVerknuepft means a Produkt would lose its Pfand flag while Varianten link it as their Pfand.
var ErrPfandNochVerknuepft = errors.New("pfand produkt still linked")

type Status string

const (
	ActiveStatus   Status = "active"
	InactiveStatus Status = "inactive"
	DeletedStatus  Status = "deleted"
)

// PfandVarianteID names the Pfand-Variante the cart adds per unit of this Variante; nil means no Pfand.
type Variante struct {
	ID              int
	Name            string
	PreisCents      int
	Status          Status
	PfandVarianteID *int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// VarianteMitProdukt verbindet eine Variante mit der ID ihres Produkts: die
// Anreicherung prüft damit die vom Client gesendete Paarung gegen die Datenbank.
// Die Zuordnung ist reine Persistenz und darum kein Feld von Variante.
type VarianteMitProdukt struct {
	Variante  Variante
	ProduktID int
}

// PreisCentsSchema is already required; call sites must not call .Required() again,
// because zog mutates the shared receiver in place.
var PreisCentsSchema = z.Int().
	GTE(1, z.Message("Preis muss mindestens 1 Cent betragen")).
	LTE(99999, z.Message("Preis zu hoch")).
	Required(z.Message("Preis muss mindestens 1 Cent betragen"))

var StatusSchema = z.StringLike[Status]().OneOf(
	[]Status{ActiveStatus, InactiveStatus, DeletedStatus},
	z.Message("Ungültiger Status"),
)

var VarianteSchema = z.Struct(z.Shape{
	"ID":              IDSchema.Required(),
	"Name":            NameSchema.Required(),
	"PreisCents":      PreisCentsSchema,
	"Status":          StatusSchema.Required(),
	"PfandVarianteID": z.Ptr(IDSchema),
	"CreatedAt":       z.Time().Required(),
	"UpdatedAt":       z.Time().Required(),
})

func (v Variante) Validate() error {
	if errs := VarianteSchema.Validate(&v); errs != nil {
		issues := z.Issues.FlattenAndCollect(errs)
		return fmt.Errorf("invalid product variant: %v", issues)
	}
	return nil
}

// NewVariante assigns no ID; the persistence layer sets it. PruefePfandVerknuepfung checks
// pfandVarianteID against the catalogue; this only checks its range.
func NewVariante(name string, preisCents int, pfandVarianteID *int) (Variante, error) {
	if issue := NameSchema.Validate(&name); issue != nil {
		return Variante{}, fmt.Errorf("invalid name")
	}

	if issues := PreisCentsSchema.Validate(&preisCents); issues != nil {
		return Variante{}, fmt.Errorf("%w: %s", ErrPreisUngueltig, issues[0].Message)
	}

	if err := validatePfandVarianteID(pfandVarianteID); err != nil {
		return Variante{}, err
	}

	variante := Variante{
		Name:            name,
		PreisCents:      preisCents,
		Status:          InactiveStatus,
		PfandVarianteID: pfandVarianteID,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	return variante, nil
}

func (v *Variante) Activate() {
	v.Status = ActiveStatus
	v.UpdatedAt = time.Now().UTC()
}

func (v *Variante) Deactivate() {
	v.Status = InactiveStatus
	v.UpdatedAt = time.Now().UTC()
}

func (v *Variante) Delete() {
	v.Status = DeletedStatus
	v.UpdatedAt = time.Now().UTC()
}

func (v *Variante) UpdateDetails(name string, preisCents int, pfandVarianteID *int) error {
	if issue := NameSchema.Validate(&name); issue != nil {
		return fmt.Errorf("invalid name")
	}

	if issues := PreisCentsSchema.Validate(&preisCents); issues != nil {
		return fmt.Errorf("%w: %s", ErrPreisUngueltig, issues[0].Message)
	}

	if err := validatePfandVarianteID(pfandVarianteID); err != nil {
		return err
	}

	v.Name = name
	v.PreisCents = preisCents
	v.PfandVarianteID = pfandVarianteID
	v.UpdatedAt = time.Now().UTC()

	return nil
}

func validatePfandVarianteID(pfandVarianteID *int) error {
	if pfandVarianteID == nil {
		return nil
	}
	if issues := IDSchema.Validate(pfandVarianteID); issues != nil {
		return fmt.Errorf("invalid pfand variante id")
	}
	return nil
}

// PruefePfandVerknuepfung checks that a Variante of quelle may link the Variante zielVarianteID of
// ziel as its Pfand. Order: a valid target first, then no chain, then the same Steuersatz; kombi
// links only kombi, without a special case.
func PruefePfandVerknuepfung(quelle Produkt, ziel Produkt, zielVarianteID int) error {
	idx := slices.IndexFunc(ziel.Varianten, func(v Variante) bool { return v.ID == zielVarianteID })
	if idx < 0 || ziel.Varianten[idx].Status == DeletedStatus || ziel.Status == DeletedStatus || !ziel.Pfand {
		return ErrPfandZielUngueltig
	}
	if quelle.Pfand {
		return ErrPfandVerkettet
	}
	if quelle.Steuersatz != ziel.Steuersatz {
		return ErrPfandSteuersatzAbweichend
	}
	return nil
}

// PruefePfandVerknuepfungen checks every link that starts at or points to a Variante of the Produkt
// produktID within katalog, which holds no deleted Produkt or Variante. A target that lost its Pfand
// flag answers ErrPfandNochVerknuepft instead of ErrPfandZielUngueltig.
func PruefePfandVerknuepfungen(katalog []Produkt, produktID int) error {
	zielProdukt := map[int]int{}
	for i := range katalog {
		for _, v := range katalog[i].Varianten {
			zielProdukt[v.ID] = i
		}
	}

	for i := range katalog {
		quelle := katalog[i]
		for _, v := range quelle.Varianten {
			if v.PfandVarianteID == nil {
				continue
			}
			var ziel Produkt
			if j, ok := zielProdukt[*v.PfandVarianteID]; ok {
				ziel = katalog[j]
			}
			if quelle.ID != produktID && ziel.ID != produktID {
				continue
			}
			err := PruefePfandVerknuepfung(quelle, ziel, *v.PfandVarianteID)
			if errors.Is(err, ErrPfandZielUngueltig) && ziel.ID == produktID && !ziel.Pfand {
				return ErrPfandNochVerknuepft
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
