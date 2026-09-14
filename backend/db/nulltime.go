package db

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"
)

// NullTime is a nullable time implementing sql.Scanner and json.Unmarshaler.
// sql.NullTime implements no json.Unmarshaler, so it cannot scan rows that
// PostgreSQL's json_agg aggregates into JSON arrays (database NULL as well as
// JSON "null" or a timestamp string).
type NullTime struct {
	Time  time.Time
	Valid bool
}

func (nt *NullTime) Scan(value any) error {
	if value == nil {
		nt.Time, nt.Valid = time.Time{}, false
		return nil
	}
	nt.Valid = true
	switch v := value.(type) {
	case time.Time:
		nt.Time = v
	default:
		var snt sql.NullTime
		if err := snt.Scan(value); err != nil {
			return err
		}
		nt.Time, nt.Valid = snt.Time, snt.Valid
	}
	return nil
}

func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

func (nt *NullTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		nt.Time, nt.Valid = time.Time{}, false
		return nil
	}
	if err := json.Unmarshal(data, &nt.Time); err != nil {
		return err
	}
	nt.Valid = true
	return nil
}

func (nt NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nt.Time)
}
