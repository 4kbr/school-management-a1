package dto

import "school-management-api/internal/models"

// StudentListResponse adalah respons untuk endpoint yang mengembalikan
// list student: GET/POST/PATCH /students.
type StudentListResponse struct {
	Status string          `json:"status"`
	Count  int             `json:"count"`
	Data   []models.Student `json:"data"`
}

// StudentDeleteResponse adalah respons untuk DELETE /students (batch).
type StudentDeleteResponse struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
	IDs    []int  `json:"ids"`
}

// StudentDeleteOneResponse adalah respons untuk DELETE /students/{id}.
type StudentDeleteOneResponse struct {
	Status string `json:"status"`
	ID     int    `json:"id"`
}
