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

// var ctx context.Context

type apiResponse struct {
	retVal any
	err    error
}

type apiRequest struct {
	verb   string
	userid string
	task   string
	status string
	taskID int
	resp   chan apiResponse
	ctx    context.Context
}

func AddTask(w http.ResponseWriter, r *http.Request) {

	var AddRequest store.ToDoItem
	if r.Method != http.MethodPost {
		slog.ErrorContext(r.Context(), "Could identify as POST request")
		return
	}
	userid := strings.Split(r.URL.Path, "/")[2]
	
	err := json.NewDecoder(r.Body).Decode(&AddRequest)
	if err != nil {
		e := fmt.Sprintf("Error - %v", err)
		http.Error(w, e, http.StatusInternalServerError)
		slog.ErrorContext(r.Context(), e)
		return
	}
	respChan := make(chan apiResponse)

	Requests <- apiRequest{verb: http.MethodPost, task: AddRequest.Task,
		status: AddRequest.Status, ctx: r.Context(), resp: respChan, userid: userid,}

	resp := <-respChan
	if resp.err != nil {
		http.Error(w, "Error in POST request", http.StatusBadRequest)
		slog.ErrorContext(r.Context(), "Error in POST request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Respond to client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := map[string]string{
		"message": "Task added successfully",
	}
	slog.InfoContext(r.Context(), "Task added successfully")
	json.NewEncoder(w).Encode(response)
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		slog.ErrorContext(r.Context(), "Could identify as DELETE request")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query()
	userid := query.Get("user")
	id := query.Get("id")
	fmt.Println(userid)
	// queryString, found := strings.CutPrefix(r.URL.RawQuery, "id=")
	// if !found {
	// 	http.Error(w, "Error in Request", http.StatusBadRequest)
	// }
	item_delete, _ := strconv.Atoi(id)
	slog.InfoContext(r.Context(), fmt.Sprintf("item to delete = %d", item_delete))
	respChan := make(chan apiResponse)
	Requests <- apiRequest{verb: http.MethodDelete, taskID: item_delete, resp: respChan, userid: userid}
	ret := <-respChan
	if ret.err != nil {
		slog.ErrorContext(r.Context(), ret.err.Error())
		http.Error(w, "Error in DELETE request", http.StatusBadRequest)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": ret.err.Error()})
		return
	}
	// Respond to client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"message": "Task deleted successfully",
	}
	slog.InfoContext(r.Context(), "Task deleted successfully")
	json.NewEncoder(w).Encode(response)
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPut {
		slog.ErrorContext(r.Context(), "Could identify as PUT request")
		return
	}
	type request struct {
		Index  int    `json:"index"`
		Task   string `json:"task"`
		Status string `json:"status"`
		UserId string `json:"userid"`
	}
	var updateRequest request
	err := json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.ErrorContext(r.Context(), err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respChan := make(chan apiResponse)
	Requests <- apiRequest{verb: http.MethodPut, task: updateRequest.Task,
		status: updateRequest.Status, taskID: updateRequest.Index,
		ctx: r.Context(), resp: respChan, userid: updateRequest.UserId}

	ret := <-respChan

	if ret.err != nil {
		http.Error(w, "Error in POST request", http.StatusBadRequest)
		w.WriteHeader(http.StatusInternalServerError)
		slog.ErrorContext(r.Context(), ret.err.Error())
		return
	}
	// Respond to client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"message": "Task updated successfully",
	}
	slog.InfoContext(r.Context(), "Task updated successfully")
	json.NewEncoder(w).Encode(response)
}

func ListHandler(w http.ResponseWriter, r *http.Request) {

	responseChan := make(chan apiResponse)

	if r.Method != http.MethodGet {
		slog.ErrorContext(r.Context(), "Could identify as GET request")
		return
	}
	userid := strings.Split(r.URL.Path, "/")[2]
	if userid==""{
		slog.ErrorContext(r.Context(), "USerid is not found")
	}
	Requests <- apiRequest{verb: http.MethodGet, resp: responseChan, userid: userid}
	result := <-responseChan
	if result.err != nil {
		http.Error(w, "Error in POST request", http.StatusBadRequest)
		w.WriteHeader(http.StatusInternalServerError)
		slog.ErrorContext(r.Context(), result.err.Error())
		return
	}
	slog.InfoContext(r.Context(), "Tasks listing")
	w.WriteHeader(http.StatusOK)

	//Render in html page
	tasks, ok := result.retVal.(map[int]store.ToDoItem)
	if !ok {
		slog.ErrorContext(r.Context(), "Error in list received")
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmp, e := template.ParseFiles("api/template/template.html")
	if e != nil {
		e := fmt.Sprintf("Error parse file- %v", e)
		slog.ErrorContext(r.Context(), e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}
	err := tmp.Execute(w, tasks)
	if err != nil {
		e := fmt.Sprintf("Error Execute file- %v", err)
		slog.ErrorContext(r.Context(), e)
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
	Requests = make(chan apiRequest, 500)
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

		for {
			select {
			case req, ok := <-Requests:
				if !ok {
					log.Print("Request channel is closed...")
				}
				switch req.verb {
				case http.MethodDelete:
					log.Println("delete request received")
					index := req.taskID
					err := store.DeleteTask(req.userid, index, req.ctx)
					req.resp <- apiResponse{nil, err}
				case http.MethodGet:
					log.Println("get request received")
					tasks, err := store.Read(req.userid, req.ctx)
					req.resp <- apiResponse{tasks, err}
				case http.MethodPost:
					log.Println("Post request received")
					err := store.Add(req.task, req.status, req.userid, req.ctx)
					req.resp <- apiResponse{nil, err}
				case http.MethodPut:
					log.Println("put request received")
					err := store.Update(req.userid, req.task, req.status, req.taskID, req.ctx)
					req.resp <- apiResponse{nil, err}
				default:
					log.Printf("Unidentifiable http method")

				}
			case <-ctx.Done():
				slog.InfoContext(ctx, "Context is closing")
				return
			}
		}
	}()
}

func main() {
	handlerOpts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}
	jsonHandler := slog.NewJSONHandler(os.Stderr, handlerOpts)
	NewHandler := &TraceIDHandle{jsonHandler}
	logger := slog.New(NewHandler)
	slog.SetDefault(logger)

	log.Println("Starting server and listening at :8080...")
	mux := http.NewServeMux()

	mux.HandleFunc("POST /add/{userID}", AddTask)
	mux.HandleFunc("DELETE /delete", DeleteTask)
	mux.HandleFunc("PUT /update", UpdateTask)
	mux.HandleFunc("GET /list/{userID}", ListHandler)

	wd, _ := os.Getwd()
	log.Println("Working directory:", wd)

	fs := http.FileServer(http.Dir("api/about"))
	mux.Handle("/about/", http.StripPrefix("/about/", fs))

	server := &http.Server{Addr: ":8080", Handler: TraceIDHandler(mux)}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	StartActor(ctx)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			slog.ErrorContext(ctx, "Server error")
		}
	}()

	<-quit
	cancel()
	<-Done
	log.Printf("Received shutdown on ctrl+c")
}
