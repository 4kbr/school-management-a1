package handlers

import (
	"fmt"
	"net/http"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello root route"))
	fmt.Println("hello root route")
}
