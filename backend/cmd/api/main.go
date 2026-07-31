package main

import (
	"fmt"
	"log"
	"net/http"
)

type user struct {
	Name string `json:"name"`
	Age  int16  `json:"age"`
	City string `json:"city"`
}

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
		default:
			w.Write([]byte("hello others Method on teachers route"))
			fmt.Println("hello others Method on teachers route")
		}
	})
	http.HandleFunc("/students", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Write([]byte("hello GET Method on students route"))
			fmt.Println("hello GET Method on students route")
		case http.MethodPost:
			w.Write([]byte("hello POST Method on students route"))
			fmt.Println("hello POST Method on students route")
		case http.MethodPut:
			w.Write([]byte("hello PUT Method on students route"))
			fmt.Println("hello PUT Method on students route")
		case http.MethodPatch:
			w.Write([]byte("hello PATCH Method on students route"))
			fmt.Println("hello PATCH Method on students route")
		case http.MethodDelete:
			w.Write([]byte("hello DELETE Method on students route"))
			fmt.Println("hello DELETE Method on students route")
		default:
			w.Write([]byte("hello others Method on students route"))
			fmt.Println("hello others Method on students route")
		}
	})
	http.HandleFunc("/execs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Write([]byte("hello GET Method on execs route"))
			fmt.Println("hello GET Method on execs route")
		case http.MethodPost:
			w.Write([]byte("hello POST Method on execs route"))
			fmt.Println("hello POST Method on execs route")
		case http.MethodPut:
			w.Write([]byte("hello PUT Method on execs route"))
			fmt.Println("hello PUT Method on execs route")
		case http.MethodPatch:
			w.Write([]byte("hello PATCH Method on execs route"))
			fmt.Println("hello PATCH Method on execs route")
		case http.MethodDelete:
			w.Write([]byte("hello DELETE Method on execs route"))
			fmt.Println("hello DELETE Method on execs route")
		default:
			w.Write([]byte("hello others Method on execs route"))
			fmt.Println("hello others Method on execs route")
		}
	})

	fmt.Println("server is running on port:", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalln("error starting the server", err)
	}
}
