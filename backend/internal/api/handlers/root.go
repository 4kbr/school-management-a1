package handlers

import (
	"fmt"
	"net/http"
)

// RootHandler godoc
// @Summary      Root endpoint
// @Description  Pesan selamat datang di API
// @Tags         root
// @Success      200 {string} string "hello root route"
// @Router       / [get]
func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello root route"))
	fmt.Println("hello root route")
}
