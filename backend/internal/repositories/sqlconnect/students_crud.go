package sqlconnect

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"school-management-api/internal/models"
	"school-management-api/pkg/utils"
)

// Sentinel error untuk kasus "student tidak ditemukan" — biar handler
// bisa membedakan 404 dari error DB lainnya.
var ErrStudentNotFound = errors.New("student not found")

// Sentinel error untuk id yang tidak valid di dalam payload batch update.
var ErrInvalidStudentID = errors.New("invalid student id in update")

// Sentinel error untuk class yang tidak ada di tabel teachers (FK constraint).
var ErrClassNotFound = errors.New("class not found")

// GetStudents membangun query SELECT dengan filter & sorting dari query params,
// menjalankannya, lalu memindai semua baris ke slice models.Student.
func GetStudents(values url.Values) ([]models.Student, error) {
	// 1. Buka koneksi database
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "database query error")
	}
	// 2. Koneksi wajib ditutup setelah selesai
	defer db.Close()

	// 3. Bangun query dasar + terapkan filter dari query params
	query := "SELECT id, first_name, last_name, email, class FROM students WHERE 1=1"
	var args []interface{}
	query, args = addStudentFilters(values, query, args)

	// 4. Terapkan sorting dari query params (mis. "?sortby=first_name:asc")
	query = addStudentSorting(values, query)

	// 5. Jalankan query SELECT dengan argumen filter dinamis
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, utils.ErrorHandler(err, "database query error")
	}
	// 6. Rows wajib di-close setelah dipakai — cegah kebocoran koneksi DB
	defer rows.Close()

	// 7. Siapkan slice hasil (bukan nil) biar JSON keluar sebagai [] bukan null
	studentList := make([]models.Student, 0)

	// 8. Loop tiap baris, scan ke struct Student, lalu append
	for rows.Next() {
		var student models.Student
		err := rows.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class)
		if err != nil {
			return nil, utils.ErrorHandler(err, "database query error")
		}
		studentList = append(studentList, student)
	}

	// 9. Cek error yang muncul SELAMA iterasi (bukan hanya saat Scan)
	if err := rows.Err(); err != nil {
		return nil, utils.ErrorHandler(err, "database query error")
	}

	// 10. Kembalikan hasil scan
	return studentList, nil
}

// addStudentSorting menambahkan klausa ORDER BY ke query dari parameter "sortby"
// dengan format "field:order" (mis. "?sortby=first_name:asc&sortby=class:desc").
func addStudentSorting(values url.Values, query string) string {
	// 1. Ambil semua nilai parameter sortby
	sortParams := values["sortby"]
	if len(sortParams) > 0 {
		// 2. Mulai klausa ORDER BY
		query += " ORDER BY "
		for i, param := range sortParams {
			// 3. Pisahkan "field:order" — format salah di-skip
			parts := strings.Split(param, ":")
			if len(parts) != 2 {
				continue
			}
			field, order := parts[0], parts[1]
			// 4. Hanya terima field & order yang valid (anti SQL injection)
			if !isValidStudentSortField(field) || !isValidSortOrder(order) {
				continue
			}
			// 5. Pisahkan kolom sorting dengan koma
			if i > 0 {
				query += " , "
			}
			query += " " + field + " " + order
		}
	}
	return query
}

// addStudentFilters menambahkan klausa WHERE dari parameter query (first_name, class, dst).
func addStudentFilters(values url.Values, query string, args []interface{}) (string, []interface{}) {
	// 1. Petakan nama parameter query ke nama kolom database
	params := map[string]string{
		"first_name": "first_name",
		"last_name":  "last_name",
		"email":      "email",
		"class":      "class",
	}
	// 2. Loop tiap filter yang didukung
	for param, dbField := range params {
		value := values.Get(param)
		// 3. Kalau nilainya ada, tambahkan kondisi AND + argumen
		if value != "" {
			query += " AND " + dbField + " = ?"
			args = append(args, value)
		}
	}
	return query, args
}

