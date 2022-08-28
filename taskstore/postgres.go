package taskstore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/overthink/go-web-example/config"
)

// PgTaskStore implements task storage and retreival in a Postgres DB.
// Technically there's very little if anything Postgres-specific in here, so
// SqlTaskStore might be a better name, but other dbs haven't been tested.
type PgTaskStore struct {
	Pool *sql.DB
}

func NewPgTaskStore(cfg config.Postgres) (*PgTaskStore, error) {
	sslMode := "disable"
	if cfg.SSLEnabled {
		sslMode = "enable"
	}
	connStr := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.UserName,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DbName,
		sslMode,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening db: %v", err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging db: %v", err)
	}
	return &PgTaskStore{Pool: db}, nil
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

type rowScanner interface {
	Scan(dest ...any) error
}

func readTask(row rowScanner) (Task, error) {
	task := Task{}
	var tags pq.StringArray
	if err := row.Scan(&task.Id, &task.Text, &tags, &task.Due); err != nil {
		return Task{}, fmt.Errorf("could not load Task from row: %v", err)
	}
	task.Tags = tags
	task.Due = task.Due.UTC()
	return task, nil
}

func readTasks(rows *sql.Rows) ([]Task, error) {
	result := make([]Task, 0)
	for rows.Next() {
		task, err := readTask(rows)
		if err != nil {
			return result, err
		}
		result = append(result, task)
	}
	return result, nil
}

func (ts *PgTaskStore) GetTask(ctx context.Context, id int) (Task, error) {
	query := "SELECT id, description, tags, due FROM task WHERE id = $1"
	task, err := readTask(ts.Pool.QueryRowContext(ctx, query, id))
	if err != nil {
		return Task{}, err
	}
	return task, nil
}

func (ts *PgTaskStore) DeleteTask(ctx context.Context, id int) error {
	query := "DELETE FROM task WHERE id = $1"
	_, err := ts.Pool.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task with id=%d: %v", id, err)
	}
	return nil
}

func (ts *PgTaskStore) DeleteAllTasks(ctx context.Context) error {
	query := "DELETE FROM task"
	_, err := ts.Pool.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete all tasks: %v", err)
	}
	return nil
}

func (ts *PgTaskStore) GetAllTasks(ctx context.Context) ([]Task, error) {
	query := "SELECT id, description, tags, due FROM task"
	rows, err := ts.Pool.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all tasks: %v", err)
	}
	return readTasks(rows)
}

func (ts *PgTaskStore) GetTasksByTag(ctx context.Context, tag string) ([]Task, error) {
	query := "SELECT id, description, tags, due FROM task WHERE $1 = ANY(tags)"
	rows, err := ts.Pool.QueryContext(ctx, query, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks by tag: %v", err)
	}
	return readTasks(rows)
}

func (ts *PgTaskStore) GetTasksByDueDate(ctx context.Context, year int, month time.Month, day int) ([]Task, error) {
	query := "SELECT id, description, tags, due FROM task WHERE date_trunc('day', due) = $1"
	rows, err := ts.Pool.QueryContext(ctx, query, fmt.Sprintf("%d-%d-%d", year, month, day))
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks by due date: %v", err)
	}
	return readTasks(rows)
}
