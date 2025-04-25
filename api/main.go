package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os/signal"

	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/sasirekha-dev/go2.0/models"
	"github.com/sasirekha-dev/go2.0/store"
)

var requests chan apiRequest
var done chan struct{}

type apiRequest struct {
	verb      string
	task      string
	status    string
	taskID    int
	resp      chan any
	respError chan error
	ctx       context.Context
}

func addTask(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		fmt.Println("Could identify as POST request")
		return
	}
	fmt.Println("Add request handler")
	var AddRequest store.ToDoItem

	err := json.NewDecoder(r.Body).Decode(&AddRequest)
	if err != nil {
		http.Error(w, "Error in POST request", http.StatusInternalServerError)
	}
	error_resp := make(chan error)
	requests <- apiRequest{verb: http.MethodPost, task: AddRequest.Task,
		status: AddRequest.Status, ctx: r.Context(), respError: error_resp}

	if <-error_resp != nil {
		http.Error(w, "Error in POST request", http.StatusBadRequest)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Respond to client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"message": "Task added successfully",
	}
	json.NewEncoder(w).Encode(response)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodDelete {
		fmt.Println("Could identify as DELETE request")
		return
	}

	queryString, found := strings.CutPrefix(r.URL.RawQuery, "id=")
	if !found {
		http.Error(w, "Error in Request", http.StatusBadRequest)
	}
	item_delete, _ := strconv.Atoi(queryString)
	log.Printf("item to delete = %d", item_delete)

	requests <- apiRequest{verb: http.MethodDelete, taskID: item_delete}
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		fmt.Println("Could identify as PUT request")
		return
	}
	type request struct {
		Index  int    `json:"index"`
		Task   string `json:"task"`
		Status string `json:"status"`
	}
	var updateRequest request
	err := json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	requests <- apiRequest{verb: http.MethodPut, task: updateRequest.Task,
		status: updateRequest.Status, taskID: updateRequest.Index, ctx: r.Context()}

}

func listHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fmt.Println("Could identify as GET request")
		return
	}

	responseChan := make(chan any)
	requests <- apiRequest{verb: http.MethodGet, resp: responseChan}
	result := <-responseChan
	tasks, ok := result.(map[int]store.ToDoItem)
	if !ok {
		log.Println("Error in list received")
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}

	tmp, e := template.ParseFiles("api/template/template.html")
	if e != nil {
		log.Printf("Error parse file- %v", e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

	err := tmp.Execute(w, tasks)
	if err != nil {
		log.Printf("Error Execute file- %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func TraceIDHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), models.TraceID, uuid.New().String())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type TraceIDHandle struct {
	slog.Handler
}

func (th *TraceIDHandle) Handle(ctx context.Context, r slog.Record) error {
	//get the value
	traceID := ctx.Value(models.TraceID)
	if trace_id, ok := traceID.(string); ok {
		r.Add(slog.String("traceID", trace_id))
	}
	return th.Handler.Handle(ctx, r)
}

func startActor() {
	log.Println("starting Actor")
	requests = make(chan apiRequest)
	processRequests(requests)
}

func processRequests(requests <-chan apiRequest) {
	done = make(chan struct{})
	// Open database connection
	file := store.Open()
	go func() {
		defer close(done)
		defer store.Close(file) //Closes the database connection
		defer log.Print("Exiting the process go routine...")
		for req := range requests {
			switch req.verb {
			case http.MethodDelete:
				log.Println("delete request received")
				index := req.taskID
				store.DeleteTask(index, req.ctx)
			case http.MethodGet:
				log.Println("get request received")
				tasks, _ := store.Read(req.ctx)
				req.resp <- tasks
			case http.MethodPost:
				log.Println("Post request received")
				err := store.Add(req.task, req.status, req.ctx)
				req.respError <- err
			case http.MethodPut:
				log.Println("put request received")
				store.Update(req.task, req.status, req.taskID, req.ctx)
			default:
				log.Printf("Unidentifiable http method")
			}
		}
	}()
}

func main() {
	handlerOpts := &slog.HandlerOptions{
		// AddSource: true,
		Level: slog.LevelDebug,
	}
	jsonHandler := slog.NewJSONHandler(os.Stderr, handlerOpts)
	NewHandler := &TraceIDHandle{jsonHandler}
	logger := slog.New(NewHandler)
	slog.SetDefault(logger)

	fmt.Println("Server starting...")
	mux := http.NewServeMux()

	mux.HandleFunc("POST /add", addTask)
	mux.HandleFunc("DELETE /delete", deleteTask)
	mux.HandleFunc("PUT /update", updateTask)
	mux.HandleFunc("GET /list", listHandler)

	wd, _ := os.Getwd()
	fmt.Println("Working directory:", wd)

	fs := http.FileServer(http.Dir("api/about"))
	mux.Handle("/about/", http.StripPrefix("/about/", fs))

	server := &http.Server{Addr: ":8080", Handler: TraceIDHandler(mux)}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	startActor()
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			slog.Error("Server error")
		}
	}()

	<-quit
	close(requests)
	<-done
	log.Printf("Received shutdown on ctrl+c")
}
