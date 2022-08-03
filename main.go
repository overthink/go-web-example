package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/overthink/go-web-example/taskstore"
)

// Struct to eventually hold all components needed by this app.
type taskServer struct {
	taskStore *taskstore.TaskStore
}

func (ts *taskServer) pingHandler(w http.ResponseWriter, req *http.Request) {
	json, err := json.Marshal(map[string]string{"message": "pong"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.Write(json)
}

func NewTaskServer() *taskServer {
	return &taskServer{taskStore: taskstore.New()}
}

func main() {
	server := NewTaskServer()
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", server.pingHandler)
	serverPort, ok := os.LookupEnv("SERVERPORT")
	if !ok {
		serverPort = "60000"
	}
	http.ListenAndServe("localhost:"+serverPort, mux)
}
