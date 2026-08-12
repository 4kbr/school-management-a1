package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"school-management-api/internal/models"
	"school-management-api/internal/models/dto"
	"school-management-api/internal/repositories/sqlconnect"
	"school-management-api/pkg/utils"
)

// ===== Helper konversi =====

// createExecReqToModel memetakan CreateExecRequest ke models.Exec.
// Password diisi LANGKAH TERPISAH oleh handler setelah di-hash (bukan di sini).
func createExecReqToModel(req dto.CreateExecRequest) models.Exec {
	return models.Exec{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Username:  req.Username,
		Role:      req.Role,
	}
}

// patchExecReqToMap mengubah PatchExecRequest menjadi map updates.
// Field yang nil (tidak dikirim client) tidak dimasukkan.
func patchExecReqToMap(req dto.PatchExecRequest) map[string]interface{} {
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
	if req.Username != nil {
		updates["username"] = *req.Username
	}
	if req.Role != nil {
		updates["role"] = *req.Role
	}
	if req.InactiveStatus != nil {
		updates["inactive_status"] = *req.InactiveStatus
	}
	// 3. Kembalikan map updates
	return updates
}

// execToResponse memetakan models.Exec ke dto.ExecResponse.
// Bertugas mengecualikan password (hash) dari response.
func execToResponse(e models.Exec) dto.ExecResponse {
	return dto.ExecResponse{
		ID:             e.ID,
		FirstName:      e.FirstName,
		LastName:       e.LastName,
		Email:          e.Email,
		Username:       e.Username,
		InactiveStatus: e.InactiveStatus,
		Role:           e.Role,
	}
}

// execListToResponse memetakan slice models.Exec ke slice dto.ExecResponse.
func execListToResponse(list []models.Exec) []dto.ExecResponse {
	out := make([]dto.ExecResponse, 0, len(list))
	for _, e := range list {
		out = append(out, execToResponse(e))
	}
	return out
}

// ===== Handler CRUD =====

// GetExecsHandler godoc
// @Summary      List execs
// @Description  Mengambil daftar exec dengan filter & sorting opsional
// @Tags         execs
// @Produce      json
// @Param        first_name  query string false "Filter by first name"
// @Param        last_name   query string false "Filter by last name"
// @Param        email       query string false "Filter by email"
// @Param        username    query string false "Filter by username"
// @Param        role        query string false "Filter by role"
// @Param        sortby      query []string false "Sort: field:asc|desc (repeatable, ex: first_name:asc)"
// @Success      200 {object} dto.ExecListResponse
// @Failure      500 {object} map[string]interface{}
// @Router       /execs [get]
func GetExecsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Delegasikan seluruh proses DB ke repository (build query + filter/sort + scan)
	execList, err := sqlconnect.GetExecs(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Response pakai DTO (password tidak disertakan)
	response := dto.ExecListResponse{
		Status: "success",
		Count:  len(execList),
		Data:   execListToResponse(execList),
	}

	// 3. Tulis response JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetOneExecHandler godoc
// @Summary      Get an exec by ID
// @Description  Mengambil satu exec berdasarkan id
// @Tags         execs
// @Produce      json
// @Param        id   path      int  true  "Exec ID"
// @Success      200  {object}  dto.ExecResponse
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /execs/{id} [get]
func GetOneExecHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Baca & konversi path parameter "id" ke int
	idStr := r.PathValue("id")
	w.Header().Set("Content-Type", "application/json")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid exec id", http.StatusBadRequest)
		return
	}

	// 2. Delegasikan query ke repository
	exec, err := sqlconnect.GetExecByID(id)
	if errors.Is(err, sqlconnect.ErrExecNotFound) {
		http.Error(w, "exec not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Response pakai DTO (password tidak disertakan)
	json.NewEncoder(w).Encode(execToResponse(*exec))
}

// AddExecsHandler godoc
// @Summary      Create an exec
// @Description  Membuat satu exec baru. Password di-hash dengan Argon2id sebelum disimpan.
// @Tags         execs
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateExecRequest  true  "Exec data"
// @Success      201   {object}  dto.ExecListResponse
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /execs [post]
func AddExecsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Decode body ke request DTO (deteksi juga tipe mismatch)
	var req dto.CreateExecRequest
	if verrs := utils.DecodeJSON(r.Body, &req); verrs != nil {
		writeValidationError(w, verrs)
		return
	}

	// 2. Validasi (full required); kalau error → 400 JSON terstruktur
	if verrs := utils.ValidateStruct(&req); verrs != nil {
		writeValidationError(w, verrs)
		return
	}

	// 3. Map DTO → model (tanpa password)
	newExec := createExecReqToModel(req)

	// 4. HASH password dengan Argon2id sebelum disimpan — JANGAN simpan plaintext
	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}
	newExec.Password = hashed

	// 5. Insert ke database
	created, err := sqlconnect.CreateExec(newExec)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 6. Response 201 pakai DTO (password tidak disertakan)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := dto.ExecListResponse{
		Status: "success",
		Count:  1,
		Data:   []dto.ExecResponse{execToResponse(*created)},
	}
	json.NewEncoder(w).Encode(response)
}

