package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/sasirekha-dev/go2.0/models"
	"github.com/sasirekha-dev/go2.0/store"
)



func hello(w http.ResponseWriter, r *http.Request){

	fmt.Fprintf(w, "hello\n %s", r.URL.Path[1:])
}

func addTask(w http.ResponseWriter, r *http.Request){
	
	if r.Method != http.MethodPost{
		fmt.Println("Could identify as POST request")
		return 
	}
	fmt.Println("Add request handler")
	var AddRequest store.ToDoItem

	err:= json.NewDecoder(r.Body).Decode(&AddRequest)
	if err != nil{
		return
	}
	err = store.AddTask(AddRequest.Task, AddRequest.Status, r.Context())
	if err!=nil{
		http.Error(w, "not able to save", http.StatusInternalServerError)
	}
}

func deleteTask(w http.ResponseWriter, r *http.Request){
	
	if r.Method != http.MethodDelete{
		fmt.Println("Could identify as DELETE request")
		return 
	}
	ToDoItems, err := store.Read(store.Filename, r.Context())
	if err!=nil{

	}
	queryString, found := strings.CutPrefix(r.URL.RawQuery, "id=")
	if !found{
		http.Error(w, "Error in Request", http.StatusBadRequest)
	}
	item_delete, _:=strconv.Atoi(queryString)
	fmt.Println(item_delete)
	err = store.DeleteTask(item_delete, ToDoItems, r.Context())
	if err!=nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}


func updateTask(w http.ResponseWriter, r *http.Request){	
	if r.Method != http.MethodPut{
		fmt.Println("Could identify as PUT request")
		return 
	}
	type request struct{
		Index int `json:"index"`
		Task string `json:"task"`
		Status string `json:"status"`
	}
	var updateRequest request
	err:= json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = store.UpdateTask(updateRequest.Task, updateRequest.Status, updateRequest.Index, r.Context())
	if err!=nil{
		http.Error(w, "not able to save", http.StatusInternalServerError)
	}
	
}

func TraceIDHandler(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		ctx:= context.WithValue(r.Context(), models.TraceID, uuid.New().String())		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type TraceIDHandle struct{
	slog.Handler
}


func (th *TraceIDHandle) Handle(ctx context.Context, r slog.Record) error{
	//get the value
	traceID := ctx.Value(models.TraceID)
	if trace_id, ok := traceID.(string); ok{
		r.Add(slog.String("traceID:", trace_id))
	}
	return th.Handler.Handle(ctx, r)
}

func main(){
	// ctx = context.WithValue(context.Background(), models.TraceID, uuid.New().String())
	handlerOpts := &slog.HandlerOptions{
		AddSource: true,
		Level: slog.LevelDebug,
	}
	jsonHandler := slog.NewJSONHandler(os.Stderr, handlerOpts)
	NewHandler := &TraceIDHandle{jsonHandler}
	logger := slog.New(NewHandler)
	slog.SetDefault(logger)

	fmt.Println("Server starting...")
	mux := http.NewServeMux()

	mux.HandleFunc("GET /home", hello)
	mux.HandleFunc("/add", addTask)
	mux.HandleFunc("/delete", deleteTask)
	mux.HandleFunc("/update", updateTask)

	fs := http.FileServer(http.Dir("about"))
	http.Handle("/about/", http.StripPrefix("/about/", fs))

	server:= &http.Server{Addr: ":8080", Handler: TraceIDHandler(mux)}

	server.ListenAndServe()
}