package repository

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/admin/table/domain"
)

// Repository implements table persistence layer using a SQL database.
type Repository struct {
	DB *sql.DB
}

type dbtable struct {
	ID        int          `db:"id"`
	Name      string       `db:"name"`
	Status    string       `db:"status"`
	CreatedAt sql.NullTime `db:"created_at"`
}

func (dt *dbtable) toDomain() domain.Table {
	return domain.Table{
		ID:        dt.ID,
		Name:      dt.Name,
		Status:    domain.Status(dt.Status),
		CreatedAt: dt.CreatedAt.Time,
	}
}
