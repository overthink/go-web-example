package taskstore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// PgTaskStore implements task storage and retreival in a Postgres DB.
// Technically there's very little if anything Postgres-specific in here, so
// SqlTaskStore might be a better name, but other dbs haven't been tested.
type PgTaskStore struct {
	pool *sql.DB
}

func (ts *PgTaskStore) CreateTask(ctx context.Context, text string, tags []string, due time.Time) (int, error) {
	query := "INSERT INTO task VALUES(description, tags, due) VALUES ($1, $2, $3) RETURNING id"
	result := 0
	err := ts.pool.QueryRowContext(ctx, query, text, pq.Array(tags), due).Scan(&result)
	if err != nil {
		return 0, fmt.Errorf("failed to create task: %v", err)
	}
	return result, nil
}

func (ts *PgTaskStore) GetTask(_ context.Context, id int) (Task, error) {
	return Task{}, nil
}

func (ts *PgTaskStore) DeleteTask(_ context.Context, id int) error {
	return nil
}

func (ts *PgTaskStore) DeleteAllTasks(_ context.Context) error {
	return nil
}

func (ts *PgTaskStore) GetAllTasks(_ context.Context) ([]Task, error) {
	return nil, nil
}

func (ts *PgTaskStore) GetTasksByTag(_ context.Context, tag string) ([]Task, error) {
	return nil, nil
}

func (ts *PgTaskStore) GetTasksByDueDate(_ context.Context, year int, month time.Month, day int) ([]Task, error) {
	return nil, nil
}
