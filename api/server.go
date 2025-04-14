package main

import (
	"fmt"
	"net/http"
)

func hello(w http.ResponseWriter, r *http.Request){

	fmt.Fprintf(w, "hello\n %s", r.URL.Path[1:])
}


func main(){
	fmt.Println("Server starting...")

	http.HandleFunc("/home", hello)
	fs := http.FileServer(http.Dir("about"))
	http.Handle("/about/", http.StripPrefix("/about/", fs))
	http.ListenAndServe(":8080", nil)
}