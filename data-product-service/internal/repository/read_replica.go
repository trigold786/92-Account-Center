package repository

import (
	"database/sql"
)

type ReadReplicaRepo struct {
	primary *sql.DB
	replica *sql.DB
}

func NewReadReplicaRepo(primary, replica *sql.DB) *ReadReplicaRepo {
	return &ReadReplicaRepo{primary: primary, replica: replica}
}

func (r *ReadReplicaRepo) ReadDB() *sql.DB {
	if r.replica != nil {
		return r.replica
	}
	return r.primary
}

func (r *ReadReplicaRepo) WriteDB() *sql.DB {
	return r.primary
}

func (r *ReadReplicaRepo) Close() error {
	if r.replica != nil && r.replica != r.primary {
		r.replica.Close()
	}
	return r.primary.Close()
}
