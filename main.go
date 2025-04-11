package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"

	"go2.0/models"
	"go2.0/store"
)

// create a wrapper around slog.handler
type TraceIDHandler struct{
	slog.Handler
}


func (th *TraceIDHandler) Handle(ctx context.Context, r slog.Record) error{
	//get the value
	traceID := ctx.Value(models.TraceID)
	if trace_id, ok := traceID.(string); ok{
		r.Add(slog.String("traceID:", trace_id))
	}
	return th.Handler.Handle(ctx, r)
}


func main() {

	ctx := context.WithValue(context.Background(), models.TraceID, "12345")

	LOG_FILE := os.Getenv("LOG_FILE")
	file, err := os.OpenFile(LOG_FILE, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	w := io.MultiWriter(os.Stderr, file)
	handlerOpts := &slog.HandlerOptions{
		// AddSource: true,
		Level: slog.LevelDebug,
	}
	jsonHandler := slog.NewJSONHandler(w, handlerOpts)
	NewHandler := &TraceIDHandler{jsonHandler}
	logger := slog.New(NewHandler)
	slog.SetDefault(logger)

	store.ToDoItems, _ = store.Read(store.Filename, ctx)

	add := flag.String("add", "", "Todo item to add")
	delete := flag.Int("delete", 0, "Delete a task")
	update := flag.Int("update", 0, "update a task")
	task := flag.String("task", "", "task description")
	status := flag.String("status", "pending", "status of the task")

	flag.Parse()

	switch {
	case *add != "":
		store.AddTask(*add, *status, ctx)

	case *delete > 0:
		err := store.DeleteTask(*delete, store.ToDoItems, ctx)
		if err != nil {
			fmt.Println("Custom error", err)
		}

	case *update > 0:
		err := store.UpdateTask(*task, *status, *update, ctx)
		if err != nil {
			fmt.Println(err)
		}
	default:
		for i := range store.ToDoItems {
			fmt.Printf("%d: Task: %s, Status: %s \n", i, store.ToDoItems[i].Task, store.ToDoItems[i].Status)
		}

	}
	done:= make(chan struct{})
	//create a cancel channel
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, os.Interrupt)	
	
	go func(){
		fmt.Println("In Go Routine")		
		s:= <-cancelChan				
		slog.InfoContext(ctx, "Received "+s.String())		
		close(done)
		fmt.Println("Go routine ended")
	}()	
	<-done

}

