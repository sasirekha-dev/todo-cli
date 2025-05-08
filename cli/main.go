package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"

	"github.com/google/uuid"
	"github.com/sasirekha-dev/go2.0/models"
	"github.com/sasirekha-dev/go2.0/store"
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

	ctx := context.WithValue(context.Background(), models.TraceID, uuid.New().String())

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

	
	
	add := flag.String("add", "", "Todo item to add")
	delete := flag.Int("delete", 0, "Delete a task")
	update := flag.Int("update", 0, "update a task")
	task := flag.String("task", "", "task description")
	status := flag.String("status", "pending", "status of the task")
	userid := flag.String("userid", "", "ID of user")

	flag.Parse()
	// store.ToDoItems, _ = store.Read(userid, ctx)

	switch {
	case *add != "":
		err:= store.Add(*add, *status, *userid, ctx)
		if err != nil {
			slog.ErrorContext(ctx, err.Error())
		}
	case *delete > 0:
		err := store.DeleteTask(*userid, *delete, ctx)
		if err != nil {
			slog.ErrorContext(ctx, err.Error())
		}

	case *update > 0:
		err := store.Update(*userid,*task, *status, *update, ctx)
		if err != nil {
			slog.ErrorContext(ctx, err.Error())
		}
	default:
		toDoList := store.Load(*userid)
		fmt.Println(toDoList)
		// for i := range toDoList {
		// 	fmt.Printf("%d. Task: %s, Status: %s \n", i, store.ToDoItems[i].Task, store.ToDoItems[i].Status)
		// }
	}
	
	done:= make(chan bool)
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

