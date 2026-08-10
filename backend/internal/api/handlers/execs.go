package handlers

import (
	"fmt"
	"net/http"
)

// ExecsHandler godoc
// @Summary      Execs placeholder
// @Description  Endpoint execs (placeholder, belum diimplementasi penuh)
// @Tags         execs
// @Success      200 {string} string "hello execs route"
// @Router       /execs/ [get]
func ExecsHandler(w http.ResponseWriter, r *http.Request) {
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
	}
}
