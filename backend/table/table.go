package table

import (
	"errors"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/rs/zerolog/log"
)

// Table represents a user in the system.
type Table struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type TablePublic struct {
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
	GetActiveTables() ([]*TablePublic, error)
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
		log.Error().Err(err).Str("name", name).Msg("Failed to create table")
		return nil, ErrDatabase
	}

	table, err := s.Persistence.GetTable(id)
	if err != nil {
		log.Error().Err(err).Int("table_id", id).Msg("Failed to retrieve table after creation")
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
		log.Error().Err(err).Int("table_id", id).Msg("Failed to update table")
		return nil, ErrDatabase
	}

	updatedTable, err := s.Persistence.GetTable(id)
	if err != nil {
		log.Error().Err(err).Int("table_id", id).Msg("Failed to retrieve updated table")
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
		log.Error().Err(err).Int("table_id", id).Msg("Failed to retrieve table")
		return nil, ErrDatabase
	}
	return table, nil
}

// GetAllTables retrieves all tables.
func (s *Service) GetAllTables() ([]*Table, error) {
	tables, err := s.Persistence.GetAllTables()
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve all tables")
		return nil, ErrDatabase
	}
	return tables, nil
}

// GetActiveTables retrieves all active tables.
func (s *Service) GetActiveTables() ([]*TablePublic, error) {
	tables, err := s.Persistence.GetActiveTables()
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve active tables")
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
		log.Error().Err(err).Int("table_id", id).Msg("Failed to activate table")
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
		log.Error().Err(err).Int("table_id", id).Msg("Failed to deactivate table")
		return ErrDatabase
	}
	return nil
}
