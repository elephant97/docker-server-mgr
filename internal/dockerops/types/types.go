package types

import "database/sql"

type ContainerInfo struct {
	ID     string         `db:"id"`
	Status sql.NullString `db:"status"`
}
