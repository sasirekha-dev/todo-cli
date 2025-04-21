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

func hello(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "hello\n %s", r.URL.Path[1:])
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
		return
	}
	err = store.AddTask(AddRequest.Task, AddRequest.Status, r.Context())
	if err != nil {
		http.Error(w, "not able to save", http.StatusInternalServerError)
	}
}

func deleteTask(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodDelete {
		fmt.Println("Could identify as DELETE request")
		return
	}
	ToDoItems, err := store.Read(store.Filename, r.Context())
	if err != nil {

	}
	queryString, found := strings.CutPrefix(r.URL.RawQuery, "id=")
	if !found {
		http.Error(w, "Error in Request", http.StatusBadRequest)
	}
	item_delete, _ := strconv.Atoi(queryString)

	err = store.DeleteTask(item_delete, ToDoItems, r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
	err = store.UpdateTask(updateRequest.Task, updateRequest.Status, updateRequest.Index, r.Context())
	if err != nil {
		http.Error(w, "not able to save", http.StatusInternalServerError)
	}

}

func listHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fmt.Println("Could identify as GET request")
		return
	}
	tasks, err := store.Read(store.Filename, r.Context())
	if err != nil {
		slog.Error("list operation from database failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	tmp, e := template.ParseFiles("api/template/template.html")
	if e != nil {
		log.Printf("Error parse file- %v", e)
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

	error := tmp.Execute(w, tasks)
	if error != nil {
		log.Printf("Error Execute file- %v", error)
		// slog.Error("Failed to parse template file")
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

func main() {
	handlerOpts := &slog.HandlerOptions{
		// AddSource: true,
		Level:     slog.LevelDebug,
	}
	jsonHandler := slog.NewJSONHandler(os.Stderr, handlerOpts)
	NewHandler := &TraceIDHandle{jsonHandler}
	logger := slog.New(NewHandler)
	slog.SetDefault(logger)

	fmt.Println("Server starting...")
	mux := http.NewServeMux()

	mux.HandleFunc("GET /home", hello)
	mux.HandleFunc("POST /add", addTask)
	mux.HandleFunc("DELETE /delete", deleteTask)
	mux.HandleFunc("PUT /update", updateTask)
	mux.HandleFunc("GET /list", listHandler)

	wd, _ := os.Getwd()
	fmt.Println("Working directory:", wd)

	fs := http.FileServer(http.Dir("api/about"))
	mux.Handle("/about/", http.StripPrefix("/about/", fs))

	server := &http.Server{Addr: ":8080", Handler: TraceIDHandler(mux)}
	quit:= make(chan os.Signal,1)
	signal.Notify(quit, os.Interrupt)
	go func(){
		log.Print("In Go Routine....")
		err := server.ListenAndServe()
		if err != nil {
			slog.Error("Server error")
		}
	}()
	<-quit
	log.Printf("Received shutdown on ctrl+c")
}
