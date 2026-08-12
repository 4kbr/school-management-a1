package router

import (
	"net/http"
	"school-management-api/internal/api/handlers"
)

// registerTeacherRoutes mendaftarkan semua route resource teachers.
// Path ditulis lengkap karena ServeMux Go tidak men-strip prefix otomatis.
func registerTeacherRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /teachers", handlers.GetTeachersHandler)
	mux.HandleFunc("POST /teachers", handlers.AddTeacherHandler)
	mux.HandleFunc("PATCH /teachers", handlers.PatchTeachersHandler)
	mux.HandleFunc("DELETE /teachers", handlers.DeleteTeachersHandler)

	mux.HandleFunc("GET /teachers/{id}", handlers.GetOneTeacherByIdHandler)
	mux.HandleFunc("PUT /teachers/{id}", handlers.UpdateOneTeacherByIdHandler)
	mux.HandleFunc("PATCH /teachers/{id}", handlers.PatchOneTeacherByIdHandler)
	mux.HandleFunc("DELETE /teachers/{id}", handlers.DeleteOneTeacherByIdHandler)

	mux.HandleFunc("GET /teachers/{id}/students", handlers.GetTeacherStudentsHandler)
	mux.HandleFunc("GET /teachers/{id}/studentcount", handlers.GetTeacherStudentCountHandler)
}
