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

// TeacherStudentsResponse adalah respons GET /teachers/{id}/students
// (daftar student di class milik teacher tersebut).
type TeacherStudentsResponse struct {
	Status    string           `json:"status"`
	TeacherID int              `json:"teacher_id"`
	Class     string           `json:"class"`
	Count     int              `json:"count"`
	Data      []models.Student `json:"data"`
}

// TeacherStudentCountResponse adalah respons GET /teachers/{id}/studentcount
// (jumlah student di class milik teacher tersebut).
type TeacherStudentCountResponse struct {
	Status    string `json:"status"`
	TeacherID int    `json:"teacher_id"`
	Class     string `json:"class"`
	Count     int    `json:"count"`
}
