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

type ServiceTable struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
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
	GetActiveTables() ([]*ServiceTable, error)
	CreateTable(name string) (int, error)
	UpdateTable(id int, name string) error
	ActivateTable(id int) error
	DeactivateTable(id int) error
}

// Service provides table-related operations.
type Service struct {
	Persistence persistence
}

// CreateTable creates a new table in the database.
func (s *Service) CreateTable(name string) (*Table, error) {
	id, err := s.Persistence.CreateTable(name)
	if err != nil {
		log.Printf("ERROR creating table: %v", err)
		return nil, ErrDatabase
	}

	table, err := s.Persistence.GetTable(id)
	if err != nil {
		log.Printf("ERROR Failed to retrieve table %d after creation: %v", id, err)
		return nil, ErrDatabase
	}

	return table, nil
}

// UpdateTable updates an existing table in the database.
func (s *Service) UpdateTable(id int, name string) (*Table, error) {
	err := s.Persistence.UpdateTable(id, name)
	if err != nil {
		if errors.Is(err, ErrTableNotFound) {
			return nil, ErrTableNotFound
		}
		log.Printf("ERROR updating table %d: %v", id, err)
		return nil, ErrDatabase
	}

	updatedTable, err := s.Persistence.GetTable(id)
	if err != nil {
		log.Printf("ERROR Failed to retrieve updated table %d: %v", id, err)
		return nil, ErrDatabase
	}

	return updatedTable, nil
}

// GetTable retrieves a table by its ID.
func (s *Service) GetTable(id int) (*Table, error) {
	table, err := s.Persistence.GetTable(id)
	if err != nil {
		if errors.Is(err, ErrTableNotFound) {
			return nil, ErrTableNotFound
		}
		log.Printf("ERROR retrieving table %d: %v", id, err)
		return nil, ErrDatabase
	}
	return table, nil
}

// GetAllTables retrieves all tables.
func (s *Service) GetAllTables() ([]*Table, error) {
	tables, err := s.Persistence.GetAllTables()
	if err != nil {
		log.Printf("ERROR retrieving all tables: %v", err)
		return nil, ErrDatabase
	}
	return tables, nil
}

// GetActiveTables retrieves all active tables.
func (s *Service) GetActiveTables() ([]*ServiceTable, error) {
	tables, err := s.Persistence.GetActiveTables()
	if err != nil {
		log.Printf("ERROR retrieving active tables: %v", err)
		return nil, ErrDatabase
	}
	return tables, nil
}

// ActivateTable sets the status of a table to active.
func (s *Service) ActivateTable(id int) error {
	err := s.Persistence.ActivateTable(id)
	if err != nil {
		if errors.Is(err, ErrTableNotFound) {
			return ErrTableNotFound
		}
		log.Printf("ERROR activating table %d: %v", id, err)
		return ErrDatabase
	}
	return nil
}

// DeactivateTable sets the status of a table to inactive.
func (s *Service) DeactivateTable(id int) error {
	err := s.Persistence.DeactivateTable(id)
	if err != nil {
		if errors.Is(err, ErrTableNotFound) {
			return ErrTableNotFound
		}
		log.Printf("ERROR deactivating table %d: %v", id, err)
		return ErrDatabase
	}
	return nil
}
