package taskstore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// PgTaskStore implements task storage and retreival in a Postgres DB.
// Technically there's very little if anything Postgres-specific in here, so
// SqlTaskStore might be a better name, but other dbs haven't been tested.
type PgTaskStore struct {
	Pool *sql.DB
}

func (ts *PgTaskStore) CreateTask(ctx context.Context, text string, tags []string, due time.Time) (int, error) {
	query := "INSERT INTO task (description, tags, due) " +
		"VALUES ($1, $2, $3) RETURNING id"
	result := 0
	err := ts.Pool.
		QueryRowContext(
			ctx,
			query,
			text,
			pq.Array(tags),
			// Round due date to micros for clarity. Postgres will truncate
			// to micros anyway, but I'd rather make it explicit.
			due.UTC().Round(time.Microsecond),
		).
		Scan(&result)
	if err != nil {
		return 0, fmt.Errorf("failed to create task: %v", err)
	}
	return result, nil
}

func (ts *PgTaskStore) GetTask(ctx context.Context, id int) (Task, error) {
	query := "SELECT id, description, tags, due FROM task WHERE id = $1"
	task := Task{}
	var tags pq.StringArray
	err := ts.Pool.QueryRowContext(ctx, query, id).
		Scan(&task.Id, &task.Text, &tags, &task.Due)
	if err != nil {
		return Task{}, fmt.Errorf("failed to get task id=%d: %v", id, err)
	}
	task.Tags = tags
	task.Due = task.Due.UTC() // pg's repr of UTC time zone is slightly different than go's, so convert it
	return task, nil
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
