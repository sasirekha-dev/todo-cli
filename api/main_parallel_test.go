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

	"testing"

	"github.com/sasirekha-dev/go2.0/store"
)

func TestParallel(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		log.Fatalf("Error creating temp file")
	}
	defer os.Remove(tempFile.Name())
	store.Filename = tempFile.Name()
	ctx, _ := context.WithCancel(context.Background())
	StartActor(ctx)
	numItemsAdd := 2

	for i := 0; i < numItemsAdd; i++ {
		task := fmt.Sprintf("AddTask-%d", i)
		t.Run(task, func(t *testing.T) {
			t.Parallel()
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
	for i := 1; i < numItemsAdd+1; i++ {
		task := fmt.Sprintf("UpdateTask-%d", i)
		t.Run(task, func(t *testing.T) {
			t.Parallel()
			updateTask := map[string]any{
				"index":  i,
				"task":   task,
				"status": "completed",
			}
			body, _ := json.Marshal(updateTask)
			req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(body))
			req = req.WithContext(context.Background())
			req.Header.Set("Content-Type", "application/json")

			reqRecorder := httptest.NewRecorder()
			t.Log("Update request called....")
			UpdateTask(reqRecorder, req)

			if reqRecorder.Code != http.StatusOK {
				t.Errorf("Expected status code-200 got %d", reqRecorder.Code)
			}
		})
	}
	for i := 1; i < numItemsAdd+1; i++ {
		task := fmt.Sprintf("UpdateTask-%d", i)
		t.Run(task, func(t *testing.T) {
			t.Parallel()
			deleteReq := fmt.Sprintf("/delete?id=%d",i)

			req := httptest.NewRequest(http.MethodPost, deleteReq, nil)
			req = req.WithContext(context.Background())
			req.Header.Set("Content-Type", "application/json")

			reqRecorder := httptest.NewRecorder()
			t.Log("Delete request called....")
			DeleteTask(reqRecorder, req)

			if reqRecorder.Code != http.StatusOK {
				t.Errorf("Expected status code-200 got %d", reqRecorder.Code)
			}
		})
	}

	
}
