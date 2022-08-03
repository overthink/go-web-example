package main

import (
	"encoding/json"
	"mime"
	"net/http"
	"os"
	"time"

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

func main() {
	s := NewServer()
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", handlePing)
	mux.HandleFunc("/task/", s.handleCreateTask)
	serverPort, ok := os.LookupEnv("SERVERPORT")
	if !ok {
		serverPort = "60000"
	}
	http.ListenAndServe("localhost:"+serverPort, mux)
}
