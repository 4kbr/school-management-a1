package sqlconnect

import (
	"database/sql"
	"errors"
	"net/url"
	"reflect"
	"strings"

	"school-management-api/internal/models"
)

// Sentinel error untuk kasus "teacher tidak ditemukan" — biar handler
// bisa membedakan 404 dari error DB lainnya.
var ErrTeacherNotFound = errors.New("teacher not found")

// Sentinel error untuk id yang tidak valid di dalam payload batch update.
var ErrInvalidTeacherID = errors.New("invalid teacher id in update")

// GetTeachers membangun query SELECT dengan filter & sorting dari query params,
// menjalankannya, lalu memindai semua baris ke slice models.Teacher.
func GetTeachers(values url.Values) ([]models.Teacher, error) {
	// 1. Buka koneksi database
	db, err := ConnectDb()
	if err != nil {
		return nil, err
	}
	// 2. Koneksi wajib ditutup setelah selesai
	defer db.Close()

	// 3. Bangun query dasar + terapkan filter dari query params
	query := "SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE 1=1"
	var args []interface{}
	query, args = addFilters(values, query, args)

	// 4. Terapkan sorting dari query params (mis. "?sortby=first_name:asc")
	query = addSorting(values, query)

	// 5. Jalankan query SELECT dengan argumen filter dinamis
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	// 6. Rows wajib di-close setelah dipakai — cegah kebocoran koneksi DB
	defer rows.Close()

	// 7. Siapkan slice hasil (bukan nil) biar JSON keluar sebagai [] bukan null
	teacherList := make([]models.Teacher, 0)

	// 8. Loop tiap baris, scan ke struct Teacher, lalu append
	for rows.Next() {
		var teacher models.Teacher
		err := rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
		if err != nil {
			return nil, err
		}
		teacherList = append(teacherList, teacher)
	}

	// 9. Cek error yang muncul SELAMA iterasi (bukan hanya saat Scan)
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 10. Kembalikan hasil scan
	return teacherList, nil
}

