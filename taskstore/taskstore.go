package taskstore

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Task struct {
	Id   int       `json:"id"`
	Text string    `json:"text"`
	Tags []string  `json:"tags"`
	Due  time.Time `json:"due"`
}

// TaskStore is an in-memory db of tasks. Its methods are threadsafe.
type TaskStore struct {
	sync.Mutex
	tasks  map[int]Task
	nextId int
}

// New is the preferred way to construct a TaskStore.
func New() *TaskStore {
	ts := &TaskStore{}
	ts.tasks = make(map[int]Task)
	ts.nextId = 0
	return ts
}

// CreateTask creates a new task and adds it to the TaskStore. Returns id of
// newly created task.
func (ts *TaskStore) CreateTask(_ context.Context, text string, tags []string, due time.Time) int {
	ts.Lock()
	defer ts.Unlock()
	task := Task{
		Id:   ts.nextId,
		Text: text,
		Tags: make([]string, len(tags)),
		Due:  due,
	}
	copy(task.Tags, tags)
	ts.tasks[task.Id] = task
	ts.nextId++
	return task.Id
}

// GetTask returns the task with the given id, or an error if not found.
func (ts *TaskStore) GetTask(_ context.Context, id int) (Task, error) {
	ts.Lock()
	defer ts.Unlock()
	t, ok := ts.tasks[id]
	if !ok {
		return Task{}, fmt.Errorf("task with id=%d not found", id)
	}
	return t, nil
}

// DeleteTask removes the task with the given id from the store. Returns an
// error if no such task exists.
func (ts *TaskStore) DeleteTask(_ context.Context, id int) error {
	ts.Lock()
	defer ts.Unlock()
	if _, ok := ts.tasks[id]; !ok {
		return fmt.Errorf("task with id=%d not found", id)
	}
	delete(ts.tasks, id)
	return nil
}

// DeleteAllTasks deletes all the tasks in the store. Returns any error.
func (ts *TaskStore) DeleteAllTasks(_ context.Context) error {
	ts.Lock()
	defer ts.Unlock()
	ts.tasks = make(map[int]Task)
	return nil
}

// GetAllTasks returns all the tasks in the store.
func (ts *TaskStore) GetAllTasks(_ context.Context) []Task {
	ts.Lock()
	defer ts.Unlock()
	result := make([]Task, 0, len(ts.tasks))
	for _, task := range ts.tasks {
		result = append(result, task)
	}
	return result
}

func (ts *TaskStore) GetTasksByTag(_ context.Context, tag string) []Task {
	ts.Lock()
	defer ts.Unlock()
	result := make([]Task, 0)
taskloop:
	for _, task := range ts.tasks {
		for _, taskTag := range task.Tags {
			if taskTag == tag {
				result = append(result, task)
				continue taskloop
			}
		}
	}
	return result
}

// GetTasksByDueDate returns all the tasks with the given due date.
func (ts *TaskStore) GetTasksByDueDate(_ context.Context, year int, month time.Month, day int) []Task {
	ts.Lock()
	defer ts.Unlock()
	result := make([]Task, 0)
	for _, task := range ts.tasks {
		y, m, d := task.Due.Date()
		if y == year && m == month && d == day {
			result = append(result, task)
		}
	}
	return result
}
