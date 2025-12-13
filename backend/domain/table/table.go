package table

import (
	"errors"
	"fmt"
	"time"

	z "github.com/Oudwins/zog"
)

type Status string

const (
	// ActiveStatus: usable for service.
	ActiveStatus Status = "active"
	// InactiveStatus: not usable for service.
	InactiveStatus Status = "inactive"
)

type Table struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

var IDSchema = z.Int().GTE(1, z.Message("Invalid table ID"))

var NameSchema = z.String().Trim().Min(3, z.Message("Name too short")).Max(30, z.Message("Name too long"))

var StatusSchema = z.StringLike[Status]().OneOf(
	[]Status{ActiveStatus, InactiveStatus},
	z.Message("Invalid status"),
)

var TableSchema = z.Struct(z.Shape{
	"ID":        IDSchema.Required(),
	"Name":      NameSchema.Required(),
	"Status":    StatusSchema.Required(),
	"CreatedAt": z.Time().Required(),
})

func (p Table) Validate() error {
	if errsMap := TableSchema.Validate(&p); errsMap != nil {
		issues := z.Issues.SanitizeMapAndCollect(errsMap)
		return fmt.Errorf("invalid table: %v", issues)
	}
	return nil
}

// NewTable creates a new Table instance after validating the input parameters.
// The new Table does not have an ID assigned; it is expected to be set by the persistence layer.
func NewTable(name string) (Table, error) {
	if issue := NameSchema.Validate(&name); issue != nil {
		return Table{}, errors.New("invalid name")
	}

	table := Table{
		Name:      name,
		Status:    InactiveStatus,
		CreatedAt: time.Now().UTC(),
	}

	return table, nil
}

func (p *Table) Activate() {
	p.Status = ActiveStatus
}

func (p *Table) Deactivate() {
	p.Status = InactiveStatus
}

func (p *Table) Rename(newName string) error {
	if issue := NameSchema.Validate(&newName); issue != nil {
		return errors.New("invalid name")
	}
	p.Name = newName
	return nil
}
