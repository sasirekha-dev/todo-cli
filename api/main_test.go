package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sasirekha-dev/go2.0/store"
)

func TestAddHandler(t *testing.T) {

	tempFile, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		t.Errorf("Error creating temp file")
	}
	defer os.Remove(tempFile.Name())
	store.Filename = tempFile.Name()
	
	newTask := store.ToDoItem{Task: "handler func", Status: "started"}
	body, _:= json.Marshal(newTask)
	req:=httptest.NewRequest(http.MethodPost, "/add", bytes.NewBuffer(body))

	req=req.WithContext(context.Background())
	req.Header.Set("Content-Type", "application/json")

	reqRecorder:=httptest.NewRecorder()
	
	addTask(reqRecorder, req)
	startActor()

	fmt.Print(reqRecorder)
}