package table

import (
	"fmt"
	"strconv"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

type ausgabeBestaetigtV1Data struct {
	AusgabeID  string              `json:"ausgabeId"`
	Positionen []positionEventData `json:"positionen"`
	Kommentar  string              `json:"kommentar"`
}

var ausgabeBestaetigtV1DataSchema = z.Struct(z.Shape{
	"AusgabeID":  z.String().UUID().Required(),
	"Positionen": z.Slice(positionSchema).Min(1).Required(),
	"Kommentar":  z.String().Max(100),
})

func NewAusgabeBestaetigtEvent(userID int, userName string, tischID int, positionen []Position, kommentar string) (e.Event, error) {
	data := ausgabeBestaetigtV1Data{
		AusgabeID:  uuid.New().String(),
		Positionen: toPositionenEventData(positionen),
		Kommentar:  kommentar,
	}

	if err := ausgabeBestaetigtV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("ausgabe bestaetigt data validation failed: %v", issues)
	}

	event, err := e.New(userID, userName, string(EventTypeAusgabeBestaetigtV1), "tisch:"+strconv.Itoa(tischID), data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}

func buildAusgabeFromEvent(event e.Event) (Ausgabe, error) {
	if event.Type != string(EventTypeAusgabeBestaetigtV1) {
		return Ausgabe{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tischID, err := parseTischIDFromSubject(event.Subject)
	if err != nil {
		return Ausgabe{}, err
	}

	data := ausgabeBestaetigtV1Data{}
	err = e.ParseData(event, &data, ausgabeBestaetigtV1DataSchema)
	if err != nil {
		return Ausgabe{}, err
	}

	ausgabe := Ausgabe{
		ID:           data.AusgabeID,
		UserID:       event.UserID,
		TischID:      tischID,
		Positionen:   fromPositionenEventData(data.Positionen),
		Kommentar:    data.Kommentar,
		AusgegebenAm: event.Time,
	}

	if err := ausgabeSchema.Validate(&ausgabe); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return Ausgabe{}, fmt.Errorf("ausgabe validation failed: %v", issues)
	}

	return ausgabe, nil
}
