// Package dto berisi request/response object (DTO) yang memisahkan
// kontrak request dari model persistensi (models.Exec).
package dto

// CreateExecRequest dipakai di POST /execs (create).
// Semua field required karena create harus data lengkap.
type CreateExecRequest struct {
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,max=100"`
	Email     string `json:"email" validate:"required,email"`
	Username  string `json:"username" validate:"required,min=3,max=50"`
	Password  string `json:"password" validate:"required,min=8,max=100"`
	Role      string `json:"role" validate:"required,max=50"`
}

// PatchExecRequest dipakai di PATCH (partial update, single & batch).
// Field bertipe pointer (*string / *bool): nil = tidak dikirim client.
// Tag `omitempty` membuat validasi dilewati kalau field nil.
// CATATAN: tidak ada field password — ganti password lewat endpoint terpisah.
type PatchExecRequest struct {
	ID             int     `json:"id" validate:"required"`
	FirstName      *string `json:"first_name" validate:"omitempty,min=1,max=100"`
	LastName       *string `json:"last_name" validate:"omitempty,max=100"`
	Email          *string `json:"email" validate:"omitempty,email"`
	Username       *string `json:"username" validate:"omitempty,min=3,max=50"`
	Role           *string `json:"role" validate:"omitempty,max=50"`
	InactiveStatus *bool   `json:"inactive_status" validate:"omitempty"`
}
