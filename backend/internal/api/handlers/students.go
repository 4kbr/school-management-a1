package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"school-management-api/internal/models"
	"school-management-api/internal/models/dto"
	"school-management-api/internal/repositories/sqlconnect"
	"school-management-api/pkg/utils"
	"strconv"
)

// handler student

// GetStudentsHandler godoc
// @Summary      List students
// @Description  Mengambil daftar student dengan filter & sorting opsional
// @Tags         students
// @Produce      json
// @Param        first_name  query string false "Filter by first name"
// @Param        last_name   query string false "Filter by last name"
// @Param        email       query string false "Filter by email"
// @Param        class       query string false "Filter by class"
// @Param        sortby      query []string false "Sort: field:asc|desc (repeatable, ex: first_name:asc)"
// @Success      200 {object} dto.StudentListResponse
// @Failure      500 {object} map[string]interface{}
// @Router       /students [get]
func GetStudentsHandler(w http.ResponseWriter, r *http.Request) {
	// seluruh proses (build query filter/sort + connect + query + scan) di repository
	studentList, err := sqlconnect.GetStudents(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Student `json:"data"`
	}{
		Status: "success",
		Count:  len(studentList),
		Data:   studentList,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetOneStudentByIdHandler godoc
// @Summary      Get a student by ID
// @Description  Mengambil satu student berdasarkan id
// @Tags         students
// @Produce      json
// @Param        id   path      int  true  "Student ID"
// @Success      200  {object}  models.Student
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /students/{id} [get]
func GetOneStudentByIdHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	w.Header().Set("Content-Type", "application/json")

	// handle path parameter
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid student id", http.StatusBadRequest)
		return
	}

	// delegasikan seluruh proses DB ke repository
	student, err := sqlconnect.GetStudentByID(id)
	if errors.Is(err, sqlconnect.ErrStudentNotFound) {
		http.Error(w, "student not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(student)
}

// AddStudentHandler godoc
// @Summary      Create students (batch)
// @Description  Membuat satu atau lebih student sekaligus. Class harus sudah ada di tabel teachers.
// @Tags         students
// @Accept       json
// @Produce      json
// @Param        body  body      []dto.CreateStudentRequest  true  "Array of students"
// @Success      201   {object}  dto.StudentListResponse
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /students [post]
func AddStudentHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Decode body ke slice request DTO (deteksi juga tipe mismatch)
	var reqs []dto.CreateStudentRequest
	if verrs := utils.DecodeJSON(r.Body, &reqs); verrs != nil {
		writeValidationError(w, verrs)
		return
	}

	// 2. Validasi tiap item; kalau ada error → 400 JSON terstruktur
	if verrs := utils.ValidateSlice(reqs); verrs != nil {
		writeValidationError(w, verrs)
		return
	}

	// 3. Map DTO → model untuk disimpan
	newStudents := make([]models.Student, len(reqs))
	for i, req := range reqs {
		newStudents[i] = createStudentReqToModel(req)
	}

	// 4. Delegasikan seluruh proses insert ke repository
	addedStudents, err := sqlconnect.CreateStudents(newStudents)
	if err != nil {
		if errors.Is(err, sqlconnect.ErrClassNotFound) {
			http.Error(w, "class not found", http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Student `json:"data"`
	}{
		Status: "success",
		Count:  len(addedStudents),
		Data:   addedStudents,
	}
	json.NewEncoder(w).Encode(response)
}

// UpdateOneStudentByIdHandler godoc
// @Summary      Update a student (full replace)
// @Description  Mengganti seluruh data student berdasarkan id. Class harus sudah ada di tabel teachers.
// @Tags         students
// @Accept       json
// @Produce      json
// @Param        id    path  int                      true  "Student ID"
// @Param        body  body  dto.UpdateStudentRequest true  "Full student data"
// @Success      200   {object}  models.Student
// @Failure      400   {object}  map[string]interface{}
// @Failure      404   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /students/{id} [put]
func UpdateOneStudentByIdHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid student id", http.StatusBadRequest)
		return
	}

	// 1. Decode body ke request DTO (deteksi juga tipe mismatch)
	var req dto.UpdateStudentRequest
	if verrs := utils.DecodeJSON(r.Body, &req); verrs != nil {
		writeValidationError(w, verrs)
		return
	}

	// 2. Validasi (full required); kalau error → 400 JSON terstruktur
	if verrs := utils.ValidateStruct(&req); verrs != nil {
		writeValidationError(w, verrs)
		return
	}

	// 3. Map DTO → model
	updatedStudent := updateStudentReqToModel(req)

	// 4. cek dulu apakah student dengan id tersebut ada (404 kalau tidak)
	existingStudent, err := sqlconnect.GetStudentByID(id)
	if err != nil {
		if errors.Is(err, sqlconnect.ErrStudentNotFound) {
			http.Error(w, "student not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. pertahankan ID asli dari DB, update field lain dari body
	updatedStudent.ID = existingStudent.ID
	err = sqlconnect.UpdateStudent(updatedStudent)
	if err != nil {
		if errors.Is(err, sqlconnect.ErrClassNotFound) {
			http.Error(w, "class not found", http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedStudent)
}

// PatchStudentsHandler godoc
// @Summary      Patch students (batch)
// @Description  Update sebagian field dari banyak student dalam satu transaksi. All-or-nothing. Class harus sudah ada di tabel teachers.
// @Tags         students
// @Accept       json
// @Produce      json
// @Param        body  body  []dto.PatchStudentRequest  true  "Array of partial updates"
// @Success      200   {object}  dto.StudentListResponse
// @Failure      400   {object}  map[string]interface{}
// @Failure      404   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /students [patch]
func PatchStudentsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Decode body ke slice request DTO (deteksi juga tipe mismatch)
	var reqs []dto.PatchStudentRequest
	if verrs := utils.DecodeJSON(r.Body, &reqs); verrs != nil {
		writeValidationError(w, verrs)
		return
	}

	// 2. Validasi tiap item (partial, via pointer+omitempty)
	if verrs := utils.ValidateSlice(reqs); verrs != nil {
		writeValidationError(w, verrs)
		return
	}

	// 3. Map DTO → map updates untuk repository
	updates := make([]map[string]interface{}, len(reqs))
	for i, req := range reqs {
		updates[i] = patchStudentReqToMap(req)
	}

	// 4. seluruh transaksi (connect + begin + loop + commit/rollback) di repository
	updatedStudents, err := sqlconnect.PatchStudents(updates)
	if err != nil {
		switch {
		case errors.Is(err, sqlconnect.ErrStudentNotFound):
			http.Error(w, "student not found", http.StatusNotFound)
		case errors.Is(err, sqlconnect.ErrInvalidStudentID):
			http.Error(w, "invalid student id in update", http.StatusBadRequest)
		case errors.Is(err, sqlconnect.ErrClassNotFound):
			http.Error(w, "class not found", http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Student `json:"data"`
	}{
		Status: "success",
		Count:  len(updatedStudents),
		Data:   updatedStudents,
	}
	json.NewEncoder(w).Encode(response)
}

// PatchOneStudentByIdHandler godoc
// @Summary      Patch a student
// @Description  Update sebagian field satu student berdasarkan id. Class harus sudah ada di tabel teachers.
// @Tags         students
// @Accept       json
// @Produce      json
// @Param        id    path  int                     true  "Student ID"
// @Param        body  body  dto.PatchStudentRequest true  "Partial fields to update"
// @Success      200   {object}  models.Student
// @Failure      400   {object}  map[string]interface{}
// @Failure      404   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /students/{id} [patch]
func PatchOneStudentByIdHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid student id", http.StatusBadRequest)
		return
	}

	// 1. Decode body ke request DTO (deteksi juga tipe mismatch)
	var req dto.PatchStudentRequest
	if verrs := utils.DecodeJSON(r.Body, &req); verrs != nil {
		writeValidationError(w, verrs)
		return
	}

	// 2. Untuk single patch, id berasal dari URL path, bukan body
	req.ID = id

	// 3. Validasi partial; kalau error → 400 JSON terstruktur
	if verrs := utils.ValidateStruct(&req); verrs != nil {
		writeValidationError(w, verrs)
		return
	}

	// 4. Map DTO → map updates untuk repository
	updates := patchStudentReqToMap(req)

	// 5. get + apply + update dilakukan sekaligus di repository
	student, err := sqlconnect.PatchStudentByID(id, updates)
	if err != nil {
		switch {
		case errors.Is(err, sqlconnect.ErrStudentNotFound):
			http.Error(w, "student not found", http.StatusNotFound)
		case errors.Is(err, sqlconnect.ErrClassNotFound):
			http.Error(w, "class not found", http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(student)
}

// DeleteStudentsHandler godoc
// @Summary      Delete students (batch)
// @Description  Hapus banyak student dalam satu transaksi. All-or-nothing.
// @Tags         students
// @Accept       json
// @Produce      json
// @Param        body  body  []int  true  "Array of student IDs"
// @Success      200   {object}  dto.StudentDeleteResponse
// @Failure      400   {object}  map[string]interface{}
// @Failure      404   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /students [delete]
func DeleteStudentsHandler(w http.ResponseWriter, r *http.Request) {
	// baca body: array id
	var ids []int
	err := json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	// seluruh transaksi delete di-handle repository
	deleted, err := sqlconnect.DeleteStudents(ids)
	if err != nil {
		if errors.Is(err, sqlconnect.ErrStudentNotFound) {
			http.Error(w, "student not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

// DeleteOneStudentByIdHandler godoc
// @Summary      Delete a student
// @Description  Hapus satu student berdasarkan id
// @Tags         students
// @Produce      json
// @Param        id  path  int  true  "Student ID"
// @Success      200  {object}  dto.StudentDeleteOneResponse
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /students/{id} [delete]
func DeleteOneStudentByIdHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid student id", http.StatusBadRequest)
		return
	}

	rowsEffected, err := sqlconnect.DeleteStudentByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsEffected == 0 {
		http.Error(w, "student not found", http.StatusNotFound)
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

// createStudentReqToModel memetakan CreateStudentRequest ke models.Student.
func createStudentReqToModel(req dto.CreateStudentRequest) models.Student {
	return models.Student{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Class:     req.Class,
	}
}

// updateStudentReqToModel memetakan UpdateStudentRequest ke models.Student.
func updateStudentReqToModel(req dto.UpdateStudentRequest) models.Student {
	return models.Student{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Class:     req.Class,
	}
}

// patchStudentReqToMap mengubah PatchStudentRequest menjadi map updates.
// Field yang nil (tidak dikirim client) tidak dimasukkan.
func patchStudentReqToMap(req dto.PatchStudentRequest) map[string]interface{} {
	// 1. Siapkan map updates
	updates := map[string]interface{}{
		"id": req.ID,
	}
	// 2. Tambahkan field yang tidak nil (karena itu yang dikirim client)
	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Class != nil {
		updates["class"] = *req.Class
	}
	// 3. Kembalikan map updates
	return updates
}
