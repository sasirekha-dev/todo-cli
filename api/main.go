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
var ctx context.Context


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
	err = store.AddTask(AddRequest.Task, AddRequest.Status, ctx)
	if err!=nil{
		http.Error(w, "not able to save", http.StatusInternalServerError)
	}
}

func deleteTask(w http.ResponseWriter, r *http.Request){
	
	if r.Method != http.MethodDelete{
		fmt.Println("Could identify as DELETE request")
		return 
	}
	ToDoItems, err := store.Read(store.Filename, ctx)
	if err!=nil{

	}
	queryString, found := strings.CutPrefix(r.URL.RawQuery, "id=")
	if !found{
		http.Error(w, "Error in Request", http.StatusBadRequest)
	}
	item_delete, _:=strconv.Atoi(queryString)
	fmt.Println(item_delete)
	err = store.DeleteTask(item_delete, ToDoItems, ctx)
	if err!=nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}


func updateTask(w http.ResponseWriter, r *http.Request){	
	if r.Method != http.MethodPut{
		fmt.Println("Could identify as PUT request")
		return 
	}
	ToDoItems, err := store.Read(store.Filename, ctx)
	if err!=nil{

	}
	queryString, found := strings.CutPrefix(r.URL.RawQuery, "id=")
	if !found{
		http.Error(w, "Error in Request", http.StatusBadRequest)
	}
	item_delete, _:=strconv.Atoi(queryString)
	fmt.Println(item_delete)
	err = store.DeleteTask(item_delete, ToDoItems, ctx)
	if err!=nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func main(){
	ctx = context.WithValue(context.Background(), models.TraceID, uuid.New().String())
	handlerOpts := &slog.HandlerOptions{
		AddSource: true,
		Level: slog.LevelDebug,
	}
	jsonHandler := slog.NewJSONHandler(os.Stderr, handlerOpts)
	logger := slog.New(jsonHandler)
	slog.SetDefault(logger)

	fmt.Println("Server starting...")

	

	http.HandleFunc("/home", hello)
	http.HandleFunc("/add", addTask)
	http.HandleFunc("/delete", deleteTask)

	fs := http.FileServer(http.Dir("about"))
	http.Handle("/about/", http.StripPrefix("/about/", fs))

	http.ListenAndServe(":8080", nil)
}