package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"school-management-api/internal/models"
	"school-management-api/internal/repositories/sqlconnect"
	"strconv"
)

// handler teacher

// GET /teachers/
func GetTeachersHandler(w http.ResponseWriter, r *http.Request) {
	// seluruh proses (build query filter/sort + connect + query + scan) di repository
	teacherList, err := sqlconnect.GetTeachers(r.URL.Query())
	if err != nil {
		fmt.Println(err)
		http.Error(w, "database query error", http.StatusInternalServerError)
		return
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GET /teachers/{id}
func GetOneTeacherByIdHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	fmt.Println("idStr:", idStr)

	w.Header().Set("Content-Type", "application/json")

	// handle path parameter
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	// delegasikan seluruh proses DB ke repository
	teacher, err := sqlconnect.GetTeacherByID(id)
	if errors.Is(err, sqlconnect.ErrTeacherNotFound) {
		http.Error(w, "teacher not found", http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println(err)
		http.Error(w, "database query error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(teacher)
}

// POST /teachers/
func AddTeacherHandler(w http.ResponseWriter, r *http.Request) {
	var newTeachers []models.Teacher
	err := json.NewDecoder(r.Body).Decode(&newTeachers)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// delegasikan seluruh proses insert ke repository
	addedTeachers, err := sqlconnect.CreateTeachers(newTeachers)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "error inserting data into database", http.StatusInternalServerError)
		return
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

// PUT /teachers/{id}
func UpdateOneTeacherByIdHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid teacher id", http.StatusBadRequest)
		return
	}

	var updatedTeacher models.Teacher
	err = json.NewDecoder(r.Body).Decode(&updatedTeacher)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	// cek dulu apakah teacher dengan id tersebut ada (404 kalau tidak)
	existingTeacher, err := sqlconnect.GetTeacherByID(id)
	if err != nil {
		if errors.Is(err, sqlconnect.ErrTeacherNotFound) {
			http.Error(w, "teacher not found", http.StatusNotFound)
			return
		}
		log.Println(err)
		http.Error(w, "unable to retrieve database", http.StatusInternalServerError)
		return
	}

	// pertahankan ID asli dari DB, update field lain dari body
	updatedTeacher.ID = existingTeacher.ID
	err = sqlconnect.UpdateTeacher(updatedTeacher)
	if err != nil {
		log.Println(err)
		http.Error(w, "error updating teacher", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTeacher)
}

// PATCH /teachers/ — batch update banyak teacher dalam SATU transaksi.
// Body: [{"id":100,"first_name":"X"}, {"id":104,"class":"9-Z"}, ...]
// Kalau salah satu gagal, SEMUA di-rollback (all-or-nothing).
func PatchTeachersHandler(w http.ResponseWriter, r *http.Request) {
	var updates []map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	// seluruh transaksi (connect + begin + loop + commit/rollback) di repository
	updatedTeachers, err := sqlconnect.PatchTeachers(updates)
	if err != nil {
		switch {
		case errors.Is(err, sqlconnect.ErrTeacherNotFound):
			http.Error(w, "teacher not found", http.StatusNotFound)
		case errors.Is(err, sqlconnect.ErrInvalidTeacherID):
			http.Error(w, "invalid teacher id in update", http.StatusBadRequest)
		default:
			log.Println(err)
			http.Error(w, "error updating teacher", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Teacher `json:"data"`
	}{
		Status: "success",
		Count:  len(updatedTeachers),
		Data:   updatedTeachers,
	}
	json.NewEncoder(w).Encode(response)
}

// PATCH /teachers/{id}
func PatchOneTeacherByIdHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid teacher id", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	// get + apply + update dilakukan sekaligus di repository
	teacher, err := sqlconnect.PatchTeacherByID(id, updates)
	if err != nil {
		if errors.Is(err, sqlconnect.ErrTeacherNotFound) {
			http.Error(w, "teacher not found", http.StatusNotFound)
			return
		}
		log.Println(err)
		http.Error(w, "error updating teacher", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teacher)
}

// DELETE /teachers/ — batch hapus banyak teacher dalam SATU transaksi.
// Body: [108, 109, 110] — array id yang mau dihapus.
// All-or-nothing: kalau satu id gak ada, SEMUA batal (rollback).
func DeleteTeachersHandler(w http.ResponseWriter, r *http.Request) {
	// baca body: array id
	var ids []int
	err := json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	// seluruh transaksi delete di-handle repository
	deleted, err := sqlconnect.DeleteTeachers(ids)
	if err != nil {
		if errors.Is(err, sqlconnect.ErrTeacherNotFound) {
			http.Error(w, "teacher not found", http.StatusNotFound)
			return
		}
		log.Println(err)
		http.Error(w, "error deleting result", http.StatusInternalServerError)
		return
	}

	// response sukses + list id yang terhapus
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
		IDs    []int  `json:"ids"`
	}{
		Status: "success",
		Count:  len(deleted),
		IDs:    deleted,
	}
	json.NewEncoder(w).Encode(response)
}

// DELETE /teachers/{id}
func DeleteOneTeacherByIdHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid teacher id", http.StatusBadRequest)
		return
	}

	rowsEffected, err := sqlconnect.DeleteTeacherByID(id)
	if err != nil {
		log.Println(err)
		http.Error(w, "error deleting result", http.StatusInternalServerError)
		return
	}
	if rowsEffected == 0 {
		http.Error(w, "teacher not found", http.StatusNotFound)
		return
	}

	// response body
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string `json:"status"`
		ID     int    `json:"id"`
	}{
		Status: "success",
		ID:     id,
	}
	json.NewEncoder(w).Encode(response)
}
