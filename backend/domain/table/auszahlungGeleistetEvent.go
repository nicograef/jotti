package table

import (
	"fmt"
	"strconv"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

type Auszahlung struct {
	ID          string
	UserID      int
	TischID     int
	BetragCents int
	Kommentar   string
	GeleistetAm time.Time
}

var auszahlungSchema = z.Struct(z.Shape{
	"ID":          z.String().UUID().Required(),
	"UserID":      z.Int().GTE(1).Required(),
	"TischID":     z.Int().GTE(1).Required(),
	"BetragCents": z.Int().GTE(1).Required(),
	"Kommentar":   z.String().Min(3).Max(100).Required(),
	"GeleistetAm": z.Time().Required(),
})

type auszahlungGeleistetV1Data struct {
	AuszahlungID string `json:"auszahlungId"`
	BetragCents  int    `json:"betragCents"`
	Kommentar    string `json:"kommentar"`
}

var auszahlungGeleistetV1DataSchema = z.Struct(z.Shape{
	"AuszahlungID": z.String().UUID().Required(),
	"BetragCents":  z.Int().GTE(1).Required(),
	"Kommentar":    z.String().Min(3).Max(100).Required(),
})

func NewAuszahlungGeleistetEvent(userID int, userName string, tischID int, betragCents int, kommentar string) (e.Event, error) {
	data := auszahlungGeleistetV1Data{
		AuszahlungID: uuid.New().String(),
		BetragCents:  betragCents,
		Kommentar:    kommentar,
	}

	if err := auszahlungGeleistetV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("auszahlung geleistet data validation failed: %v", issues)
	}

	event, err := e.New(userID, userName, string(EventTypeAuszahlungGeleistetV1), "tisch:"+strconv.Itoa(tischID), data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}

func buildAuszahlungFromEvent(event e.Event) (Auszahlung, error) {
	if event.Type != string(EventTypeAuszahlungGeleistetV1) {
		return Auszahlung{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tischID, err := parseTischIDFromSubject(event.Subject)
	if err != nil {
		return Auszahlung{}, err
	}

	data := auszahlungGeleistetV1Data{}
	err = e.ParseData(event, &data, auszahlungGeleistetV1DataSchema)
	if err != nil {
		return Auszahlung{}, err
	}

	auszahlung := Auszahlung{
		ID:          data.AuszahlungID,
		UserID:      event.UserID,
		TischID:     tischID,
		BetragCents: data.BetragCents,
		Kommentar:   data.Kommentar,
		GeleistetAm: event.Time,
	}

	if err := auszahlungSchema.Validate(&auszahlung); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return Auszahlung{}, fmt.Errorf("auszahlung validation failed: %v", issues)
	}

	return auszahlung, nil
}
