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
	UpdatedAt    time.Time
}

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
	b := Betreiber{
		Vereinsname:  vereinsname,
		Strasse:      strasse,
		Plz:          plz,
		Ort:          ort,
		Steuernummer: steuernummer,
		UstID:        ustId,
		UpdatedAt:    time.Now().UTC(),
	}
	if err := b.Validate(); err != nil {
		return Betreiber{}, err
	}
	return b, nil
}
