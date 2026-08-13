package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"school-management-api/internal/models/dto"
	"school-management-api/internal/repositories/sqlconnect"
	"school-management-api/pkg/utils"
)

// LoginHandler godoc
// @Summary      Login an exec
// @Description  Autentikasi exec via email & password. Password diverifikasi dengan Argon2id. Sukses → token JWT dikirim di cookie sekaligus di JSON payload. Akun dengan inactive_status=true ditolak.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.LoginRequest  true  "Email & password"
// @Success      200  {object}  dto.LoginResponse
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /execs/login [post]
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Decode body ke request DTO (deteksi juga tipe mismatch)
	var req dto.LoginRequest
	if verrs := utils.DecodeJSON(r.Body, &req); verrs != nil {
		writeValidationError(w, verrs)
		return
	}

	// 2. Validasi (email + password required); kalau error → 400 JSON terstruktur
	if verrs := utils.ValidateStruct(&req); verrs != nil {
		writeValidationError(w, verrs)
		return
	}

	// 3. Ambil exec berdasarkan email (404 kalau email tidak ada)
	exec, err := sqlconnect.GetExecByEmail(req.Email)
	if err != nil {
		if errors.Is(err, sqlconnect.ErrExecNotFound) {
			http.Error(w, "invalid email or password", http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Tolak akun nonaktif (403)
	if exec.InactiveStatus {
		http.Error(w, "account is inactive", http.StatusForbidden)
		return
	}

	// 5. Verifikasi password (argon2) — 401 kalau salah
	matched, err := utils.VerifyPassword(exec.Password, req.Password)
	if err != nil {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}
	if !matched {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	// 6. Buat token JWT (id, email, role)
	token, err := utils.CreateToken(exec.ID, exec.Email, exec.Role)
	if err != nil {
		http.Error(w, "failed to create token", http.StatusInternalServerError)
		return
	}

	// 7. Set token ke cookie. HttpOnly = tidak bisa dibaca JS; SameSite=Lax;
	//    Secure dikontrol env biar tetap jalan di dev (cert self-signed).
	secure := os.Getenv("COOKIE_SECURE") == "true"
	http.SetCookie(w, &http.Cookie{
		Name:     utils.AuthCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})

	// 8. Response sukses: token di payload + data exec (tanpa password)
	w.Header().Set("Content-Type", "application/json")
	response := dto.LoginResponse{
		Status:  "success",
		Message: "login successful",
		Token:   token,
		Data:    execToResponse(*exec),
	}
	json.NewEncoder(w).Encode(response)
}

// LogoutHandler godoc
// @Summary      Logout an exec
// @Description  Menghapus cookie autentikasi sehingga token di client tidak berlaku lagi.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  dto.LogoutResponse
// @Router       /execs/logout [post]
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Hapus cookie dengan meng-set nilainya kosong + masa berlaku lewat
	secure := os.Getenv("COOKIE_SECURE") == "true"
	http.SetCookie(w, &http.Cookie{
		Name:     utils.AuthCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(1, 0), // epoch — langsung kedaluwarsa di browser lama
	})

	// 2. Response sukses
	w.Header().Set("Content-Type", "application/json")
	response := dto.LogoutResponse{
		Status:  "success",
		Message: "logout successful",
	}
	json.NewEncoder(w).Encode(response)
}