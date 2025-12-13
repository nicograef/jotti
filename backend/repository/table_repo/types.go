package table_repo

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/domain/table"
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

func (dt *dbtable) toDomain() table.Table {
	return table.Table{
		ID:        dt.ID,
		Name:      dt.Name,
		Status:    table.Status(dt.Status),
		CreatedAt: dt.CreatedAt.Time,
	}
}
