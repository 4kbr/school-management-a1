package router

import (
	"net/http"
	"school-management-api/internal/api/handlers"
)

// registerStudentRoutes mendaftarkan semua route resource students.
func registerStudentRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /students", handlers.GetStudentsHandler)
	mux.HandleFunc("POST /students", handlers.AddStudentHandler)
	mux.HandleFunc("PATCH /students", handlers.PatchStudentsHandler)
	mux.HandleFunc("DELETE /students", handlers.DeleteStudentsHandler)

	mux.HandleFunc("GET /students/{id}", handlers.GetOneStudentByIdHandler)
	mux.HandleFunc("PUT /students/{id}", handlers.UpdateOneStudentByIdHandler)
	mux.HandleFunc("PATCH /students/{id}", handlers.PatchOneStudentByIdHandler)
	mux.HandleFunc("DELETE /students/{id}", handlers.DeleteOneStudentByIdHandler)
}
