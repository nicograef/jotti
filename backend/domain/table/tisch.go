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
	// DeletedStatus: soft-deleted, not visible.
	DeletedStatus Status = "deleted"
)

type Tisch struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

var TischIDSchema = z.Int().GTE(1, z.Message("Invalid table ID"))

var TischNameSchema = z.String().Trim().Min(3, z.Message("Name too short")).Max(100, z.Message("Name too long"))

var TischStatusSchema = z.StringLike[Status]().OneOf(
	[]Status{ActiveStatus, InactiveStatus, DeletedStatus},
	z.Message("Invalid status"),
)

var TischSchema = z.Struct(z.Shape{
	"ID":        TischIDSchema.Required(),
	"Name":      TischNameSchema.Required(),
	"Status":    TischStatusSchema.Required(),
	"CreatedAt": z.Time().Required(),
	"UpdatedAt": z.Time().Required(),
})

func (t Tisch) Validate() error {
	if errs := TischSchema.Validate(&t); errs != nil {
		issues := z.Issues.FlattenAndCollect(errs)
		return fmt.Errorf("invalid table: %v", issues)
	}
	return nil
}

func NewTisch(name string) (Tisch, error) {
	if issue := TischNameSchema.Validate(&name); issue != nil {
		return Tisch{}, errors.New("invalid name")
	}

	tisch := Tisch{
		Name:      name,
		Status:    InactiveStatus,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	return tisch, nil
}

func (t *Tisch) Activate() {
	t.Status = ActiveStatus
	t.UpdatedAt = time.Now().UTC()
}

func (t *Tisch) Deactivate() {
	t.Status = InactiveStatus
	t.UpdatedAt = time.Now().UTC()
}

func (t *Tisch) Delete() {
	t.Status = DeletedStatus
	t.UpdatedAt = time.Now().UTC()
}

func (t *Tisch) Rename(newName string) error {
	if issue := TischNameSchema.Validate(&newName); issue != nil {
		return errors.New("invalid name")
	}
	t.Name = newName
	t.UpdatedAt = time.Now().UTC()
	return nil
}
