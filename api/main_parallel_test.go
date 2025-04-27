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

	"github.com/sasirekha-dev/go2.0/store"
)

func TestParallelAdd(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		log.Fatalf("Error creating temp file")
	}
	defer os.Remove(tempFile.Name())
	store.Filename = tempFile.Name()
	ctx, _ := context.WithCancel(context.Background())
	StartActor(ctx)
	numItemsAdd := 10
	var wg sync.WaitGroup
	for i := 0; i < numItemsAdd; i++ {
		task := fmt.Sprintf("Task-%d", i)
		
		t.Run(task, func(t *testing.T) {
			// t.Parallel()		
			wg.Add(1) // Increment the WaitGroup counter	
			defer wg.Done()
			newTask := store.ToDoItem{Task: task, Status: "started"}
			body, _ := json.Marshal(newTask)
			req := httptest.NewRequest(http.MethodPost, "/add", bytes.NewBuffer(body))
			req = req.WithContext(context.Background())
			req.Header.Set("Content-Type", "application/json")

			reqRecorder := httptest.NewRecorder()
			t.Log("Add request called....")
			AddTask(reqRecorder, req)

			if reqRecorder.Code != http.StatusCreated {
				t.Errorf("Expected status code-201 got %d", reqRecorder.Code)
			}
		})
	}

	wg.Wait()

	t.Run("verify data", func(t *testing.T) {

		var data map[int]store.ToDoItem
		ReadFile, _ := os.Open(store.Filename)
		defer ReadFile.Close()

		decoder := json.NewDecoder(ReadFile)
		decoder.Decode(&data)
		fmt.Printf("The File content - %v", data)
		if len(data) != numItemsAdd {
			t.Errorf("Expected %d items got %d", numItemsAdd, len(data))
		}
	})
}
