package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"

	"testing"

	"github.com/sasirekha-dev/go2.0/models"
	"github.com/sasirekha-dev/go2.0/store"
)
func TestParallelOptimized(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		log.Fatalf("Error creating temp file")
	}
	defer os.Remove(tempFile.Name())
	store.Filename = tempFile.Name()
	ctx := context.WithValue(context.Background(), models.TraceID, "test")
	StartActor(ctx)

	numItemsAdd := 5

	var wg sync.WaitGroup

	
	wg.Add(numItemsAdd)
	for i := 0; i < numItemsAdd; i++ {
		go func(i int) {  
			defer wg.Done()

			newTask := store.ToDoItem{Task: fmt.Sprintf("AddTask-%d", i), Status: "started"}
			body, _ := json.Marshal(newTask)
			req := httptest.NewRequest(http.MethodPost, "/add/100", bytes.NewBuffer(body))
			req = req.WithContext(context.Background())
			req.Header.Set("Content-Type", "application/json")

			reqRecorder := httptest.NewRecorder()
			AddTask(reqRecorder, req)

			if reqRecorder.Code != http.StatusCreated {
				t.Errorf("Expected status code-201 got %d", reqRecorder.Code)
			}
		}(i)
	}
	wg.Wait() 
	
	wg.Add(numItemsAdd)
	for i := 1; i <= numItemsAdd; i++ {
		go func(i int) {
			defer wg.Done()

			updateTask := map[string]any{
				"index":  i,
				"task":   fmt.Sprintf("UpdateTask-%d", i),
				"status": "completed",
				"userid": "100",
			}
			body, _ := json.Marshal(updateTask)
			req := httptest.NewRequest(http.MethodPut, "/update", bytes.NewBuffer(body))
			req = req.WithContext(context.Background())
			req.Header.Set("Content-Type", "application/json")

			reqRecorder := httptest.NewRecorder()
			UpdateTask(reqRecorder, req)

			if reqRecorder.Code != http.StatusOK {
				t.Errorf("Expected status code-200 got %d", reqRecorder.Code)
			}
		}(i)
	}
	wg.Wait() 
	
	wg.Add(numItemsAdd)
	for i := 1; i <= numItemsAdd; i++ {
		go func(i int) {
			defer wg.Done()

			deleteReq := fmt.Sprintf("/delete?id=%d&user=100", i)
			req := httptest.NewRequest(http.MethodDelete, deleteReq, nil)
			req = req.WithContext(context.Background())
			req.Header.Set("Content-Type", "application/json")

			reqRecorder := httptest.NewRecorder()
			DeleteTask(reqRecorder, req)

			if reqRecorder.Code != http.StatusOK {
				t.Errorf("Expected status code-200 got %d", reqRecorder.Code)
			}
		}(i)
	}
	wg.Wait() 
}

func BenchmarkTodoAdd(b *testing.B) {
	tempFile, err := os.CreateTemp("", "bench_*.json")
	if err != nil {
		b.Fatalf("Error creating temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	store.Filename = tempFile.Name()
	ctx := context.WithValue(context.Background(), models.TraceID, "bench-test")
	StartActor(ctx)

	// Reset timer to ignore setup time
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		newTask := store.ToDoItem{
			Task:   fmt.Sprintf("Task-%d", i),
			Status: "started",
		}
		body, _ := json.Marshal(newTask)
		req := httptest.NewRequest(http.MethodPost, "/add/100", bytes.NewBuffer(body))
		req = req.WithContext(context.Background())
		req.Header.Set("Content-Type", "application/json")

		reqRecorder := httptest.NewRecorder()
		AddTask(reqRecorder, req)

		if reqRecorder.Code != http.StatusCreated {
			b.Errorf("Failed to add task at %d: got status %d", i, reqRecorder.Code)
		}
	}
}
