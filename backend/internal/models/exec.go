package models

import "database/sql"

type Exec struct {
	ID                  int            `json:"id,omitempty" db:"id"`
	FirstName           string         `json:"first_name,omitempty" db:"first_name"`
	LastName            string         `json:"last_name,omitempty" db:"last_name"`
	Email               string         `json:"email,omitempty" db:"email"`
	Username            string         `json:"username,omitempty" db:"username"`
	Password            string         `json:"password,omitempty" db:"password"`
	PasswordChangedAt   sql.NullString `json:"password_changed_at" db:"password_changed_at"`
	UserCreatedAt       sql.NullString `json:"user_created_at" db:"user_created_at"`
	PasswordResetCode   sql.NullString `json:"password_reset_code" db:"password_reset_code"`
	PasswordCodeExpires sql.NullString `json:"password_code_expires" db:"password_code_expires"`
	InactiveStatus      bool           `json:"inactive_status,omitempty" db:"inactive_status"`
	Role                string         `json:"role,omitempty" db:"role"`
}
