package middlewares

import (
	"fmt"
	"net/http"
	"time"
)

func ResponseTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("received request in ResponseTime")
		start := time.Now()

		// create a custom ResponseWriter to capture statuscode
		wrappedWriter := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		//calculate the duration
		duration := time.Since(start)
		w.Header().Set("X-Response-Time", duration.String())
		next.ServeHTTP(wrappedWriter, r)

		//LOG The request details
		duration = time.Since(start)
		fmt.Printf("Method: %s, URL: %s, Status: %d, Duration: %v\n", r.Method, r.URL, wrappedWriter.status, duration.String())
		fmt.Println("sent response from response time middleware")
	})
}

// responseWriter
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
