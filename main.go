package main

import (
	"encoding/json"
	"mime"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/overthink/go-web-example/taskstore"
)

func jsonResponse(w http.ResponseWriter, value interface{}) {
	js, err := json.Marshal(value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func handlePing(w http.ResponseWriter, req *http.Request) {
	jsonResponse(w, map[string]string{"message": "pong"})
}

// All the components/deps required by the app live in this struct.
type server struct {
	taskStore *taskstore.TaskStore
}

func NewServer() *server {
	return &server{taskStore: taskstore.New()}
}

func (s *server) handleCreateTask(w http.ResponseWriter, req *http.Request) {
	type postData struct {
		Text string    `json:"text"`
		Tags []string  `json:"tags"`
		Due  time.Time `json:"due"`
	}
	type response struct {
		Id int `json:"id"`
	}

	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if mediatype != "application/json" {
		http.Error(w, "expect application/json Content-Type", http.StatusUnsupportedMediaType)
		return
	}

	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var pd postData
	if err := dec.Decode(&pd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := s.taskStore.CreateTask(pd.Text, pd.Tags, pd.Due)
	jsonResponse(w, response{id})
}

func (s *server) handleGetTask(w http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(req, "id"))
	if err != nil {
		http.Error(w, "could not parse id", http.StatusBadRequest)
		return
	}
	task, err := s.taskStore.GetTask(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, task)
}

func (s *server) handleGetTasksByTag(w http.ResponseWriter, req *http.Request) {
	tag := chi.URLParam(req, "tag")
	if len(tag) == 0 {
		http.Error(w, "could not parse tag", http.StatusBadRequest)
		return
	}
	jsonResponse(w, s.taskStore.GetTasksByTag(tag))
}

func (s *server) handleGetTasksByDueDate(w http.ResponseWriter, req *http.Request) {
	year, _ := strconv.Atoi(chi.URLParam(req, "yyyy"))
	month, _ := strconv.Atoi(chi.URLParam(req, "mm"))
	day, _ := strconv.Atoi(chi.URLParam(req, "dd"))
	jsonResponse(w, s.taskStore.GetTasksByDueDate(year, time.Month(month), day))
}

func (s *server) handleGetAllTasks(w http.ResponseWriter, req *http.Request) {
	jsonResponse(w, s.taskStore.GetAllTasks())
}

func (s *server) handleDeleteAllTasks(w http.ResponseWriter, req *http.Request) {
	err := s.taskStore.DeleteAllTasks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *server) handleDeleteTask(w http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(req, "id"))
	if err != nil {
		http.Error(w, "could not parse id", http.StatusBadRequest)
		return
	}
	err = s.taskStore.DeleteTask(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}

func main() {
	s := NewServer()
	router := chi.NewRouter()
	router.Get("/ping", handlePing)
	router.Post("/tasks", s.handleCreateTask)
	router.Delete("/tasks", s.handleDeleteAllTasks)
	router.Get("/tasks", s.handleGetAllTasks)
	router.Get("/tasks/{id}", s.handleGetTask)
	router.Delete("/tasks/{id}", s.handleDeleteTask)
	router.Get("/tasks/by-tag/{tag}", s.handleGetTasksByTag)
	router.Get(`/tasks/by-due-date/{yyyy:\d{4}}-{mm:\d\d}-{dd:\d\d}`, s.handleGetTasksByDueDate)
	serverPort, ok := os.LookupEnv("SERVERPORT")
	if !ok {
		serverPort = "60000"
	}
	http.ListenAndServe("localhost:"+serverPort, router)
}
