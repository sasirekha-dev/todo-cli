**GO Academy**

This project is to create a to-do app which allows users to create a list of to-do tasks. It is intended for you to use
a range of different techniques and features of Golang and tasks followed from top to bottom, building up the
application as you go. Don't be afraid to set aside your design at the end of a step if the next step would require
massive refactor - it is often simpler to design an application from a clean page and copy in useful code.

A to-do item should include a description and a status of "not started", "started", or "completed". Users should be able
to display a list of all to-do items including their statuses, and should be able to update a to-do item to change its
status or delete it entirely. Feel free to add any additional functionality if you want.

1) **Basic CLI application**

- Create a command line application that uses flags to accept a to-do item adds it to an empty list of to-do items and
  prints the list to console
- After printing the list of to-do items, save them to a file on disk
- When the application starts, load all to-do items from disk before adding new item
- Allow the user to update the description of a to-do item or delete it

2) **Advanced functionality**

- Use the "log/slog" structured logging package to log errors and when data is saved to disk
- Use the "context" package to add a TraceID to enable traceability of calls through the solution by adding it to all
  logs
- Separate the core todo store logic into a different package/module to main/CLI code
- Write unit tests to cover usefully testable code
- Use the "os/signal" package and ensure that the application only exits when it receives the interrupt signal (ctrl+c)

3) **API**

- Use ServeMux in the "net/http" package to expose json http endpoints: "/create", "/get", "/update", and "/delete"
- Add a middleware to create a context with TraceID

4) **Web page**

- Use "http.FileServer" to serve a static page to a new "/about" endpoint
- Use "html/template" to serve a dynamic page containing a list of all to-do items to a new "/list" endpoint

5) **Concurrency**

- Use the Actor/Communicating Sequential Processes (CSP) pattern to support concurrent reads and concurrent safe write
- Use Parallel tests to validate that the solution is concurrent safe

6) **Repl**
- When the application runs it should ask the user to input text into the console to create, read, update, or delete list items in a loop.

To send requests in CLI mode, follow below steps
- Add - ```go run cli\main.go --add="task" --status="started"```
- Update- ```go run cli\main.go --update=1 --task="task" --status="completed"```
- Delete- ```go run cli\main.go --delete=1```
**Note** Logger is setup to save the logs in a log file. Hence it is required to set the environment variable 
```set LOG_FILE=app.log```


To run in API mode
- Add - ```http://localhost:8080/add``` with body of request as ```{"task": "foo", "status":"started"}```
- Update- ```http://localhost:8080/update``` with body as ```{"index":1, "task":"foo", "status":"completed"}```
- Delete- ```http://localhost:8080/delete?id=1```
- List - ```http://localhost:8080/list```
This endpoint lists all the tasks from the database as HTML page
- About - ```http:localhost:8080/about/about.html``` Displays a static webpage  

