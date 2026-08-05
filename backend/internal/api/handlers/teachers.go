package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"school-management-api/internal/models"
	"school-management-api/internal/repositories/sqlconnect"
	"strconv"
	"strings"
	"sync"
)

var (
	teachers = make(map[int]models.Teacher)
	mutex    = &sync.Mutex{}
	nextID   = 1
)

func init() {
	teachers[nextID] = models.Teacher{ID: nextID, FirstName: "John", LastName: "Doe", Class: "1A", Subject: "Math"}
	nextID++
	teachers[nextID] = models.Teacher{ID: nextID, FirstName: "Jane", LastName: "Doofy", Class: "2B", Subject: "Science"}
	nextID++
	teachers[nextID] = models.Teacher{ID: nextID, FirstName: "Jane", LastName: "Em", Class: "3B", Subject: "English"}
	nextID++
}

func TeachersHandler(w http.ResponseWriter, r *http.Request) {
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

// handler teacher
func getTeachersHandler(w http.ResponseWriter, r *http.Request) {

	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr := strings.TrimPrefix(path, "/")
	fmt.Println("idStr:", idStr)

	w.Header().Set("Content-Type", "application/json")
	if idStr == "" {

		firstName := r.URL.Query().Get("first_name")
		lastName := r.URL.Query().Get("last_name")

		teacherList := make([]models.Teacher, 0, len(teachers))
		for _, teacher := range teachers {
			if (firstName == "" || teacher.FirstName == firstName) && (lastName == "" || teacher.LastName == lastName) {
				teacherList = append(teacherList, teacher)
			}
		}

		response := struct {
			Status string           `json:"status"`
			Count  int              `json:"count"`
			Data   []models.Teacher `json:"data"`
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

	db, err := sqlconnect.ConnectDb()
	if err != nil {
		http.Error(w, "error connecting to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var newTeachers []models.Teacher
	err = json.NewDecoder(r.Body).Decode(&newTeachers)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	stmt, err := db.Prepare("INSERT INTO teachers (first_name, last_name, email, class, subject) VALUES (?,?,?,?,?)")
	if err != nil {
		http.Error(w, "error in preparing SQL query", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		res, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.Class, newTeacher.Subject)
		if err != nil {
			http.Error(w, "error inserting data intto database", http.StatusInternalServerError)
			return
		}
		lastId, err := res.LastInsertId()
		if err != nil {
			http.Error(w, "error getting last insert ID", http.StatusInternalServerError)
			return
		}
		newTeacher.ID = int(lastId)
		addedTeachers[i] = newTeacher
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Teacher `json:"data"`
	}{
		Status: "success",
		Count:  len(addedTeachers),
		Data:   addedTeachers,
	}
	json.NewEncoder(w).Encode(response)
}
