// Package dto berisi request/response object (DTO) khusus autentikasi (login/logout).
package dto

// LoginRequest dipakai di POST /execs/login.
// Email & password keduanya required.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse adalah respons sukses login. Token disertakan DI SINI (payload)
// sekaligus diset sebagai cookie oleh handler.
// Data exec tidak menyertakan password (dipakai ulang ExecResponse).
type LoginResponse struct {
	Status  string       `json:"status"`
	Message string       `json:"message"`
	Token   string       `json:"token"`
	Data    ExecResponse `json:"data"`
}

// LogoutResponse adalah respons sukses logout. Handler bertugas menghapus cookie.
type LogoutResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}