package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// todo: nanti pindahkan ke .env
	port := ":3000"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello root route"))
		fmt.Println("hello root route")
	})
	http.HandleFunc("/teachers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Write([]byte("hello GET Method on teachers route"))
			fmt.Println("hello GET Method on teachers route")
		case http.MethodPost:
			w.Write([]byte("hello POST Method on teachers route"))
			fmt.Println("hello POST Method on teachers route")
		case http.MethodPut:
			w.Write([]byte("hello PUT Method on teachers route"))
			fmt.Println("hello PUT Method on teachers route")
		case http.MethodPatch:
			w.Write([]byte("hello PATCH Method on teachers route"))
			fmt.Println("hello PATCH Method on teachers route")
		case http.MethodDelete:
			w.Write([]byte("hello DELETE Method on teachers route"))
			fmt.Println("hello DELETE Method on teachers route")
		}
		w.Write([]byte("hello others Method on teachers route"))
		fmt.Println("hello others Method on teachers route")
		return
	})
	http.HandleFunc("/students", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello students route"))
		fmt.Println("hello students route")
	})
	http.HandleFunc("/execs", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello execs route"))
		fmt.Println("hello execs route")
	})

	fmt.Println("server is running on port:", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalln("error starting the server", err)
	}
}
