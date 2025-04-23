package store_test

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/sasirekha-dev/go2.0/models"
	"github.com/sasirekha-dev/go2.0/store"
)

func TestAddTask(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test_*.json")
		if err != nil {
			t.Errorf("Error creating temp file")
		}
		defer os.Remove(tempFile.Name())
	store.Filename = tempFile.Name()
	t.Run("test with valid input", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), models.TraceID, "123")
		//setup
		store.ToDoItems = map[int]store.ToDoItem{}

		//when
		store.AddTask("task1", "pending", ctx)

		//assert
		item, ok := store.ToDoItems[1]
		if !ok {
			t.Errorf("expected key 1 does not exists")
		}
		if item.Task != "task1" {
			t.Errorf("expected task does not exists")
		}
		if item.Status != "pending" {
			t.Errorf("expected status does not exists")
		}
	})
	t.Run("test with empty inputs", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), models.TraceID, "test-trace")
		//setup
		store.ToDoItems = map[int]store.ToDoItem{}

		//when
		store.AddTask("", "", ctx)

		//assert
		if len(store.ToDoItems) != 0 {
			t.Errorf("expected key 1 does not exists")
		}
	})
}

func TestDeleteTask(t *testing.T) {
	t.Run("test with valid input", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), models.TraceID, "123")
		//setup
		store.ToDoItems = map[int]store.ToDoItem{
			1: {Task: "Existing Task", Status: "done"},
		}

		//when
		got := store.DeleteTask(1, store.ToDoItems, ctx)

		//assert
		if got != nil {
			t.Errorf("Delete action failed")
		}
	})

	t.Run("test with out of boundary index value", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), models.TraceID, "123")
		//setup
		store.ToDoItems = map[int]store.ToDoItem{
			1: {Task: "Existing Task", Status: "done"},
		}

		//when
		got := store.DeleteTask(2, store.ToDoItems, ctx)

		//assert
		if got.Error() != "Out of limit index" {
			t.Errorf("Delete action failed")
		}
	})
}

func TestUpdateTask(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test_*.json")
		if err != nil {
			t.Errorf("Error creating temp file")
		}
		defer os.Remove(tempFile.Name())
	store.Filename = tempFile.Name()
	t.Run("test with valid input", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), models.TraceID, "123")
		//setup
		store.ToDoItems = map[int]store.ToDoItem{
			1: {
				Task:   "task2",
				Status: "pending",
			},
		}

		//when
		store.UpdateTask("task1", "pending", 1, ctx)

		//assert
		item, ok := store.ToDoItems[1]
		if !ok {
			t.Errorf("expected key 1 does not exists")
		}
		if item.Task != "task1" {
			t.Errorf("expected task does not exists")
		}
		if item.Status != "pending" {
			t.Errorf("expected status does not exists")
		}
	})
	t.Run("test with empty inputs", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), models.TraceID, "test-trace")
		//setup
		store.ToDoItems = map[int]store.ToDoItem{}

		//when
		got := store.UpdateTask("", "pending", 1, ctx)

		//assert
		if got.Error() != "Out of range" {
			t.Errorf("Update failed")
		}

	})
	t.Run("test when task is empty", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), models.TraceID, "123")
		//setup
		store.ToDoItems = map[int]store.ToDoItem{
			1: {
				Task:   "task2",
				Status: "pending",
			},
		}

		//when
		store.UpdateTask("", "completed", 1, ctx)

		//assert
		item, ok := store.ToDoItems[1]
		if !ok || item.Task != "task2" || item.Status != "completed" {
			t.Errorf("Update test case failed")
		}
	})
}

func TestLoadFile(t *testing.T) {
	ctx := context.WithValue(context.Background(), models.TraceID, "123")
	t.Run("Load File", func(t *testing.T) {
		//setup
		tempFile, err := os.CreateTemp("", "test_*.json")
		if err != nil {
			t.Errorf("Error creating temp file")
		}
		defer os.Remove(tempFile.Name())
		encoder := json.NewEncoder(tempFile)
		if err := encoder.Encode(map[int]store.ToDoItem{
			1: {Task: "abc", Status: "pending"},
		}); err != nil {
			t.Errorf("Cannot write to temp file")
		}
		//when
		data, err := store.Read(tempFile.Name(), ctx)

		//assert
		if !reflect.DeepEqual(data, map[int]store.ToDoItem{1: {Task: "abc", Status: "pending"}}) || err != nil {
			t.Errorf("Read test case failed")
		}

	})
}

func TestSaveFile(t *testing.T) {
	ctx := context.WithValue(context.Background(), models.TraceID, "123")
	t.Run("Save to file", func(t *testing.T) {
		//setup
		tempFile, err := os.CreateTemp("", "test_*.json")
		if err != nil {
			t.Errorf("Error creating temp file")
		}
		defer os.Remove(tempFile.Name())
		store.Filename = tempFile.Name()

		//when
		newData := map[int]store.ToDoItem{
			2: {Task: "Experiment GoLang", Status: "pending"},
		}
		
		err = store.Save(newData, ctx)
		//assert
		if err != nil {
			t.Errorf("Save test case failed")
		}

		content := map[int]store.ToDoItem{}
		decoder := json.NewDecoder(tempFile)
		_ = decoder.Decode(&content)
		if !reflect.DeepEqual(content, newData){
			t.Errorf("Save values are different")
		}
	})
}
