package store

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"

	"go2.0/models"
)

var Filename = "list.json"

type ToDoItem struct {
	Task   string `json:"task"`
	Status string `json:"status"`
}

var ToDoItems map[int]ToDoItem

type errorMsg string

func (t ToDoItem) LogValue() slog.Value {
	return slog.StringValue(fmt.Sprintf("Task-%s with status-%s", t.Task, t.Status))
}

func Save(data map[int]ToDoItem) error{
	file, err := os.Create(Filename)
	if err != nil {
		log.Fatal("failed to create file")
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(&data); err != nil {
		log.Fatalf("Failed to write to file: %v", err)
		return errorMsg("Not able to save")
	}
	return nil
}

func Read(Filename string, ctx context.Context) (map[int]ToDoItem, error) {
	var data map[int]ToDoItem
	file, err := os.Open(Filename)
	if err != nil {
		file, err := os.Create(Filename)
		if err != nil {
			log.Fatal("error creating file")
		}
		defer file.Close()
		return make(map[int]ToDoItem), nil
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		if err.Error() == "EOF" {
			return make(map[int]ToDoItem), nil
		}
		log.Fatalf("Failed to read from file: %v", err)
	}
	slog.Info("List all Tasks", "traceID", ctx.Value(models.TraceID))
	return data, err
}

func AddTask(insertData string, status string, ctx context.Context) {
	// get length of the list
	totalItems := len(ToDoItems)
	if insertData != "" && status != "" {
		newToDoItem := ToDoItem{insertData, status}
		ToDoItems[totalItems+1] = newToDoItem
		Save(ToDoItems)
		slog.Info("Add Task", "task", newToDoItem, "traceID", ctx.Value(models.TraceID))
	}
}

func (error_msg errorMsg) Error() string {
	return string(error_msg)
}

func DeleteTask(taskNumber int, file_content map[int]ToDoItem, ctx context.Context) error {
	if taskNumber > 0 {
		_, key_present := file_content[taskNumber]
		if key_present {
			del_task := file_content[taskNumber]
			delete(file_content, taskNumber)
			Save(file_content)
			slog.Info("Delete Task", "task", del_task, "traceID", ctx.Value(models.TraceID))
		} else {
			slog.Info("Delete Task", "Message:", "Task is not present", "traceID", ctx.Value(models.TraceID))
			return errorMsg("Out of limit index")
		}		
	}
	return nil
}

func UpdateTask(task string, status string, index int, ctx context.Context) error {
	if index > 0 {
		update_item, exists := ToDoItems[index]
		if exists {
			if task == "" {
				update_item.Status = status
			} else if status == "" {
				update_item.Task = task
			} else {
				update_item = ToDoItem{Task: task, Status: status}
			}
		} else{
			return errorMsg("Out of range")
		}
		ToDoItems[index] = update_item
		Save(ToDoItems)
		slog.Info("Update Task", "task", ToDoItem{task, status}, "traceID", ctx.Value(models.TraceID))
		
	}
	return nil
}