// isValidStudentSortField memastikan field sorting termasuk daftar yang diizinkan.
func isValidStudentSortField(field string) bool {
	validFields := map[string]bool{
		"first_name": true,
		"last_name":  true,
		"email":      true,
		"class":      true,
	}
	return validFields[field]
}

// GetStudentByID mengambil SATU student berdasarkan id.
// Mengembalikan ErrStudentNotFound kalau id tidak ditemukan.
func GetStudentByID(id int) (*models.Student, error) {
	// 1. Buka koneksi database
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "database query error")
	}
	// 2. Koneksi wajib ditutup setelah selesai
	defer db.Close()

	// 3. Siapkan variabel penampung hasil query
	var student models.Student

	// 4. Jalankan SELECT by id dan langsung scan ke struct
	err = db.QueryRow(
		"SELECT id, first_name, last_name, email, class FROM students WHERE id = ?",
		id,
	).Scan(
		&student.ID,
		&student.FirstName,
		&student.LastName,
		&student.Email,
		&student.Class,
	)
	if err == sql.ErrNoRows {
		// 5. Id tidak ada → kembalikan sentinel biar handler bisa jawab 404
		return nil, ErrStudentNotFound
	} else if err != nil {
		// 6. Error DB lain → log detail & return pesan ramah
		return nil, utils.ErrorHandler(err, "database query error")
	}

	// 7. Berhasil → kembalikan pointer ke student
	return &student, nil
}

// CreateStudents memasukkan banyak student sekaligus lewat prepared statement.
// Mengisi ID tiap student dari LastInsertId, lalu mengembalikan slice-nya.
// Sebelum insert, class tiap student divalidasi ada di tabel teachers (FK).
func CreateStudents(students []models.Student) ([]models.Student, error) {
	// 1. Buka koneksi database
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error inserting data into database")
	}
	// 2. Koneksi wajib ditutup setelah selesai
	defer db.Close()

	// 3. Prepare statement INSERT biar dipakai ulang tiap loop (lebih efisien & aman)
	stmt, err := db.Prepare(
		generateStudentInsertQuery(models.Student{}),
	)
	if err != nil {
		return nil, utils.ErrorHandler(err, "error inserting data into database")
	}
	// 4. Statement wajib di-close setelah selesai
	defer stmt.Close()

	// 5. Siapkan slice hasil dengan panjang sama dengan input
	addedStudents := make([]models.Student, len(students))

	// 6. Loop tiap student, validasi class, eksekusi insert, lalu isi ID
	for i, newStudent := range students {
		// 6a. Validasi FK: class harus sudah ada di tabel teachers
		if !classExists(db, newStudent.Class) {
			return nil, ErrClassNotFound
		}

		// 7. Ambil nilai field + eksekusi insert
		values := getStudentStructValues(newStudent)
		res, err := stmt.Exec(values...)
		if err != nil {
			return nil, utils.ErrorHandler(err, "error inserting data into database")
		}

		// 8. Ambil auto-increment ID hasil insert
		lastId, err := res.LastInsertId()
		if err != nil {
			return nil, utils.ErrorHandler(err, "error inserting data into database")
		}

		// 9. Simpan ID hasil insert ke struct biar bisa dibalas ke client
		newStudent.ID = int(lastId)
		addedStudents[i] = newStudent
	}

	// 10. Kembalikan slice student yang sudah berisi ID
	return addedStudents, nil
}

