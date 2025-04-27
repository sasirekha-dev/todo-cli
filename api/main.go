package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/sasirekha-dev/go2.0/models"
	"github.com/sasirekha-dev/go2.0/store"
)

var Requests chan apiRequest
var Done chan struct{}
var ctx context.Context

type apiRequest struct {
	verb      string
	task      string
	status    string
	taskID    int
	resp      chan any
	respError chan error
	ctx       context.Context
}

func AddTask(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		fmt.Println("Could identify as POST request")
		return
	}

	var AddRequest store.ToDoItem

	err := json.NewDecoder(r.Body).Decode(&AddRequest)
	if err != nil {
		e := fmt.Sprintf("Error - %v", err)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}
	error_resp := make(chan error)
	Requests <- apiRequest{verb: http.MethodPost, task: AddRequest.Task,
		status: AddRequest.Status, ctx: r.Context(), respError: error_resp}

	if <-error_resp != nil {
		http.Error(w, "Error in POST request", http.StatusBadRequest)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Respond to client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := map[string]string{
		"message": "Task added successfully",
	}
	json.NewEncoder(w).Encode(response)
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	error_resp := make(chan error)
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

	Requests <- apiRequest{verb: http.MethodDelete, taskID: item_delete, respError: error_resp}
	if e:= <-error_resp; e != nil {
		slog.ErrorContext(r.Context(), e.Error())
		http.Error(w, "Error in DELETE request", http.StatusBadRequest)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": e.Error()})
		return
	}
	// Respond to client
	
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"message": "Task deleted successfully",
	}
	json.NewEncoder(w).Encode(response)
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	error_resp := make(chan error)
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

	Requests <- apiRequest{verb: http.MethodPut, task: updateRequest.Task,
		status: updateRequest.Status, taskID: updateRequest.Index,
		ctx: r.Context(), respError: error_resp}

	if e := <-error_resp; e != nil {
		http.Error(w, "Error in POST request", http.StatusBadRequest)
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(e)
		return
	}
	// Respond to client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"message": "Task updated successfully",
	}
	json.NewEncoder(w).Encode(response)
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	errChan := make(chan error)
	responseChan := make(chan any)

	if r.Method != http.MethodGet {
		fmt.Println("Could identify as GET request")
		return
	}

	Requests <- apiRequest{verb: http.MethodGet, resp: responseChan, respError: errChan}
	result := <-responseChan
	if e := <-errChan; e != nil {
		http.Error(w, "Error in POST request", http.StatusBadRequest)
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(e)
		return
	}
	w.WriteHeader(http.StatusOK)

	//Render in html page

	tasks, ok := result.(map[int]store.ToDoItem)
	if !ok {
		log.Println("Error in list received")
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

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

func StartActor(ctx context.Context) {
	log.Println("starting Actor")
	Requests = make(chan apiRequest)
	Done = make(chan struct{})

	processRequests(ctx, Requests)
}

func processRequests(ctx context.Context, Requests <-chan apiRequest) {
	Done = make(chan struct{})
	// Open database connection
	file := store.Open()
	go func() {
		defer close(Done)
		defer store.Close(file) //Closes the database connection
		defer log.Print("Exiting the process go routine...")

		for req := range Requests{
			switch req.verb {
			case http.MethodDelete:
				log.Println("delete request received")
				index := req.taskID
				err:=store.DeleteTask(index, req.ctx)
				req.respError <- err
			case http.MethodGet:
				log.Println("get request received")
				tasks, err := store.Read(req.ctx)
				req.resp <- tasks
				req.respError<-err
			case http.MethodPost:
				log.Println("Post request received")
				err := store.Add(req.task, req.status, req.ctx)
				req.respError <- err
			case http.MethodPut:
				log.Println("put request received")
				err:=store.Update(req.task, req.status, req.taskID, req.ctx)
				req.respError<-err
			default:
				log.Printf("Unidentifiable http method")
			}
		}

}()
}


func main() {
	handlerOpts := &slog.HandlerOptions{
		AddSource: true,
		Level: slog.LevelDebug,
	}
	jsonHandler := slog.NewJSONHandler(os.Stderr, handlerOpts)
	NewHandler := &TraceIDHandle{jsonHandler}
	logger := slog.New(NewHandler)
	slog.SetDefault(logger)

	// ctx, cancel := context.WithCancel(ctx)
	// defer cancel()
	fmt.Println("Starting server and listening at :8080...")
	mux := http.NewServeMux()

	mux.HandleFunc("POST /add", AddTask)
	mux.HandleFunc("DELETE /delete", DeleteTask)
	mux.HandleFunc("PUT /update", UpdateTask)
	mux.HandleFunc("GET /list", ListHandler)

	wd, _ := os.Getwd()
	fmt.Println("Working directory:", wd)

	fs := http.FileServer(http.Dir("api/about"))
	mux.Handle("/about/", http.StripPrefix("/about/", fs))

	server := &http.Server{Addr: ":8080", Handler: TraceIDHandler(mux)}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	StartActor(ctx)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			slog.Error("Server error")
		}
	}()

	<-quit
	close(Requests)
	<-Done
	log.Printf("Received shutdown on ctrl+c")
}
