package table_service

import (
	"errors"

	z "github.com/Oudwins/zog"
)

type Table struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// IDSchema defines the schema for a user ID.
var IDSchema = z.Int().GTE(1, z.Message("Invalid table ID"))

// NameSchema defines the schema for a table's name.
var NameSchema = z.String().Trim().Min(3, z.Message("Name too short")).Max(30, z.Message("Name too long"))

// ErrTableNotFound is returned when a table is not found.
var ErrTableNotFound = errors.New("table not found")

// ErrDatabase is returned when there is a database error.
var ErrDatabase = errors.New("database error")