// addSorting menambahkan klausa ORDER BY ke query dari parameter "sortby"
// dengan format "field:order" (mis. "?sortby=first_name:asc&sortby=class:desc").
func addSorting(values url.Values, query string) string {
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
			if !isValidSortField(field) || !isValidSortOrder(order) {
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

// addFilters menambahkan klausa WHERE dari parameter query (first_name, class, dst).
func addFilters(values url.Values, query string, args []interface{}) (string, []interface{}) {
	// 1. Petakan nama parameter query ke nama kolom database
	params := map[string]string{
		"first_name": "first_name",
		"last_name":  "last_name",
		"email":      "email",
		"class":      "class",
		"subject":    "subject",
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

// isValidSortOrder memastikan arah sorting cuma "asc" atau "desc".
func isValidSortOrder(order string) bool {
	return order == "asc" || order == "desc"
}

// isValidSortField memastikan field sorting termasuk daftar yang diizinkan.
func isValidSortField(field string) bool {
	validFields := map[string]bool{
		"first_name": true,
		"last_name":  true,
		"email":      true,
		"class":      true,
		"subject":    true,
	}
	return validFields[field]
}

// GetTeacherByID mengambil SATU teacher berdasarkan id.
// Mengembalikan ErrTeacherNotFound kalau id tidak ditemukan.
func GetTeacherByID(id int) (*models.Teacher, error) {
	// 1. Buka koneksi database
	db, err := ConnectDb()
	if err != nil {
		return nil, err
	}
	// 2. Koneksi wajib ditutup setelah selesai
	defer db.Close()

	// 3. Siapkan variabel penampung hasil query
	var teacher models.Teacher

	// 4. Jalankan SELECT by id dan langsung scan ke struct
	err = db.QueryRow(
		"SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?",
		id,
	).Scan(
		&teacher.ID,
		&teacher.FirstName,
		&teacher.LastName,
		&teacher.Email,
		&teacher.Class,
		&teacher.Subject,
	)
	if err == sql.ErrNoRows {
		// 5. Id tidak ada → kembalikan sentinel biar handler bisa jawab 404
		return nil, ErrTeacherNotFound
	} else if err != nil {
		// 6. Error DB lain → teruskan ke atas
		return nil, err
	}

	// 7. Berhasil → kembalikan pointer ke teacher
	return &teacher, nil
}

// CreateTeachers memasukkan banyak teacher sekaligus lewat prepared statement.
// Mengisi ID tiap teacher dari LastInsertId, lalu mengembalikan slice-nya.
func CreateTeachers(teachers []models.Teacher) ([]models.Teacher, error) {
	// 1. Buka koneksi database
	db, err := ConnectDb()
	if err != nil {
		return nil, err
	}
	// 2. Koneksi wajib ditutup setelah selesai
	defer db.Close()

	// 3. Prepare statement INSERT biar dipakai ulang tiap loop (lebih efisien & aman)
	stmt, err := db.Prepare(
		"INSERT INTO teachers (first_name, last_name, email, class, subject) VALUES (?,?,?,?,?)",
	)
	if err != nil {
		return nil, err
	}
	// 4. Statement wajib di-close setelah selesai
	defer stmt.Close()

	// 5. Siapkan slice hasil dengan panjang sama dengan input
	addedTeachers := make([]models.Teacher, len(teachers))

	// 6. Loop tiap teacher, eksekusi insert, lalu isi ID dari LastInsertId
	for i, newTeacher := range teachers {
		res, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.Class, newTeacher.Subject)
		if err != nil {
			return nil, err
		}

		// 7. Ambil auto-increment ID hasil insert
		lastId, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}

		// 8. Simpan ID hasil insert ke struct biar bisa dibalas ke client
		newTeacher.ID = int(lastId)
		addedTeachers[i] = newTeacher
	}

	// 9. Kembalikan slice teacher yang sudah berisi ID
	return addedTeachers, nil
}

// UpdateTeacher melakukan full UPDATE satu teacher berdasarkan id.
func UpdateTeacher(teacher models.Teacher) error {
	// 1. Buka koneksi database
	db, err := ConnectDb()
	if err != nil {
		return err
	}
	// 2. Koneksi wajib ditutup setelah selesai
	defer db.Close()

	// 3. Eksekusi query UPDATE dengan semua field + id sebagai kriteria
	_, err = db.Exec(
		"UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?",
		teacher.FirstName,
		teacher.LastName,
		teacher.Email,
		teacher.Class,
		teacher.Subject,
		teacher.ID,
	)
	// 4. Kembalikan error (nil kalau sukses)
	return err
}

// PatchTeacherByID menggabungkan get + apply + update untuk SATU teacher.
// Handler cukup memanggil satu fungsi ini tanpa tahu detail DB.
func PatchTeacherByID(id int, updates map[string]interface{}) (*models.Teacher, error) {
	// 1. Buka koneksi database
	db, err := ConnectDb()
	if err != nil {
		return nil, err
	}
	// 2. Koneksi wajib ditutup setelah selesai
	defer db.Close()

	// 3. Ambil teacher yang ada dulu (404 kalau tidak ketemu)
	var teacher models.Teacher
	err = db.QueryRow(
		"SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?",
		id,
	).Scan(
		&teacher.ID,
		&teacher.FirstName,
		&teacher.LastName,
		&teacher.Email,
		&teacher.Class,
		&teacher.Subject,
	)
	if err == sql.ErrNoRows {
		return nil, ErrTeacherNotFound
	} else if err != nil {
		return nil, err
	}

	// 4. Terapkan field yang dikirim body ke struct (via reflect)
	applyUpdates(&teacher, updates)

	// 5. Simpan perubahan ke database
	_, err = db.Exec(
		"UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?",
		teacher.FirstName,
		teacher.LastName,
		teacher.Email,
		teacher.Class,
		teacher.Subject,
		teacher.ID,
	)
	if err != nil {
		return nil, err
	}

	// 6. Kembalikan teacher yang sudah ter-update
	return &teacher, nil
}

// PatchTeachers mengupdate BANYAK teacher dalam SATU transaksi (all-or-nothing).
// Body: [{"id":100,"first_name":"X"}, {"id":104,"class":"9-Z"}, ...]
func PatchTeachers(updates []map[string]interface{}) ([]models.Teacher, error) {
	// 1. Buka koneksi database
	db, err := ConnectDb()
	if err != nil {
		return nil, err
	}
	// 2. Koneksi wajib ditutup setelah selesai
	defer db.Close()

	// 3. Mulai transaksi — semua operasi dalam satu transaksi biar konsisten
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	// 4. Rollback otomatis kalau ada error di tengah; no-op kalau sudah Commit
	defer tx.Rollback()

	// 5. Siapkan penampung hasil update
	updatedTeachers := make([]models.Teacher, 0, len(updates))

	// 6. Loop tiap update
	for _, update := range updates {
		// 7. JSON decode angka jadi float64 — konversi manual, BUKAN .(int)
		idFloat, ok := update["id"].(float64)
		if !ok {
			return nil, ErrInvalidTeacherID
		}
		id := int(idFloat)

		// 8. SELECT teacher by id memakai tx biar dalam transaksi yang sama
		var teacher models.Teacher
		err := tx.QueryRow(
			"SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?",
			id,
		).Scan(
			&teacher.ID,
			&teacher.FirstName,
			&teacher.LastName,
			&teacher.Email,
			&teacher.Class,
			&teacher.Subject,
		)
		if err == sql.ErrNoRows {
			return nil, ErrTeacherNotFound
		} else if err != nil {
			return nil, err
		}

		// 9. Terapkan field yang dikirim body ke struct (via reflect)
		applyUpdates(&teacher, update)

		// 10. UPDATE memakai tx
		_, err = tx.Exec(
			"UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?",
			teacher.FirstName,
			teacher.LastName,
			teacher.Email,
			teacher.Class,
			teacher.Subject,
			teacher.ID,
		)
		if err != nil {
			return nil, err
		}

		// 11. Kumpulkan teacher hasil update
		updatedTeachers = append(updatedTeachers, teacher)
	}

	// 12. Semua sukses → commit (baru beneran tersimpan)
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 13. Kembalikan list teacher yang ter-update
	return updatedTeachers, nil
}

// DeleteTeacherByID menghapus SATU teacher berdasarkan id.
// Mengembalikan jumlah baris yang terpengaruh (0 = id tidak ada).
func DeleteTeacherByID(id int) (int64, error) {
	// 1. Buka koneksi database
	db, err := ConnectDb()
	if err != nil {
		return 0, err
	}
	// 2. Koneksi wajib ditutup setelah selesai
	defer db.Close()

	// 3. Eksekusi query DELETE by id
	result, err := db.Exec("DELETE FROM teachers WHERE id = ?", id)
	if err != nil {
		return 0, err
	}

	// 4. Ambil berapa baris yang terhapus — 0 berarti id tidak ditemukan
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	// 5. Kembalikan jumlah baris terhapus
	return rowsAffected, nil
}

// DeleteTeachers menghapus BANYAK teacher dalam SATU transaksi (all-or-nothing).
// Body: [108, 109, 110] — array id yang mau dihapus.
// Kalau satu id tidak ada, SEMUA di-rollback dan mengembalikan ErrTeacherNotFound.
func DeleteTeachers(ids []int) ([]int, error) {
	// 1. Buka koneksi database
	db, err := ConnectDb()
	if err != nil {
		return nil, err
	}
	// 2. Koneksi wajib ditutup setelah selesai
	defer db.Close()

	// 3. Mulai transaksi — semua delete dalam satu transaksi biar konsisten
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	// 4. Rollback otomatis kalau ada error di tengah; no-op kalau sudah Commit
	defer tx.Rollback()

	// 5. Siapkan penampung id yang beneran terhapus
	deleted := make([]int, 0, len(ids))

	// 6. Loop tiap id, delete pakai tx (bukan db) biar dalam transaksi yang sama
	for _, id := range ids {
		result, err := tx.Exec("DELETE FROM teachers WHERE id = ?", id)
		if err != nil {
			return nil, err
		}

		// 7. Cek berapa baris terhapus lewat RowsAffected
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return nil, err
		}

		// 8. Kalau 0 baris = id gak ada di DB → error, semua di-rollback
		if rowsAffected == 0 {
			return nil, ErrTeacherNotFound
		}

		// 9. Kumpulkan id yang beneran terhapus
		deleted = append(deleted, id)
	}

	// 10. Semua sukses → commit (baru beneran tersimpan)
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 11. Kembalikan list id yang terhapus
	return deleted, nil
}

// applyUpdates mengubah field struct models.Teacher sesuai map updates (via reflect).
// - Cocokkan key body ke json tag field struct
// - Skip "id" biar user gak bisa ubah ID
// - type-assertion aman: value tipe salah di-skip, bukan panic
func applyUpdates(existingTeacher *models.Teacher, updates map[string]interface{}) {
	// .Elem() karena pointer — biar field bisa di-Set
	teacherVal := reflect.ValueOf(existingTeacher).Elem()
	// metadata tipe: nama field + json tag
	teacherType := teacherVal.Type()

	for k, v := range updates {
		// id gak boleh di-update lewat reflect — skip biar user gak bisa ubah ID
		if k == "id" {
			continue
		}

		// loop tiap field struct, cari yang json tag-nya cocok sama key body
		for i := 0; i < teacherVal.NumField(); i++ {
			field := teacherType.Field(i)
			jsonTag := field.Tag.Get("json")
			// Extract the field name from json tag (e.g., "first_name,omitempty" -> "first_name")
			jsonFieldName := strings.Split(jsonTag, ",")[0]

			if jsonFieldName == k {
				fieldVal := teacherVal.Field(i)
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