// generateStudentInsertQuery membuat query INSERT dinamis dari struct tags `db`.
// Contoh hasil: "INSERT INTO students (first_name, last_name, ...) VALUES (?,?,...)"
func generateStudentInsertQuery(model interface{}) string {
	// 1. Ambil type info struct lewat reflect (metadata field + tag)
	modelType := reflect.TypeOf(model)
	// 2. Siapkan penampung nama kolom & placeholder (?)
	var columns, placeholders string

	// 3. Loop tiap field struct
	for i := 0; i < modelType.NumField(); i++ {
		// 4. Baca tag `db` pada field
		dbTag := modelType.Field(i).Tag.Get("db")
		// 5. Rapikan tag: buang suffix ",omitempty" kalau ada
		dbTag = strings.TrimSuffix(dbTag, ",omitempty")

		// 6. Lewati field tanpa tag db atau field id (auto-increment, tidak ikut INSERT)
		if dbTag != "" && dbTag != "id" {
			// 7. Tambahkan koma pemisah untuk kolom & placeholder berikutnya
			if columns != "" {
				columns += ", "
				placeholders += ", "
			}
			// 8. Akumulasi nama kolom dan placeholder
			columns += dbTag
			placeholders += "?"
		}
	}

	// 9. Susun query INSERT lengkap
	return fmt.Sprintf("INSERT INTO students (%s) VALUES (%s)", columns, placeholders)
}

// getStudentStructValues mengambil nilai tiap field struct (berdasarkan tag `db`)
// dalam urutan yang SAMA dengan kolom di generateStudentInsertQuery.
func getStudentStructValues(model interface{}) []interface{} {
	// 1. Ambil value + type info struct lewat reflect
	modelValue := reflect.ValueOf(model)
	modelType := modelValue.Type()
	// 2. Siapkan slice penampung nilai field
	values := []interface{}{}

	// 3. Loop tiap field struct
	for i := 0; i < modelType.NumField(); i++ {
		// 4. Baca tag `db` pada field
		dbTag := modelType.Field(i).Tag.Get("db")
		// 5. Rapikan tag: buang suffix ",omitempty" kalau ada
		dbTag = strings.TrimSuffix(dbTag, ",omitempty")

		// 6. Lewati field tanpa tag db atau field id — jumlah values harus
		//    sama dengan jumlah placeholder di generateStudentInsertQuery
		if dbTag != "" && dbTag != "id" {
			// 7. Ambil nilai field dan append ke slice
			values = append(values, modelValue.Field(i).Interface())
		}
	}

	// 8. Kembalikan slice nilai field
	return values
}

// classExists mengecek apakah sebuah class sudah ada di tabel teachers.
// Dipakai untuk memvalidasi FK students.class → teachers.class.
func classExists(db *sql.DB, class string) bool {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM teachers WHERE class = ?", class).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// classExistsTx sama seperti classExists, tapi memakai transaksi (untuk batch).
func classExistsTx(tx *sql.Tx, class string) bool {
	var count int
	err := tx.QueryRow("SELECT COUNT(*) FROM teachers WHERE class = ?", class).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// UpdateStudent melakukan full UPDATE satu student berdasarkan id.
// Class baru divalidasi dulu (FK ke teachers).
func UpdateStudent(student models.Student) error {
	// 1. Buka koneksi database
	db, err := ConnectDb()
	if err != nil {
		return utils.ErrorHandler(err, "error updating student")
	}
	// 2. Koneksi wajib ditutup setelah selesai
	defer db.Close()

	// 3. Validasi FK: class harus sudah ada di tabel teachers
	if !classExists(db, student.Class) {
		return ErrClassNotFound
	}

	// 4. Eksekusi query UPDATE dengan semua field + id sebagai kriteria
	_, err = db.Exec(
		"UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ? WHERE id = ?",
		student.FirstName,
		student.LastName,
		student.Email,
		student.Class,
		student.ID,
	)
	if err != nil {
		// 5. Log detail & return pesan ramah
		return utils.ErrorHandler(err, "error updating student")
	}
	return nil
}

// PatchStudentByID menggabungkan get + apply + update untuk SATU student.
// Handler cukup memanggil satu fungsi ini tanpa tahu detail DB.
func PatchStudentByID(id int, updates map[string]interface{}) (*models.Student, error) {
	// 1. Buka koneksi database
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "unable to retrieve student")
	}
	// 2. Koneksi wajib ditutup setelah selesai
	defer db.Close()

	// 3. Ambil student yang ada dulu (404 kalau tidak ketemu)
	var student models.Student
	err = db.QueryRow(
		"SELECT id, first_name, last_name, email, class FROM students WHERE id = ?",
		id,
	).Scan(
		&student.ID,
		&student.FirstName,
		&student.LastName,
		&student.Email,
		&student.Class,
	)
	if err == sql.ErrNoRows {
		return nil, ErrStudentNotFound
	} else if err != nil {
		return nil, utils.ErrorHandler(err, "unable to retrieve student")
	}

	// 4. Terapkan field yang dikirim body ke struct (via reflect)
	applyStudentUpdates(&student, updates)

	// 5. Validasi FK: class baru harus ada di tabel teachers
	if !classExists(db, student.Class) {
		return nil, ErrClassNotFound
	}

	// 6. Simpan perubahan ke database
	_, err = db.Exec(
		"UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ? WHERE id = ?",
		student.FirstName,
		student.LastName,
		student.Email,
		student.Class,
		student.ID,
	)
	if err != nil {
		return nil, utils.ErrorHandler(err, "error updating student")
	}

	// 7. Kembalikan student yang sudah ter-update
	return &student, nil
}

