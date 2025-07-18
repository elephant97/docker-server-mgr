package types

import "database/sql"

type ContainerInfo struct {
	ID     string         `db:"id"`
	Status sql.NullString `db:"status"`
}

type ContainerStatus struct {
	ID        string         `db:"id"`
	Status    sql.NullString `db:"status"`
	Name      sql.NullString `db:"container_name"`
	DeletedAt sql.NullTime   `db:"deleted_at"`
}

type ImageStatus struct {
	ID     string         `db:"id"`
	Status sql.NullString `db:"status"`
}
