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

type ContainerAllInfo struct {
	Status   string
	Image    string
	Tag      string
	Name     string
	CreateAt string
	Ports    []PortMapping
}

type PortMapping struct {
	HostPort      string
	ContainerPort string
}
