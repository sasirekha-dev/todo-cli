package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"

	"github.com/sasirekha-dev/go2.0/store"
)

func main() {
	ctx := context.WithoutCancel(context.Background())
	scanner := bufio.NewScanner(os.Stdin)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	for {

		fmt.Println("1. Add Task")
		fmt.Println("2. Update Task")
		fmt.Println("3. Delete Task")
		fmt.Println("4. List Tasks")
		fmt.Println("Enter Choice:")
		scanner.Scan()
		choice := scanner.Text()

		switch choice {
		case "1":
			fmt.Println("Enter the task:")
			scanner.Scan()
			task := scanner.Text()
			fmt.Println("Enter the status: ")
			scanner.Scan()
			status := scanner.Text()
			if err := store.Add(task, status, ctx); err != nil {
				slog.ErrorContext(ctx, err.Error())
			}
		case "2":
			fmt.Println("Enter the task to update:")
			scanner.Scan()
			task := scanner.Text()
			fmt.Println("Enter the status to update: ")
			scanner.Scan()
			status := scanner.Text()
			fmt.Println("Enter the index: ")
			scanner.Scan()
			index, e := strconv.Atoi(scanner.Text())
			if e != nil {
				slog.ErrorContext(ctx, "Index must be a integer value")
			}
			if err := store.Update(task, status, index, ctx); err != nil {
				slog.ErrorContext(ctx, err.Error())
			}
		case "3":
			fmt.Println("Enter the index of task to delete")
			scanner.Scan()
			index, e := strconv.Atoi(scanner.Text())
			if e != nil {
				slog.ErrorContext(ctx, "Index must be a integer value")
			}
			if err := store.DeleteTask(index, ctx); err != nil {
				slog.ErrorContext(ctx, err.Error())
			}
		case "4":
			list, e := store.Read(ctx)
			if e != nil {
				slog.ErrorContext(ctx, e.Error())
			}

			for index, item := range list {
				fmt.Printf("%d. %s - %s", index, item.Task, item.Status)
			}

		default:
			fmt.Println("Not a valid input...")
			select{
			case <-quit:
				slog.InfoContext(ctx, "Quit command received")
				return
			case <-ctx.Done():
				slog.InfoContext(ctx, "context closed")
				return
			default:
				continue
			}
			
		}

	}

}
