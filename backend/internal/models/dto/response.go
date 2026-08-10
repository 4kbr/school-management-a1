package dto

import "school-management-api/internal/models"

// TeacherListResponse adalah respons untuk endpoint yang mengembalikan
// list teacher: GET/POST/PATCH /teachers.
type TeacherListResponse struct {
	Status string           `json:"status"`
	Count  int              `json:"count"`
	Data   []models.Teacher `json:"data"`
}

// TeacherDeleteResponse adalah respons untuk DELETE /teachers (batch).
type TeacherDeleteResponse struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
	IDs    []int  `json:"ids"`
}

// TeacherDeleteOneResponse adalah respons untuk DELETE /teachers/{id}.
type TeacherDeleteOneResponse struct {
	Status string `json:"status"`
	ID     int    `json:"id"`
}
