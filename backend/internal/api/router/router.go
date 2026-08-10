package router

import (
	"net/http"

	_ "school-management-api/docs"
	"school-management-api/internal/api/handlers"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func Router() *http.ServeMux {
	// todo: saat ini memang sengaja pakai cara router versi go yang lama, nanti diganti ke versi terbaru
	// mux = "resepsionis" server: request masuk diteruskan ke handler sesuai path
	// TODO: refactor ke method routing: mux.HandleFunc("GET /teachers", ...)
	//       biar gak perlu switch r.Method manual di dalam handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.RootHandler)

	// Swagger UI (dokumentasi OpenAPI). Akses: /swagger/index.html
	// docs/ harus di-generate dulu via `make swagger` (swag init).
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	mux.HandleFunc("GET /teachers", handlers.GetTeachersHandler)
	mux.HandleFunc("POST /teachers", handlers.AddTeacherHandler)
	mux.HandleFunc("PATCH /teachers", handlers.PatchTeachersHandler)
	mux.HandleFunc("DELETE /teachers", handlers.DeleteTeachersHandler)

	mux.HandleFunc("GET /teachers/{id}", handlers.GetOneTeacherByIdHandler)
	mux.HandleFunc("PUT /teachers/{id}", handlers.UpdateOneTeacherByIdHandler)
	mux.HandleFunc("PATCH /teachers/{id}", handlers.PatchOneTeacherByIdHandler)
	mux.HandleFunc("DELETE /teachers/{id}", handlers.DeleteOneTeacherByIdHandler)

	mux.HandleFunc("/students/", handlers.StudentsHandler)
	mux.HandleFunc("/execs/", handlers.ExecsHandler)

	return mux
}