// PatchExecsHandler godoc
// @Summary      Patch execs (batch)
// @Description  Update sebagian field dari banyak exec dalam satu transaksi. All-or-nothing.
// @Tags         execs
// @Accept       json
// @Produce      json
// @Param        body  body  []dto.PatchExecRequest  true  "Array of partial updates"
// @Success      200   {object}  dto.ExecListResponse
// @Failure      400   {object}  map[string]interface{}
// @Failure      404   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /execs [patch]
func PatchExecsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Decode body ke slice request DTO (deteksi juga tipe mismatch)
	var reqs []dto.PatchExecRequest
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
		updates[i] = patchExecReqToMap(req)
	}

	// 4. seluruh transaksi (connect + begin + loop + commit/rollback) di repository
	updatedExecs, err := sqlconnect.PatchExecs(updates)
	if err != nil {
		switch {
		case errors.Is(err, sqlconnect.ErrExecNotFound):
			http.Error(w, "exec not found", http.StatusNotFound)
		case errors.Is(err, sqlconnect.ErrInvalidExecID):
			http.Error(w, "invalid exec id in update", http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// 5. Response pakai DTO (password tidak disertakan)
	w.Header().Set("Content-Type", "application/json")
	response := dto.ExecListResponse{
		Status: "success",
		Count:  len(updatedExecs),
		Data:   execListToResponse(updatedExecs),
	}
	json.NewEncoder(w).Encode(response)
}

// PatchOneExecHandler godoc
// @Summary      Patch an exec
// @Description  Update sebagian field satu exec berdasarkan id.
// @Tags         execs
// @Accept       json
// @Produce      json
// @Param        id    path  int                     true  "Exec ID"
// @Param        body  body  dto.PatchExecRequest    true  "Partial fields to update"
// @Success      200   {object}  dto.ExecResponse
// @Failure      400   {object}  map[string]interface{}
// @Failure      404   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /execs/{id} [patch]
func PatchOneExecHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Baca & konversi path parameter "id" ke int
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid exec id", http.StatusBadRequest)
		return
	}

	// 2. Decode body ke request DTO (deteksi juga tipe mismatch)
	var req dto.PatchExecRequest
	if verrs := utils.DecodeJSON(r.Body, &req); verrs != nil {
		writeValidationError(w, verrs)
		return
	}

	// 3. Untuk single patch, id berasal dari URL path, bukan body
	req.ID = id

	// 4. Validasi partial; kalau error → 400 JSON terstruktur
	if verrs := utils.ValidateStruct(&req); verrs != nil {
		writeValidationError(w, verrs)
		return
	}

	// 5. Map DTO → map updates untuk repository
	updates := patchExecReqToMap(req)

	// 6. get + apply + update dilakukan sekaligus di repository
	exec, err := sqlconnect.PatchExecByID(id, updates)
	if err != nil {
		switch {
		case errors.Is(err, sqlconnect.ErrExecNotFound):
			http.Error(w, "exec not found", http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// 7. Response pakai DTO (password tidak disertakan)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(execToResponse(*exec))
}

// DeleteOneExecHandler godoc
// @Summary      Delete an exec
// @Description  Hapus satu exec berdasarkan id
// @Tags         execs
// @Produce      json
// @Param        id  path  int  true  "Exec ID"
// @Success      200  {object}  dto.ExecDeleteResponse
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /execs/{id} [delete]
func DeleteOneExecHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Baca & konversi path parameter "id" ke int
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid exec id", http.StatusBadRequest)
		return
	}

	// 2. Eksekusi delete di repository
	rowsAffected, err := sqlconnect.DeleteExecByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Kalau 0 baris terhapus → id tidak ada di DB
	if rowsAffected == 0 {
		http.Error(w, "exec not found", http.StatusNotFound)
		return
	}

	// 4. Response sukses
	w.Header().Set("Content-Type", "application/json")
	response := dto.ExecDeleteResponse{
		Status: "success",
		ID:     id,
	}
	json.NewEncoder(w).Encode(response)
}
