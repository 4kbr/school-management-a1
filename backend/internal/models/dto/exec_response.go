package dto

// ExecResponse adalah respons SATU exec. Sengaja TIDAK menyertakan field
// password (hash) agar tidak bocor ke response API.
type ExecResponse struct {
	ID             int    `json:"id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Email          string `json:"email"`
	Username       string `json:"username"`
	InactiveStatus bool   `json:"inactive_status"`
	Role           string `json:"role"`
}

// ExecListResponse adalah respons untuk endpoint yang mengembalikan
// list exec: GET /execs, POST /execs, PATCH /execs.
type ExecListResponse struct {
	Status string         `json:"status"`
	Count  int            `json:"count"`
	Data   []ExecResponse `json:"data"`
}

// ExecDeleteResponse adalah respons untuk DELETE /execs/{id}.
type ExecDeleteResponse struct {
	Status string `json:"status"`
	ID     int    `json:"id"`
}
