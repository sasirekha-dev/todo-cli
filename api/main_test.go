package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"github.com/sasirekha-dev/go2.0/store"
)

func Setup(t *testing.T){
	tempFile, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		log.Fatalf("Error creating temp file")
	}
	defer os.Remove(tempFile.Name())
	store.Filename = tempFile.Name()
}

func TestAddHandler(t *testing.T) {
	Setup(t)
	
	newTask := store.ToDoItem{Task: "handler func", Status: "started"}
	body, _ := json.Marshal(newTask)
	req := httptest.NewRequest(http.MethodPost, "/add", bytes.NewBuffer(body))

	req = req.WithContext(context.Background())
	req.Header.Set("Content-Type", "application/json")

	reqRecorder := httptest.NewRecorder()
	// Start the actor first
	ctx,_:=context.WithCancel(context.Background())
	StartActor(ctx)
	defer func() {
		close(Requests)
		<-Done
	}()
	AddTask(reqRecorder, req)

	if reqRecorder.Code != http.StatusCreated {
		t.Errorf("Expected status code-201 got %d", reqRecorder.Code)
	}
}

func TestUpdate(t *testing.T){
	updatePayload := map[string]any{
		"index":  1,
		"task":   "Updated Task",
		"status": "completed",
	}
	body, _:=json.Marshal(updatePayload)
	req := httptest.NewRequest(http.MethodPut, "/update", bytes.NewBuffer(body))
	req = req.WithContext(context.Background())
	reqRecorder := httptest.NewRecorder()
	ctx,_:=context.WithCancel(context.Background())
	StartActor(ctx)
	defer func() {
		close(Requests)
		<-Done
	}()
	UpdateTask(reqRecorder, req)
	if reqRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code-200 got %d", reqRecorder.Code)
	}
}

func TestListHandler(t *testing.T){

	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	req = req.WithContext(context.Background())
	reqRecorder := httptest.NewRecorder()
	ctx,_:=context.WithCancel(context.Background())
	StartActor(ctx)
	defer func() {
		close(Requests)
		<-Done
	}()
	ListHandler(reqRecorder, req)
	if reqRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code-200 got %d", reqRecorder.Code)
	}
}


func TestDeleteHandler(t *testing.T) {
	// tempFile, _ := os.Open(store.Filename)
	// data := map[int]store.ToDoItem{}
	// data[1] = store.ToDoItem{Task: "", Status: "not started"}

	// encoder := json.NewEncoder(tempFile)
	// encoder.Encode(data)
	// Setup(t)

	req := httptest.NewRequest(http.MethodDelete, "/delete?id=1", nil)
	req = req.WithContext(context.Background())
	// Start the actor first
	

	reqRecorder := httptest.NewRecorder()
	
	ctx,_ :=context.WithCancel(context.Background())
	
	StartActor(ctx)
	defer func() {
		close(Requests)
		<-Done
	}()
	DeleteTask(reqRecorder, req)
	if reqRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code-200 got %d", reqRecorder.Code)
	}
}

