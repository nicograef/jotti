package db

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"
)

// NullTime is a custom nullable time type that implements both sql.Scanner and
// json.Unmarshaler. This is necessary because the standard library's sql.NullTime
// does not implement json.Unmarshaler, which causes issues when using PostgreSQL's
// json_agg function to aggregate rows into JSON arrays. When scanning JSON from
// json_agg results, we need a type that can handle both database NULL values and
// JSON "null" or timestamp strings.
type NullTime struct {
	Time  time.Time
	Valid bool
}

// Scan implements sql.Scanner for database reads.
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
		// Fallback to sql.NullTime for edge cases
		var snt sql.NullTime
		if err := snt.Scan(value); err != nil {
			return err
		}
		nt.Time, nt.Valid = snt.Time, snt.Valid
	}
	return nil
}

// Value implements driver.Valuer for database writes.
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

// UnmarshalJSON implements json.Unmarshaler for JSON parsing.
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

// MarshalJSON implements json.Marshaler for JSON encoding.
func (nt NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nt.Time)
}
