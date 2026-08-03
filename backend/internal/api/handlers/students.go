package handlers

import (
	"fmt"
	"net/http"
)

func StudentsHandler(w http.ResponseWriter, r *http.Request) {
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
	}
}