// PatchStudents mengupdate BANYAK student dalam SATU transaksi (all-or-nothing).
// Body: [{"id":100,"first_name":"X"}, {"id":104,"class":"9-Z"}, ...]
func PatchStudents(updates []map[string]interface{}) ([]models.Student, error) {
	// 1. Buka koneksi database
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error updating student")
	}
	// 2. Koneksi wajib ditutup setelah selesai
	defer db.Close()

	// 3. Mulai transaksi — semua operasi dalam satu transaksi biar konsisten
	tx, err := db.Begin()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error starting transaction")
	}
	// 4. Rollback otomatis kalau ada error di tengah; no-op kalau sudah Commit
	defer tx.Rollback()

	// 5. Siapkan penampung hasil update
	updatedStudents := make([]models.Student, 0, len(updates))

	// 6. Loop tiap update
	for _, update := range updates {
		// 7. Ambil id dari map — dikirim sebagai int dari DTO
		id, ok := update["id"].(int)
		if !ok {
			return nil, ErrInvalidStudentID
		}

		// 8. SELECT student by id memakai tx biar dalam transaksi yang sama
		var student models.Student
		err := tx.QueryRow(
			"SELECT id, first_name, last_name, email, class FROM students WHERE id = ?",
			id,
		).Scan(
			&student.ID,
			&student.FirstName,
			&student.LastName,
			&student.Email,
			&student.Class,
		)
		if err == sql.ErrNoRows {
			return nil, ErrStudentNotFound
		} else if err != nil {
			return nil, utils.ErrorHandler(err, "unable to retrieve student")
		}

		// 9. Terapkan field yang dikirim body ke struct (via reflect)
		applyStudentUpdates(&student, update)

		// 10. Validasi FK: class baru harus ada di tabel teachers
		if !classExistsTx(tx, student.Class) {
			return nil, ErrClassNotFound
		}

		// 11. UPDATE memakai tx
		_, err = tx.Exec(
			"UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ? WHERE id = ?",
			student.FirstName,
			student.LastName,
			student.Email,
			student.Class,
			student.ID,
		)
		if err != nil {
			return nil, utils.ErrorHandler(err, "error updating student")
		}

		// 12. Kumpulkan student hasil update
		updatedStudents = append(updatedStudents, student)
	}

	// 13. Semua sukses → commit (baru beneran tersimpan)
	if err := tx.Commit(); err != nil {
		return nil, utils.ErrorHandler(err, "error committing transaction")
	}

	// 14. Kembalikan list student yang ter-update
	return updatedStudents, nil
}

// DeleteStudentByID menghapus SATU student berdasarkan id.
// Mengembalikan jumlah baris yang terpengaruh (0 = id tidak ada).
func DeleteStudentByID(id int) (int64, error) {
	// 1. Buka koneksi database
	db, err := ConnectDb()
	if err != nil {
		return 0, utils.ErrorHandler(err, "error deleting result")
	}
	// 2. Koneksi wajib ditutup setelah selesai
	defer db.Close()

	// 3. Eksekusi query DELETE by id
	result, err := db.Exec("DELETE FROM students WHERE id = ?", id)
	if err != nil {
		return 0, utils.ErrorHandler(err, "error deleting result")
	}

	// 4. Ambil berapa baris yang terhapus — 0 berarti id tidak ditemukan
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, utils.ErrorHandler(err, "error deleting result")
	}

	// 5. Kembalikan jumlah baris terhapus
	return rowsAffected, nil
}

