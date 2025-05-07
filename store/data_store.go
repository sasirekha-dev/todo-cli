package store

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"


)

var Filename = filepath.Join("database", "list.json")

type ToDoItem struct {
	Task   string `json:"task"`
	Status string `json:"status"`
}

var ToDoItems map[string]map[int]ToDoItem

type errorMsg string

func (t ToDoItem) LogValue() slog.Value {
	return slog.StringValue(fmt.Sprintf("Task-%s with status-%s", t.Task, t.Status))
}

func Save(userid string, data map[int]ToDoItem, ctx context.Context) error {
	content := LoadContent()
	content[userid]=data
	bytes, err:= json.Marshal(content)
	if err!=nil{
		slog.ErrorContext(ctx, "Error with json data")
		return nil
	}
	err = os.WriteFile(Filename, bytes, os.FileMode(os.O_RDWR))
	if err!=nil{
		slog.ErrorContext(ctx, "Error writing to file")
		return nil
	}
	slog.InfoContext(ctx, "Saving to file", "Filename", Filename)
	return nil
}

func LoadContent()map[string]map[int]ToDoItem{
	var content map[string]map[int]ToDoItem
	file, e := os.Open(Filename)
	if e!=nil{
		log.Fatalf("Failed to read from file: %v", e)
	}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&content); err != nil {
		if err.Error() == "EOF" {
			log.Println("EOF reached...")
			return make(map[string]map[int]ToDoItem)
		}
		log.Fatalf("Failed to read from file: %v", err)
	}
	return content
}

func Load(userid string)  (map[int]ToDoItem){
	var content map[string]map[int]ToDoItem
	log.Print(Filename)
	file, e := os.Open(Filename)
	if e!=nil{
		log.Fatalf("Failed to read from file: %v", e)
	}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&content); err != nil {
		if err.Error() == "EOF" {
			log.Println("EOF reached...")
			return make(map[int]ToDoItem)
		}
		log.Fatalf("Failed to read from file: %v", err)
	}
	data, found:= content[userid]
	if !found{
		return make(map[int]ToDoItem)
	}
	return data
}

func Read(userid string, ctx context.Context) (map[int]ToDoItem, error) {

	data := Load(userid)
	// file, e := os.Open(Filename)
	// if e!=nil{
	// 	log.Fatalf("Failed to read from file: %v", e)
	// 	return nil, e
	// }
	// decoder := json.NewDecoder(file)
	// if err := decoder.Decode(&content); err != nil {
	// 	if err.Error() == "EOF" {
	// 		log.Println("EOF reached...")
	// 		return make(map[int]ToDoItem), nil
	// 	}
	// 	log.Fatalf("Failed to read from file: %v", err)
	// 	return nil, err
	// }

	slog.InfoContext(ctx, "List all Tasks")
	return data, nil
}

func Add(insertData string, status string, userid string, ctx context.Context) error{
	//get data related to user
	ToDoItems:=Load(userid)



	maxKey:=0
	for id:= range ToDoItems{
		if id>maxKey{
			maxKey = id
		}
	}
	if insertData != "" && status != "" {
		newToDoItem := ToDoItem{insertData, status}
		ToDoItems[maxKey+1] = newToDoItem

		err := Save(userid,ToDoItems, ctx)
		if err!=nil{
			return err
		}
		slog.InfoContext(ctx, "Add Task", "task", newToDoItem)
	}
	return nil
}

func (error_msg errorMsg) Error() string {
	return string(error_msg)
}

func DeleteTask(userid string, taskNumber int, ctx context.Context) error {
	file_content := Load(userid)

	if taskNumber > 0 {
		_, key_present := file_content[taskNumber]
		if key_present {
			del_task := file_content[taskNumber]
			delete(file_content, taskNumber)
			Save(userid, file_content, ctx)		
			slog.InfoContext(ctx, "Delete Task", "task", del_task.Task, "status", del_task.Status)
			log.Print(ctx)
		} else {
			slog.InfoContext(ctx, "Delete Task", "Message:", "Task is not present")
			return errorMsg("Out of limit index")
		}
	}
	return nil
}


func Update(userid string, task string, status string, index int, ctx context.Context) error {
	ToDoItems:= Load(userid)
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
		Save(userid, ToDoItems, ctx)
		slog.InfoContext(ctx, "Update Task", "task", ToDoItem{task, status})
	}
	return nil
}

func Close(file *os.File) error {
	log.Println("CLOSING FILE...")
	if file != nil {
		err := file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func Open() *os.File {
	file, err := os.Open(Filename)
	if err != nil {
		log.Fatal("error creating file")
		// defer file.Close()
		return nil
	}
	// defer file.Close()
	log.Println("File is opened")
	return file
}
