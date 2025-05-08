package store_test

import (
	"context"
	"encoding/json"
	"io"
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
		store.ToDoItems = map[string]map[int]store.ToDoItem{}

		//when
		store.Add("task1", "pending", "test", ctx)

		//assert
		decoder := json.NewDecoder(tempFile)
		var data map[string]map[int]store.ToDoItem
		decoder.Decode(&data)
		if data["test"][1].Task != "task1" && data["test"][1].Status != "pending" {
			t.Errorf("expected data does not exists")
		}

	})
	t.Run("test with empty inputs", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), models.TraceID, "test-trace")
		//setup
		store.ToDoItems = map[string]map[int]store.ToDoItem{}

		//when
		store.Add("", "", "test",ctx)

		//assert
		decoder := json.NewDecoder(tempFile)
		var data map[int]store.ToDoItem
		decoder.Decode(&data)
		if len(data) != 0 {
			t.Errorf("Empty task is added")
		}
	})
}

func TestDeleteTask(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test_*.json")
	if err != nil {
		t.Errorf("Error creating temp file")
	}
	defer os.Remove(tempFile.Name())
	store.Filename = tempFile.Name()
	store.ToDoItems = map[string]map[int]store.ToDoItem{
		"test": {1: {Task: "Existing Task", Status: "done"}},
	}
	t.Run("test with valid input", func(t *testing.T) {
		//setup
		ctx := context.WithValue(context.Background(), models.TraceID, "123")

		encoder := json.NewEncoder(tempFile)
		encoder.Encode(store.ToDoItems)
		//when
		got := store.DeleteTask("test", 1, ctx)

		//assert
		if got != nil {
			t.Errorf("Delete action failed")
		}
	})

	t.Run("test with out of boundary index value", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), models.TraceID, "123")
		//setup

		//when
		got := store.DeleteTask("test",2, ctx)

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
		store.ToDoItems = map[string]map[int]store.ToDoItem{
			"test":{1: {Task: "Existing Task", Status: "done"}},
		}
		encoder := json.NewEncoder(tempFile)
		encoder.Encode(store.ToDoItems)
		//when
		store.Update("test","task1", "pending", 1, ctx)

		//assert
		tempFile.Seek(0, io.SeekStart)
		decoder := json.NewDecoder(tempFile)
		var data map[string]map[int]store.ToDoItem
		decoder.Decode(&data)

		if data["test"][1].Status != "pending" {
			t.Errorf("Update failed-%v", data)
		}
	})
	t.Run("test with empty inputs", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), models.TraceID, "test-trace")
		//setup
		// store.ToDoItems = map[int]store.ToDoItem{}
		// encoder := json.NewEncoder(tempFile)
		// encoder.Encode(store.ToDoItems)

		//when
		got := store.Update("test","", "pending", 2, ctx)

		//assert
		if got.Error() != "Out of range" {
			t.Errorf("Update failed")
		}

	})
	t.Run("test when task is empty", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), models.TraceID, "123")
		//setup
		store.ToDoItems = map[string]map[int]store.ToDoItem{
			"test":{1: {
				Task:   "task2",
				Status: "pending",
			}},
		}
		encoder := json.NewEncoder(tempFile)
		encoder.Encode(store.ToDoItems)
		//when
		store.Update("test","", "completed", 1, ctx)

		//assert
		
		decoder := json.NewDecoder(tempFile)
		tempFile.Seek(0, io.SeekStart)
		var data map[string]map[int]store.ToDoItem
		decoder.Decode(&data)

		if data["test"][1].Task != "task2" && data["test"][1].Status != "completed" {
			t.Errorf("Update test case failed-%v", data)
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
		store.Filename = tempFile.Name()
		encoder := json.NewEncoder(tempFile)
		store.ToDoItems = map[string]map[int]store.ToDoItem{"test":{1: {Task: "abc", Status: "pending"}}}
		encoder.Encode(store.ToDoItems)
		//when
		data, err := store.Read("test",ctx)

		//assert
		if !reflect.DeepEqual(data, map[int]store.ToDoItem{1: {Task: "abc", Status: "pending"}}) || err != nil {
			t.Errorf("Read test case failed-%v", data)
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

		err = store.Save("test",newData, ctx)
		//assert
		if err != nil {
			t.Errorf("Save test case failed")
		}

		content := map[string]map[int]store.ToDoItem{}
		decoder := json.NewDecoder(tempFile)
		_ = decoder.Decode(&content)
		if !reflect.DeepEqual(content["test"], newData) {
			t.Errorf("Save values are different")
		}
	})
}