// DeleteStudents menghapus BANYAK student dalam SATU transaksi (all-or-nothing).
// Body: [108, 109, 110] — array id yang mau dihapus.
// Kalau satu id tidak ada, SEMUA di-rollback dan mengembalikan ErrStudentNotFound.
func DeleteStudents(ids []int) ([]int, error) {
	// 1. Buka koneksi database
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error deleting result")
	}
	// 2. Koneksi wajib ditutup setelah selesai
	defer db.Close()

	// 3. Mulai transaksi — semua delete dalam satu transaksi biar konsisten
	tx, err := db.Begin()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error starting transaction")
	}
	// 4. Rollback otomatis kalau ada error di tengah; no-op kalau sudah Commit
	defer tx.Rollback()

	// 5. Siapkan penampung id yang beneran terhapus
	deleted := make([]int, 0, len(ids))

	// 6. Loop tiap id, delete pakai tx (bukan db) biar dalam transaksi yang sama
	for _, id := range ids {
		result, err := tx.Exec("DELETE FROM students WHERE id = ?", id)
		if err != nil {
			return nil, utils.ErrorHandler(err, "error deleting result")
		}

		// 7. Cek berapa baris terhapus lewat RowsAffected
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return nil, utils.ErrorHandler(err, "error deleting result")
		}

		// 8. Kalau 0 baris = id gak ada di DB → error, semua di-rollback
		if rowsAffected == 0 {
			return nil, ErrStudentNotFound
		}

		// 9. Kumpulkan id yang beneran terhapus
		deleted = append(deleted, id)
	}

	// 10. Semua sukses → commit (baru beneran tersimpan)
	if err := tx.Commit(); err != nil {
		return nil, utils.ErrorHandler(err, "error committing transaction")
	}

	// 11. Kembalikan list id yang terhapus
	return deleted, nil
}

// applyStudentUpdates mengubah field struct models.Student sesuai map updates (via reflect).
// - Cocokkan key body ke json tag field struct
// - Skip "id" biar user gak bisa ubah ID
// - type-assertion aman: value tipe salah di-skip, bukan panic
func applyStudentUpdates(existingStudent *models.Student, updates map[string]interface{}) {
	// .Elem() karena pointer — biar field bisa di-Set
	studentVal := reflect.ValueOf(existingStudent).Elem()
	// metadata tipe: nama field + json tag
	studentType := studentVal.Type()

	for k, v := range updates {
		// id gak boleh di-update lewat reflect — skip biar user gak bisa ubah ID
		if k == "id" {
			continue
		}

		// loop tiap field struct, cari yang json tag-nya cocok sama key body
		for i := 0; i < studentVal.NumField(); i++ {
			field := studentType.Field(i)
			jsonTag := field.Tag.Get("json")
			// Extract the field name from json tag (e.g., "first_name,omitempty" -> "first_name")
			jsonFieldName := strings.Split(jsonTag, ",")[0]

			if jsonFieldName == k {
				fieldVal := studentVal.Field(i)
				// field cuma bisa di-Set kalau exported + lewat pointer (.Elem())
				if fieldVal.CanSet() {
					switch fieldVal.Kind() {
					case reflect.String:
						// type-assertion AMAN: kalau v bukan string, skip, jangan panic
						if str, ok := v.(string); ok {
							fieldVal.SetString(str)
						}
					case reflect.Int, reflect.Int64:
						// JSON decode angka jadi float64 — konversi dulu ke int64
						if intVal, ok := v.(float64); ok {
							fieldVal.SetInt(int64(intVal))
						}
					}
				}
				// udah ketemu field yang cocok, stop loop field, lanjut key berikutnya
				break
			}
		}
	}
}
