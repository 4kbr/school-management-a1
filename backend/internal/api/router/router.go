package router

import (
	"net/http"

	_ "school-management-api/docs"
	"school-management-api/internal/api/handlers"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func MainRouter() *http.ServeMux {
	// Semua route didaftarkan di SATU mux yang sama. ServeMux di Go tidak
	// men-strip prefix seperti Router di Express — jadi sub-resource tetap
	// ditulis path lengkap ("GET /teachers/{id}"), cuma organisasinya
	// dipisah per file lewat fungsi registrar di bawah ini.
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.RootHandler)

	// Swagger UI (dokumentasi OpenAPI). Akses: /swagger/index.html
	// docs/ harus di-generate dulu via `make swagger` (swag init).
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	registerTeacherRoutes(mux)
	registerStudentRoutes(mux)
	registerExecRoutes(mux)

	return mux
}
