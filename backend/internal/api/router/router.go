package router

import (
	"net/http"
	"school-management-api/internal/api/handlers"
)

func Router() *http.ServeMux {
	// todo: saat ini memang sengaja pakai cara router versi go yang lama, nanti diganti ke versi terbaru
	// mux = "resepsionis" server: request masuk diteruskan ke handler sesuai path
	// TODO: refactor ke method routing: mux.HandleFunc("GET /teachers", ...)
	//       biar gak perlu switch r.Method manual di dalam handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.RootHandler)
	mux.HandleFunc("/teachers/", handlers.TeachersHandler)
	mux.HandleFunc("/students/", handlers.StudentsHandler)
	mux.HandleFunc("/execs/", handlers.ExecsHandler)

	return mux
}
