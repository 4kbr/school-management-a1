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

// handler teacher

// GetTeachersHandler godoc
// @Summary      List teachers
// @Description  Mengambil daftar teacher dengan filter & sorting opsional
// @Tags         teachers
// @Produce      json
// @Param        first_name  query string false "Filter by first name"
// @Param        last_name   query string false "Filter by last name"
// @Param        email       query string false "Filter by email"
// @Param        class       query string false "Filter by class"
// @Param        subject     query string false "Filter by subject"
// @Param        sortby      query []string false "Sort: field:asc|desc (repeatable, ex: first_name:asc)"
// @Success      200 {object} dto.TeacherListResponse
// @Failure      500 {object} map[string]interface{}
// @Router       /teachers [get]
func GetTeachersHandler(w http.ResponseWriter, r *http.Request) {
	// seluruh proses (build query filter/sort + connect + query + scan) di repository
	teacherList, err := sqlconnect.GetTeachers(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

// GetOneTeacherByIdHandler godoc
// @Summary      Get a teacher by ID
// @Description  Mengambil satu teacher berdasarkan id
// @Tags         teachers
// @Produce      json
// @Param        id   path      int  true  "Teacher ID"
// @Success      200  {object}  models.Teacher
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /teachers/{id} [get]
func GetOneTeacherByIdHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	w.Header().Set("Content-Type", "application/json")

	// handle path parameter
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid teacher id", http.StatusBadRequest)
		return
	}

	// delegasikan seluruh proses DB ke repository
	teacher, err := sqlconnect.GetTeacherByID(id)
	if errors.Is(err, sqlconnect.ErrTeacherNotFound) {
		http.Error(w, "teacher not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(teacher)
}

// AddTeacherHandler godoc
// @Summary      Create teachers (batch)
// @Description  Membuat satu atau lebih teacher sekaligus
// @Tags         teachers
// @Accept       json
// @Produce      json
// @Param        body  body      []dto.CreateTeacherRequest  true  "Array of teachers"
// @Success      201   {object}  dto.TeacherListResponse
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /teachers [post]
func AddTeacherHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Decode body ke slice request DTO (deteksi juga tipe mismatch)
	var reqs []dto.CreateTeacherRequest
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
	newTeachers := make([]models.Teacher, len(reqs))
	for i, req := range reqs {
		newTeachers[i] = createReqToModel(req)
	}

	// 4. Delegasikan seluruh proses insert ke repository
	addedTeachers, err := sqlconnect.CreateTeachers(newTeachers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

// UpdateOneTeacherByIdHandler godoc
// @Summary      Update a teacher (full replace)
// @Description  Mengganti seluruh data teacher berdasarkan id
// @Tags         teachers
// @Accept       json
// @Produce      json
// @Param        id    path  int                      true  "Teacher ID"
// @Param        body  body  dto.UpdateTeacherRequest true  "Full teacher data"
// @Success      200   {object}  models.Teacher
// @Failure      400   {object}  map[string]interface{}
// @Failure      404   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /teachers/{id} [put]
func UpdateOneTeacherByIdHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid teacher id", http.StatusBadRequest)
		return
	}

	// 1. Decode body ke request DTO (deteksi juga tipe mismatch)
	var req dto.UpdateTeacherRequest
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
	updatedTeacher := updateReqToModel(req)

	// 4. cek dulu apakah teacher dengan id tersebut ada (404 kalau tidak)
	existingTeacher, err := sqlconnect.GetTeacherByID(id)
	if err != nil {
		if errors.Is(err, sqlconnect.ErrTeacherNotFound) {
			http.Error(w, "teacher not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. pertahankan ID asli dari DB, update field lain dari body
	updatedTeacher.ID = existingTeacher.ID
	err = sqlconnect.UpdateTeacher(updatedTeacher)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTeacher)
}

// PatchTeachersHandler godoc
// @Summary      Patch teachers (batch)
// @Description  Update sebagian field dari banyak teacher dalam satu transaksi. All-or-nothing.
// @Tags         teachers
// @Accept       json
// @Produce      json
// @Param        body  body  []dto.PatchTeacherRequest  true  "Array of partial updates"
// @Success      200   {object}  dto.TeacherListResponse
// @Failure      400   {object}  map[string]interface{}
// @Failure      404   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /teachers [patch]
func PatchTeachersHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Decode body ke slice request DTO (deteksi juga tipe mismatch)
	var reqs []dto.PatchTeacherRequest
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
		updates[i] = patchReqToMap(req)
	}

	// 4. seluruh transaksi (connect + begin + loop + commit/rollback) di repository
	updatedTeachers, err := sqlconnect.PatchTeachers(updates)
	if err != nil {
		switch {
		case errors.Is(err, sqlconnect.ErrTeacherNotFound):
			http.Error(w, "teacher not found", http.StatusNotFound)
		case errors.Is(err, sqlconnect.ErrInvalidTeacherID):
			http.Error(w, "invalid teacher id in update", http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

// PatchOneTeacherByIdHandler godoc
// @Summary      Patch a teacher
// @Description  Update sebagian field satu teacher berdasarkan id
// @Tags         teachers
// @Accept       json
// @Produce      json
// @Param        id    path  int                     true  "Teacher ID"
// @Param        body  body  dto.PatchTeacherRequest true  "Partial fields to update"
// @Success      200   {object}  models.Teacher
// @Failure      400   {object}  map[string]interface{}
// @Failure      404   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /teachers/{id} [patch]
func PatchOneTeacherByIdHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid teacher id", http.StatusBadRequest)
		return
	}

	// 1. Decode body ke request DTO (deteksi juga tipe mismatch)
	var req dto.PatchTeacherRequest
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
	updates := patchReqToMap(req)

	// 4. get + apply + update dilakukan sekaligus di repository
	teacher, err := sqlconnect.PatchTeacherByID(id, updates)
	if err != nil {
		if errors.Is(err, sqlconnect.ErrTeacherNotFound) {
			http.Error(w, "teacher not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teacher)
}

// DeleteTeachersHandler godoc
// @Summary      Delete teachers (batch)
// @Description  Hapus banyak teacher dalam satu transaksi. All-or-nothing.
// @Tags         teachers
// @Accept       json
// @Produce      json
// @Param        body  body  []int  true  "Array of teacher IDs"
// @Success      200   {object}  dto.TeacherDeleteResponse
// @Failure      400   {object}  map[string]interface{}
// @Failure      404   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /teachers [delete]
func DeleteTeachersHandler(w http.ResponseWriter, r *http.Request) {
	// baca body: array id
	var ids []int
	err := json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
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

// DeleteOneTeacherByIdHandler godoc
// @Summary      Delete a teacher
// @Description  Hapus satu teacher berdasarkan id
// @Tags         teachers
// @Produce      json
// @Param        id  path  int  true  "Teacher ID"
// @Success      200  {object}  dto.TeacherDeleteOneResponse
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /teachers/{id} [delete]
func DeleteOneTeacherByIdHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid teacher id", http.StatusBadRequest)
		return
	}

	rowsEffected, err := sqlconnect.DeleteTeacherByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

// writeValidationError menulis response 400 berisi daftar error per field.
func writeValidationError(w http.ResponseWriter, verrs []utils.ValidationError) {
	// 1. Tandai sebagai error + status 400
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	// 2. Tulis daftar error terstruktur
	json.NewEncoder(w).Encode(struct {
		Status string                   `json:"status"`
		Errors []utils.ValidationError `json:"errors"`
	}{
		Status: "error",
		Errors: verrs,
	})
}

// createReqToModel memetakan CreateTeacherRequest ke models.Teacher.
func createReqToModel(req dto.CreateTeacherRequest) models.Teacher {
	return models.Teacher{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Class:     req.Class,
		Subject:   req.Subject,
	}
}

// updateReqToModel memetakan UpdateTeacherRequest ke models.Teacher.
func updateReqToModel(req dto.UpdateTeacherRequest) models.Teacher {
	return models.Teacher{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Class:     req.Class,
		Subject:   req.Subject,
	}
}

// patchReqToMap mengubah PatchTeacherRequest menjadi map updates.
// Field yang nil (tidak dikirim client) tidak dimasukkan.
func patchReqToMap(req dto.PatchTeacherRequest) map[string]interface{} {
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
	if req.Subject != nil {
		updates["subject"] = *req.Subject
	}
	// 3. Kembalikan map updates
	return updates
}
