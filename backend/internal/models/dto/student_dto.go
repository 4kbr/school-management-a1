package dto

// CreateStudentRequest dipakai di POST /students/ (create batch).
// Semua field required karena create harus data lengkap.
type CreateStudentRequest struct {
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,max=100"`
	Email     string `json:"email" validate:"required,email"`
	Class     string `json:"class" validate:"required,max=50"`
}

// UpdateStudentRequest dipakai di PUT /students/{id} (full replace).
// Semua field required, sama seperti create.
type UpdateStudentRequest struct {
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,max=100"`
	Email     string `json:"email" validate:"required,email"`
	Class     string `json:"class" validate:"required,max=50"`
}

// PatchStudentRequest dipakai di PATCH (partial update, single & batch).
// Field bertipe pointer (*string): nil = tidak dikirim client.
// Tag `omitempty` membuat validasi dilewati kalau field nil.
type PatchStudentRequest struct {
	ID        int     `json:"id" validate:"required"`
	FirstName *string `json:"first_name" validate:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name" validate:"omitempty,max=100"`
	Email     *string `json:"email" validate:"omitempty,email"`
	Class     *string `json:"class" validate:"omitempty,max=50"`
}
