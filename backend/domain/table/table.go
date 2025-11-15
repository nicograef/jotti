package table

import (
	"errors"
	"log"
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

// ErrDatabase is returned when there is a database error.
var ErrDatabase = errors.New("database error")

type persistence interface {
	GetTable(id int) (*Table, error)
	GetAllTables() ([]*Table, error)
	CreateTable(name string) (int, error)
	UpdateTable(id int, name string) error
}

// Service provides table-related operations.
type Service struct {
	DB persistence
}

// CreateTable creates a new table in the database.
func (s *Service) CreateTable(name string) (*Table, error) {
	id, err := s.DB.CreateTable(name)
	if err != nil {
		log.Printf("Error creating table: %v", err)
		return nil, ErrDatabase
	}

	return s.DB.GetTable(id)
}

// UpdateTable updates an existing table in the database.
func (s *Service) UpdateTable(id int, name string) (*Table, error) {
	err := s.DB.UpdateTable(id, name)
	if err != nil {
		if errors.Is(err, ErrTableNotFound) {
			return nil, ErrTableNotFound
		}
		log.Printf("Error updating table: %v", err)
		return nil, ErrDatabase
	}

	return s.DB.GetTable(id)
}

// GetTable retrieves a table by its ID.
func (s *Service) GetTable(id int) (*Table, error) {
	table, err := s.DB.GetTable(id)
	if err != nil {
		if errors.Is(err, ErrTableNotFound) {
			return nil, ErrTableNotFound
		}
		log.Printf("Error retrieving table: %v", err)
		return nil, ErrDatabase
	}
	return table, nil
}

// GetAllTables retrieves all tables.
func (s *Service) GetAllTables() ([]*Table, error) {
	tables, err := s.DB.GetAllTables()
	if err != nil {
		log.Printf("Error retrieving tables: %v", err)
		return nil, ErrDatabase
	}
	return tables, nil
}
