package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"os/signal"
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
type app struct {
	config    Config
	taskStore taskstore.TaskStore
}

func (a *app) handleCreateTask(w http.ResponseWriter, req *http.Request) {
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

	id, err := a.taskStore.CreateTask(req.Context(), pd.Text, pd.Tags, pd.Due)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, response{id})
}

func (a *app) handleGetTask(w http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(req, "id"))
	if err != nil {
		http.Error(w, "could not parse id", http.StatusBadRequest)
		return
	}
	task, err := a.taskStore.GetTask(req.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, task)
}

func (a *app) handleGetTasksByTag(w http.ResponseWriter, req *http.Request) {
	tag := chi.URLParam(req, "tag")
	if len(tag) == 0 {
		http.Error(w, "could not parse tag", http.StatusBadRequest)
		return
	}
	tasks, err := a.taskStore.GetTasksByTag(req.Context(), tag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, tasks)
}

func (a *app) handleGetTasksByDueDate(w http.ResponseWriter, req *http.Request) {
	year, _ := strconv.Atoi(chi.URLParam(req, "yyyy"))
	month, _ := strconv.Atoi(chi.URLParam(req, "mm"))
	day, _ := strconv.Atoi(chi.URLParam(req, "dd"))
	tasks, err := a.taskStore.GetTasksByDueDate(req.Context(), year, time.Month(month), day)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, tasks)
}

func (a *app) handleGetAllTasks(w http.ResponseWriter, req *http.Request) {
	tasks, err := a.taskStore.GetAllTasks(req.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, tasks)
}

func (a *app) handleDeleteAllTasks(w http.ResponseWriter, req *http.Request) {
	err := a.taskStore.DeleteAllTasks(req.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *app) handleDeleteTask(w http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(req, "id"))
	if err != nil {
		http.Error(w, "could not parse id", http.StatusBadRequest)
		return
	}
	err = a.taskStore.DeleteTask(req.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}

func createServer(config HttpServerConfig, router *chi.Mux) (*http.Server, chan struct{}) {
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.ListenAddress, config.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(config.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(config.WriteTimeoutSeconds) * time.Second,
		ConnState: func(conn net.Conn, state http.ConnState) {
			log.Printf("ConnState: %v, %v", conn.RemoteAddr(), state)
		},
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		log.Printf("caught sigint, shutting down")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("error when shutting down: %v", err)
		}
		close(idleConnsClosed)
	}()
	return server, idleConnsClosed
}

func main() {
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	app := &app{
		config:    config,
		taskStore: taskstore.NewInMemTaskStore(),
	}

	router := chi.NewRouter()
	router.Get("/ping", handlePing)
	router.Post("/tasks", app.handleCreateTask)
	router.Delete("/tasks", app.handleDeleteAllTasks)
	router.Get("/tasks", app.handleGetAllTasks)
	router.Get("/tasks/{id}", app.handleGetTask)
	router.Delete("/tasks/{id}", app.handleDeleteTask)
	router.Get("/tasks/by-tag/{tag}", app.handleGetTasksByTag)
	router.Get(`/tasks/by-due-date/{yyyy:\d{4}}-{mm:\d\d}-{dd:\d\d}`, app.handleGetTasksByDueDate)

	server, idleConnsClosed := createServer(config.HttpServer, router)
	log.Println("server started")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("HTTP server ListenAndServe: %v", err)
	}
	<-idleConnsClosed
	log.Println("exiting")

}
