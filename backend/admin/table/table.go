package table

import (
	"errors"
	"time"

	z "github.com/Oudwins/zog"
)

// Table represents a user in the system.
type Table struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

// IDSchema defines the schema for a user ID.
var IDSchema = z.Int().GTE(1, z.Message("Invalid table ID"))

// NameSchema defines the schema for a table's name.
var NameSchema = z.String().Trim().Min(3, z.Message("Name too short")).Max(30, z.Message("Name too long"))

// StatusSchema defines the schema for a table status.
var StatusSchema = z.StringLike[Status]().OneOf(
	[]Status{ActiveStatus, InactiveStatus, DeletedStatus},
	z.Message("Invalid status"),
)

// ErrTableNotFound is returned when a table is not found.
var ErrTableNotFound = errors.New("table not found")

// ErrTableAlreadyExists is returned when a table already exists.
var ErrTableAlreadyExists = errors.New("table already exists")

// ErrDatabase is returned when there is a database error.
var ErrDatabase = errors.New("database error")

// Status represents the status of a table.
type Status string

const (
	// ActiveStatus indicates the table is active and usable for service.
	ActiveStatus Status = "active"
	// InactiveStatus indicates the table is inactive and not currently in use.
	InactiveStatus Status = "inactive"
	// DeletedStatus indicates the table has been deleted and is no longer in use.
	DeletedStatus Status = "deleted"
)
