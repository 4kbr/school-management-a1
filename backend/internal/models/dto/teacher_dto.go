// Package dto berisi request/response object (DTO) yang memisahkan
// kontrak request dari model persistensi (models.Teacher).
package dto

// CreateTeacherRequest dipakai di POST /teachers/ (create batch).
// Semua field required karena create harus data lengkap.
type CreateTeacherRequest struct {
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,max=100"`
	Email     string `json:"email" validate:"required,email"`
	Class     string `json:"class" validate:"required,max=50"`
	Subject   string `json:"subject" validate:"required,max=100"`
}

// UpdateTeacherRequest dipakai di PUT /teachers/{id} (full replace).
// Semua field required, sama seperti create.
type UpdateTeacherRequest struct {
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,max=100"`
	Email     string `json:"email" validate:"required,email"`
	Class     string `json:"class" validate:"required,max=50"`
	Subject   string `json:"subject" validate:"required,max=100"`
}

// PatchTeacherRequest dipakai di PATCH (partial update, single & batch).
// Field bertipe pointer (*string): nil = tidak dikirim client.
// Tag `omitempty` membuat validasi dilewati kalau field nil.
type PatchTeacherRequest struct {
	ID        int     `json:"id" validate:"required"`
	FirstName *string `json:"first_name" validate:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name" validate:"omitempty,max=100"`
	Email     *string `json:"email" validate:"omitempty,email"`
	Class     *string `json:"class" validate:"omitempty,max=50"`
	Subject   *string `json:"subject" validate:"omitempty,max=100"`
}

// TeacherStudentsRequest dipakai di GET /teachers/{id}/students.
// Menampung query param untuk filter & sort daftar student milik teacher tsb.
type TeacherStudentsRequest struct {
	FirstName string   `json:"first_name" validate:"omitempty,max=100"`
	LastName  string   `json:"last_name" validate:"omitempty,max=100"`
	Email     string   `json:"email" validate:"omitempty,email"`
	SortBy    []string `json:"sortby"`
}
