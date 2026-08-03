package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	mw "school-management-api/internal/api/middlewares"
	"strconv"
	"strings"
	"sync"
)

type Teacher struct {
	ID        int    `json:"id,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Class     string `json:"class,omitempty"`
	Subject   string `json:"subject,omitempty"`
}

var (
	teachers = make(map[int]Teacher)
	mutex    = &sync.Mutex{}
	nextID   = 1
)

func init() {
	teachers[nextID] = Teacher{ID: nextID, FirstName: "John", LastName: "Doe", Class: "1A", Subject: "Math"}
	nextID++
	teachers[nextID] = Teacher{ID: nextID, FirstName: "Jane", LastName: "Doofy", Class: "2B", Subject: "Science"}
	nextID++
	teachers[nextID] = Teacher{ID: nextID, FirstName: "Jane", LastName: "Em", Class: "3B", Subject: "English"}
	nextID++
}

// handler teacher
func getTeachersHandler(w http.ResponseWriter, r *http.Request) {

	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr := strings.TrimPrefix(path, "/")
	fmt.Println("idStr:", idStr)

	w.Header().Set("Content-Type", "application/json")
	if idStr == "" {

		firstName := r.URL.Query().Get("first_name")
		lastName := r.URL.Query().Get("last_name")

		teacherList := make([]Teacher, 0, len(teachers))
		for _, teacher := range teachers {
			if (firstName == "" || teacher.FirstName == firstName) && (lastName == "" || teacher.LastName == lastName) {
				teacherList = append(teacherList, teacher)
			}
		}

		response := struct {
			Status string    `json:"status"`
			Count  int       `json:"count"`
			Data   []Teacher `json:"data"`
		}{
			Status: "success",
			Count:  len(teacherList),
			Data:   teacherList,
		}

		json.NewEncoder(w).Encode(response)
	}

	//handle path parameter
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	teacher, exists := teachers[id]
	if !exists {
		http.Error(w, "teacher not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(teacher)
}
func addTeacherHandler(w http.ResponseWriter, r *http.Request) {
	// todo: ini untuk apa?
	mutex.Lock()
	defer mutex.Unlock()

	var newTeachers []Teacher
	err := json.NewDecoder(r.Body).Decode(&newTeachers)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	addedTeachers := make([]Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		newTeacher.ID = nextID
		teachers[nextID] = newTeacher
		addedTeachers[i] = newTeacher
		nextID++
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string    `json:"status"`
		Count  int       `json:"count"`
		Data   []Teacher `json:"data"`
	}{
		Status: "success",
		Count:  len(addedTeachers),
		Data:   addedTeachers,
	}
	json.NewEncoder(w).Encode(response)

}

// handler teacher - END

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello root route"))
	fmt.Println("hello root route")
}
func teachersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTeachersHandler(w, r)
		fmt.Println("hello GET Method on teachers route")
	case http.MethodPost:
		addTeacherHandler(w, r)
		fmt.Println("hello POST Method on teachers route")
	case http.MethodPut:
		w.Write([]byte("hello PUT Method on teachers route"))
		fmt.Println("hello PUT Method on teachers route")
	case http.MethodPatch:
		w.Write([]byte("hello PATCH Method on teachers route"))
		fmt.Println("hello PATCH Method on teachers route")
	case http.MethodDelete:
		w.Write([]byte("hello DELETE Method on teachers route"))
		fmt.Println("hello DELETE Method on teachers route")
	}
}
func studentsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("hello GET Method on students route"))
		fmt.Println("hello GET Method on students route")
	case http.MethodPost:
		w.Write([]byte("hello POST Method on students route"))
		fmt.Println("hello POST Method on students route")
	case http.MethodPut:
		w.Write([]byte("hello PUT Method on students route"))
		fmt.Println("hello PUT Method on students route")
	case http.MethodPatch:
		w.Write([]byte("hello PATCH Method on students route"))
		fmt.Println("hello PATCH Method on students route")
	case http.MethodDelete:
		w.Write([]byte("hello DELETE Method on students route"))
		fmt.Println("hello DELETE Method on students route")
	}
}

func execsHandler(w http.ResponseWriter, r *http.Request) {
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

func main() {
	// todo: pindahkan ke .env/config: port, cert path, key path
	port := ":3000"

	// mux = "resepsionis" server: request masuk diteruskan ke handler sesuai path
	// TODO: refactor ke method routing: mux.HandleFunc("GET /teachers", ...)
	//       biar gak perlu switch r.Method manual di dalam handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/teachers/", teachersHandler)
	mux.HandleFunc("/students/", studentsHandler)
	mux.HandleFunc("/execs/", execsHandler)

	// TLS: MinVersion TLS 1.2 — versi lama (SSLv3, TLS 1.0/1.1) sudah insecure
	// cert.pem & key.pem: self-signed, cuma untuk development.
	// Cara generate: lihat backend/docs/command.md
	cert := "cert.pem"
	key := "key.pem"
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// // init rate limiter
	// rl := mw.NewRateLimiter(10, time.Minute)

	// // setup hpp
	// hppOptions := mw.HPPOptions{
	// 	CheckQuery:                  true,
	// 	CheckBody:                   true,
	// 	CheckBodyOnlyForContentType: "application/json",
	// 	Whitelist:                   []string{"allowedParam"},
	// }

	// secureMux :=
	// 	mw.Cors(
	// 		mw.SecurityHeaders(
	// 			rl.Middleware(
	// 				mw.ResponseTime(
	// 					mw.Hpp(hppOptions)(
	// 						mw.Compression(
	// 							mux,
	// 						),
	// 					),
	// 				),
	// 			),
	// 		),
	// 	)
	secureMux := ApplyMiddlewares(mux, mw.SecurityHeaders)
	// secureMux := ApplyMiddlewares(
	// 	mux, mw.Hpp(hppOptions), mw.Compression, mw.SecurityHeaders, mw.ResponseTime, rl.Middleware, mw.Cors,
	// )
	// create custom server
	server := &http.Server{
		Addr:    port,
		Handler: secureMux,
		// Handler: mw.ResponseTime(mw.SecurityHeaders(mw.Cors(mux))),
		// Handler:   mw.Cors(mux),
		TLSConfig: tlsConfig,
	}

	fmt.Println("server is running on port:", port)
	err := server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("error starting the server", err)
	}
}

// Middleware is a function that wraps a httpHandler with additional functionallity
type Middleware func(http.Handler) http.Handler

func ApplyMiddlewares(handler http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}
