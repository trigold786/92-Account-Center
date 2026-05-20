package repository

import (
	"context"
	"database/sql"
	"testing"
)

type mockDB struct {
	queryRowFn func(ctx context.Context, query string, args ...interface{}) *sql.Row
	execFn     func(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func (m *mockDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if m.queryRowFn != nil {
		return m.queryRowFn(ctx, query, args...)
	}
	return nil
}

func (m *mockDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if m.execFn != nil {
		return m.execFn(ctx, query, args...)
	}
	return nil, nil
}

func TestReadFromReplica(t *testing.T) {
	primary := &sql.DB{}
	replica := &sql.DB{}

	repo := NewReadReplicaRepo(primary, replica)

	readDB := repo.ReadDB()
	if readDB != replica {
		t.Fatal("ReadDB should return replica when replica is set")
	}
}

func TestReadFromReplicaFallback(t *testing.T) {
	primary := &sql.DB{}

	repo := NewReadReplicaRepo(primary, nil)

	readDB := repo.ReadDB()
	if readDB != primary {
		t.Fatal("ReadDB should return primary when replica is nil")
	}
}

func TestWriteToPrimary(t *testing.T) {
	primary := &sql.DB{}
	replica := &sql.DB{}

	repo := NewReadReplicaRepo(primary, replica)

	writeDB := repo.WriteDB()
	if writeDB != primary {
		t.Fatal("WriteDB should always return primary")
	}
}

func TestWriteToPrimaryNoReplica(t *testing.T) {
	primary := &sql.DB{}

	repo := NewReadReplicaRepo(primary, nil)

	writeDB := repo.WriteDB()
	if writeDB != primary {
		t.Fatal("WriteDB should always return primary")
	}
}
