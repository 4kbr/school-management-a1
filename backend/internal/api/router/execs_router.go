package router

import (
	"net/http"
	"school-management-api/internal/api/handlers"
)

// registerExecRoutes mendaftarkan semua route resource execs.
// Path ditulis lengkap karena ServeMux Go tidak men-strip prefix otomatis.
func registerExecRoutes(mux *http.ServeMux) {

	mux.HandleFunc("GET /execs", handlers.GetExecsHandler)
	mux.HandleFunc("POST /execs", handlers.AddExecsHandler)
	mux.HandleFunc("PATCH /execs", handlers.PatchExecsHandler)

	mux.HandleFunc("GET /execs/{id}", handlers.GetOneExecHandler)
	mux.HandleFunc("PATCH /execs/{id}", handlers.PatchOneExecHandler)
	mux.HandleFunc("DELETE /execs/{id}", handlers.DeleteOneExecHandler)

	// TODO (auth, belum diimplementasi):
	// mux.HandleFunc("POST /execs/{id}/updatepassword", handlers.UpdatePasswordHandler)
	// mux.HandleFunc("POST /execs/login", handlers.LoginHandler)
	// mux.HandleFunc("POST /execs/logout", handlers.LogoutHandler)
	// mux.HandleFunc("POST /execs/forgotpassword", handlers.ForgotPasswordHandler)
	// mux.HandleFunc("POST /execs/resetpassword/reset/{resetcode}", handlers.ResetPasswordHandler)
}
