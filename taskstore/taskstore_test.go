package taskstore

import (
	"context"
	"database/sql"
	"log"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TaskStoreTestSuite struct {
	suite.Suite
	store      TaskStore
	beforeTest func(*TaskStoreTestSuite, string, string)
}

func (suite *TaskStoreTestSuite) BeforeTest(suiteName, testName string) {
	if suite.beforeTest == nil {
		return
	}
	suite.beforeTest(suite, suiteName, testName)
}

func (suite *TaskStoreTestSuite) TestCreateAndGetTask() {
	store := suite.store
	t := suite.T()
	date := time.Now()
	ctx := context.Background()
	id, err := store.CreateTask(ctx, "Hi", []string{"a", "b", "c"}, date)
	assert.NoError(t, err)
	assert.Equal(t, 1, id)
	// assert.Equal(t, 1, store.nextId)

	task, err := store.GetTask(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, task.Id)
	assert.Equal(t, []string{"a", "b", "c"}, task.Tags)
	assert.Equal(t, date.UTC().Round(time.Microsecond), task.Due)

	_, err = store.GetTask(ctx, 999)
	assert.Error(t, err)
}

func (suite *TaskStoreTestSuite) TestDeleteTask() {
	store := suite.store
	t := suite.T()
	ctx := context.Background()
	id, err := store.CreateTask(ctx, "Hi", nil, time.Now())
	assert.NoError(t, err)
	err = store.DeleteTask(ctx, id)
	assert.NoError(t, err)
	_, err = store.GetTask(ctx, id)
	assert.Error(t, err)
}

func (suite *TaskStoreTestSuite) TestDeleteAllTasks() {
	ctx := context.Background()
	store := suite.store
	_, err := store.CreateTask(ctx, "Hi", nil, time.Now())
	suite.NoError(err)
	_, err = store.CreateTask(ctx, "Hi2", nil, time.Now())
	suite.NoError(err)
	tasks, err := store.GetAllTasks(ctx)
	suite.NoError(err)
	suite.Len(tasks, 2)
	suite.NoError(store.DeleteAllTasks(ctx))
	tasks, err = store.GetAllTasks(ctx)
	suite.NoError(err)
	suite.Empty(tasks)
}

func (suite *TaskStoreTestSuite) TestGetAllTasks() {
	store := suite.store
	t := suite.T()
	ctx := context.Background()
	store.CreateTask(ctx, "Hi", nil, time.Now())
	store.CreateTask(ctx, "Hi2", nil, time.Now())
	tasks, err := store.GetAllTasks(ctx)
	assert.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func (suite *TaskStoreTestSuite) TestGetTasksByTag() {
	store := suite.store
	t := suite.T()
	ctx := context.Background()
	store.CreateTask(ctx, "Task1", nil, time.Now())
	store.CreateTask(ctx, "Task2", []string{"a", "b"}, time.Now())
	store.CreateTask(ctx, "Task3", []string{"b", "c", "d"}, time.Now())
	store.CreateTask(ctx, "Task4", []string{"a", "c"}, time.Now())
	tasks, err := store.GetTasksByTag(ctx, "a")
	assert.NoError(t, err)
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Text < tasks[j].Text
	})
	require.Len(t, tasks, 2)
	assert.Equal(t, "Task2", tasks[0].Text)
	assert.Equal(t, "Task4", tasks[1].Text)
}

func mustParseDate(dateStr string) time.Time {
	time, err := time.Parse("2006-Jan-02", dateStr)
	if err != nil {
		log.Fatal(err)
	}
	return time
}

func (suite *TaskStoreTestSuite) TestGetTasksByDueDate() {
	store := suite.store
	t := suite.T()
	ctx := context.Background()
	store.CreateTask(ctx, "Task1", nil, mustParseDate("1995-Feb-02"))
	store.CreateTask(ctx, "Task2", nil, mustParseDate("2011-Jan-14"))
	store.CreateTask(ctx, "Task3", nil, mustParseDate("2008-Aug-13"))
	store.CreateTask(ctx, "Task4", nil, mustParseDate("1995-Feb-02"))
	store.CreateTask(ctx, "Task5", nil, mustParseDate("1980-Mar-05"))

	y, m, d := mustParseDate("1980-Mar-05").Date()
	tasks, err := store.GetTasksByDueDate(ctx, y, m, d)
	assert.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "Task5", tasks[0].Text)

	tests := []struct {
		date        string
		expectedNum int
	}{
		{"2022-Jul-30", 0},
		{"2008-Aug-13", 1},
		{"2011-Jan-14", 1},
		{"1995-Feb-02", 2},
	}
	for _, test := range tests {
		t.Run(test.date, func(t *testing.T) {
			y, m, d := mustParseDate(test.date).Date()
			tasks, err := store.GetTasksByDueDate(ctx, y, m, d)
			assert.NoError(t, err)
			assert.Len(t, tasks, test.expectedNum)
		})
	}
}

func TestInMemoryTaskStore(t *testing.T) {
	cleanDb := func(suite *TaskStoreTestSuite, suiteName, testName string) {
		suite.store = NewInMemTaskStore()
	}
	testSuite := &TaskStoreTestSuite{
		store:      NewInMemTaskStore(),
		beforeTest: cleanDb,
	}
	suite.Run(t, testSuite)
}

func TestPgTaskStore(t *testing.T) {
	connStr := "postgresql://tasks_dev:@localhost/tasks_dev?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, db.PingContext(ctx))

	cleanDb := func(suite *TaskStoreTestSuite, suiteName, testName string) {
		pool := suite.store.(*PgTaskStore).Pool
		_, err := pool.ExecContext(ctx, "DELETE FROM task")
		require.NoError(suite.T(), err)
		_, err = pool.ExecContext(ctx, "ALTER SEQUENCE task_id_seq RESTART WITH 1")
		require.NoError(suite.T(), err)
	}

	testSuite := &TaskStoreTestSuite{
		store:      &PgTaskStore{Pool: db},
		beforeTest: cleanDb,
	}
	suite.Run(t, testSuite)
}
