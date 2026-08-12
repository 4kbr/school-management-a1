package sqlconnect

import (
	"database/sql"
	"errors"
	"net/url"
	"reflect"
	"strings"

	"school-management-api/internal/models"
	"school-management-api/pkg/utils"
)

// Sentinel error untuk kasus "exec tidak ditemukan" — biar handler
// bisa membedakan 404 dari error DB lainnya.
var ErrExecNotFound = errors.New("exec not found")

// Sentinel error untuk id yang tidak valid di dalam payload batch update.
var ErrInvalidExecID = errors.New("invalid exec id in update")

// daftar kolom yang di-SELECT + di-SCAN. Ditulis manual (bukan via refleksi)
// karena model Exec punya field sql.NullString yang tidak cocok untuk
// helper generateInsertQuery/getStructValues milik Teacher/Student.
const execSelectCols = "id, first_name, last_name, email, username, password, password_changed_at, user_created_at, password_reset_code, password_code_expires, inactive_status, role"

// scanExec memindai satu baris hasil query ke struct models.Exec.
// Dipakai ulang oleh semua fungsi agar tidak duplikasi logika Scan.
func scanExec(row interface{ Scan(...any) error }) (*models.Exec, error) {
	var e models.Exec
	err := row.Scan(
		&e.ID,
		&e.FirstName,
		&e.LastName,
		&e.Email,
		&e.Username,
		&e.Password,
		&e.PasswordChangedAt,   // sql.NullString
		&e.UserCreatedAt,       // sql.NullString
		&e.PasswordResetCode,   // sql.NullString
		&e.PasswordCodeExpires, // sql.NullString
		&e.InactiveStatus,
		&e.Role,
	)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// GetExecs membangun query SELECT dengan filter & sorting dari query params,
// menjalankannya, lalu memindai semua baris ke slice models.Exec.
func GetExecs(values url.Values) ([]models.Exec, error) {
	// 1. Buka koneksi database
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "database query error")
	}
	// 2. Koneksi wajib ditutup setelah selesai
	defer db.Close()

	// 3. Bangun query dasar + terapkan filter dari query params
	query := "SELECT " + execSelectCols + " FROM execs WHERE 1=1"
	var args []interface{}
	query, args = addExecFilters(values, query, args)

	// 4. Terapkan sorting dari query params (mis. "?sortby=first_name:asc")
	query = addExecSorting(values, query)

	// 5. Jalankan query SELECT dengan argumen filter dinamis
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, utils.ErrorHandler(err, "database query error")
	}
	// 6. Rows wajib di-close setelah dipakai — cegah kebocoran koneksi DB
	defer rows.Close()

	// 7. Siapkan slice hasil (bukan nil) biar JSON keluar sebagai [] bukan null
	execList := make([]models.Exec, 0)

	// 8. Loop tiap baris, scan ke struct Exec, lalu append
	for rows.Next() {
		exec, err := scanExec(rows)
		if err != nil {
			return nil, utils.ErrorHandler(err, "database query error")
		}
		execList = append(execList, *exec)
	}

	// 9. Cek error yang muncul SELAMA iterasi (bukan hanya saat Scan)
	if err := rows.Err(); err != nil {
		return nil, utils.ErrorHandler(err, "database query error")
	}

	// 10. Kembalikan hasil scan
	return execList, nil
}

// addExecSorting menambahkan klausa ORDER BY ke query dari parameter "sortby"
// dengan format "field:order" (mis. "?sortby=first_name:asc&sortby=username:desc").
func addExecSorting(values url.Values, query string) string {
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
			if !isValidExecSortField(field) || !isValidSortOrder(order) {
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

// addExecFilters menambahkan klausa WHERE dari parameter query (first_name, email, dst).
func addExecFilters(values url.Values, query string, args []interface{}) (string, []interface{}) {
	// 1. Petakan nama parameter query ke nama kolom database
	params := map[string]string{
		"first_name": "first_name",
		"last_name":  "last_name",
		"email":      "email",
		"username":   "username",
		"role":       "role",
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

// isValidExecSortField memastikan field sorting termasuk daftar yang diizinkan.
func isValidExecSortField(field string) bool {
	validFields := map[string]bool{
		"first_name": true,
		"last_name":  true,
		"email":      true,
		"username":   true,
		"role":       true,
	}
	return validFields[field]
}

// GetExecByID mengambil SATU exec berdasarkan id.
// Mengembalikan ErrExecNotFound kalau id tidak ditemukan.
func GetExecByID(id int) (*models.Exec, error) {
	// 1. Buka koneksi database
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "database query error")
	}
	// 2. Koneksi wajib ditutup setelah selesai
	defer db.Close()

	// 3. Jalankan SELECT by id lalu scan ke struct Exec
	row := db.QueryRow("SELECT "+execSelectCols+" FROM execs WHERE id = ?", id)
	exec, err := scanExec(row)
	if err == sql.ErrNoRows {
		// 4. Id tidak ada → kembalikan sentinel biar handler bisa jawab 404
		return nil, ErrExecNotFound
	} else if err != nil {
		// 5. Error DB lain → log detail & return pesan ramah
		return nil, utils.ErrorHandler(err, "database query error")
	}

	// 6. Berhasil → kembalikan pointer ke exec
	return exec, nil
}

// CreateExec memasukkan SATU exec baru.
// Query ditulis manual: field yang diisi client (first_name, last_name, email,
// username, password, inactive_status, role). Field server-managed dibiarkan
// default DB: user_created_at = CURRENT_TIMESTAMP, sisanya NULL.
// Password yang dikirim HARUS sudah ter-hash oleh handler.
func CreateExec(exec models.Exec) (*models.Exec, error) {
	// 1. Buka koneksi database
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error inserting data into database")
	}
	// 2. Koneksi wajib ditutup setelah selesai
	defer db.Close()

	// 3. Eksekusi INSERT manual dengan semua kolom yang disediakan client
	res, err := db.Exec(
		`INSERT INTO execs
			(first_name, last_name, email, username, password, inactive_status, role)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		exec.FirstName,
		exec.LastName,
		exec.Email,
		exec.Username,
		exec.Password,
		exec.InactiveStatus,
		exec.Role,
	)
	if err != nil {
		return nil, utils.ErrorHandler(err, "error inserting data into database")
	}

	// 4. Ambil auto-increment ID hasil insert
	lastId, err := res.LastInsertId()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error inserting data into database")
	}

	// 5. Set ID ke struct biar bisa dibalas ke client
	exec.ID = int(lastId)

	// 6. Kembalikan exec yang sudah berisi ID
	return &exec, nil
}

// PatchExecByID menggabungkan get + apply + update untuk SATU exec.
// Handler cukup memanggil satu fungsi ini tanpa tahu detail DB.
func PatchExecByID(id int, updates map[string]interface{}) (*models.Exec, error) {
	// 1. Buka koneksi database
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "unable to retrieve exec")
	}
	// 2. Koneksi wajib ditutup setelah selesai
	defer db.Close()

	// 3. Ambil exec yang ada dulu (404 kalau tidak ketemu)
	row := db.QueryRow("SELECT "+execSelectCols+" FROM execs WHERE id = ?", id)
	exec, err := scanExec(row)
	if err == sql.ErrNoRows {
		return nil, ErrExecNotFound
	} else if err != nil {
		return nil, utils.ErrorHandler(err, "unable to retrieve exec")
	}

	// 4. Terapkan field yang dikirim body ke struct (via reflect)
	applyExecUpdates(exec, updates)

	// 5. Simpan perubahan ke database
	_, err = db.Exec(
		`UPDATE execs
		   SET first_name = ?, last_name = ?, email = ?, username = ?,
		       inactive_status = ?, role = ?
		 WHERE id = ?`,
		exec.FirstName,
		exec.LastName,
		exec.Email,
		exec.Username,
		exec.InactiveStatus,
		exec.Role,
		exec.ID,
	)
	if err != nil {
		return nil, utils.ErrorHandler(err, "error updating exec")
	}

	// 6. Kembalikan exec yang sudah ter-update
	return exec, nil
}

// PatchExecs mengupdate BANYAK exec dalam SATU transaksi (all-or-nothing).
// Body: [{"id":1,"first_name":"X"}, {"id":2,"role":"admin"}, ...]
func PatchExecs(updates []map[string]interface{}) ([]models.Exec, error) {
	// 1. Buka koneksi database
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error updating exec")
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
	updatedExecs := make([]models.Exec, 0, len(updates))

	// 6. Loop tiap update
	for _, update := range updates {
		// 7. Ambil id dari map — dikirim sebagai int dari DTO
		id, ok := update["id"].(int)
		if !ok {
			return nil, ErrInvalidExecID
		}

		// 8. SELECT exec by id memakai tx biar dalam transaksi yang sama
		row := tx.QueryRow("SELECT "+execSelectCols+" FROM execs WHERE id = ?", id)
		exec, err := scanExec(row)
		if err == sql.ErrNoRows {
			return nil, ErrExecNotFound
		} else if err != nil {
			return nil, utils.ErrorHandler(err, "unable to retrieve exec")
		}

		// 9. Terapkan field yang dikirim body ke struct (via reflect)
		applyExecUpdates(exec, update)

		// 10. UPDATE memakai tx
		_, err = tx.Exec(
			`UPDATE execs
			   SET first_name = ?, last_name = ?, email = ?, username = ?,
			       inactive_status = ?, role = ?
			 WHERE id = ?`,
			exec.FirstName,
			exec.LastName,
			exec.Email,
			exec.Username,
			exec.InactiveStatus,
			exec.Role,
			exec.ID,
		)
		if err != nil {
			return nil, utils.ErrorHandler(err, "error updating exec")
		}

		// 11. Kumpulkan exec hasil update
		updatedExecs = append(updatedExecs, *exec)
	}

	// 12. Semua sukses → commit (baru beneran tersimpan)
	if err := tx.Commit(); err != nil {
		return nil, utils.ErrorHandler(err, "error committing transaction")
	}

	// 13. Kembalikan list exec yang ter-update
	return updatedExecs, nil
}

// DeleteExecByID menghapus SATU exec berdasarkan id.
// Mengembalikan jumlah baris yang terpengaruh (0 = id tidak ada).
func DeleteExecByID(id int) (int64, error) {
	// 1. Buka koneksi database
	db, err := ConnectDb()
	if err != nil {
		return 0, utils.ErrorHandler(err, "error deleting result")
	}
	// 2. Koneksi wajib ditutup setelah selesai
	defer db.Close()

	// 3. Eksekusi query DELETE by id
	result, err := db.Exec("DELETE FROM execs WHERE id = ?", id)
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

// applyExecUpdates mengubah field struct models.Exec sesuai map updates (via reflect).
// - Cocokkan key body ke json tag field struct
// - Skip "id" biar user gak bisa ubah ID
// - type-assertion aman: value tipe salah di-skip, bukan panic
// - Handle inactive_status sebagai bool
func applyExecUpdates(existingExec *models.Exec, updates map[string]interface{}) {
	// .Elem() karena pointer — biar field bisa di-Set
	execVal := reflect.ValueOf(existingExec).Elem()
	// metadata tipe: nama field + json tag
	execType := execVal.Type()

	for k, v := range updates {
		// id gak boleh di-update lewat reflect — skip biar user gak bisa ubah ID
		if k == "id" {
			continue
		}

		// loop tiap field struct, cari yang json tag-nya cocok sama key body
		for i := 0; i < execVal.NumField(); i++ {
			field := execType.Field(i)
			jsonTag := field.Tag.Get("json")
			// Extract the field name from json tag (e.g., "first_name,omitempty" -> "first_name")
			jsonFieldName := strings.Split(jsonTag, ",")[0]

			if jsonFieldName == k {
				fieldVal := execVal.Field(i)
				// field cuma bisa di-Set kalau exported + lewat pointer (.Elem())
				if fieldVal.CanSet() {
					switch fieldVal.Kind() {
					case reflect.String:
						// type-assertion AMAN: kalau v bukan string, skip, jangan panic
						if str, ok := v.(string); ok {
							fieldVal.SetString(str)
						}
					case reflect.Bool:
						// inactive_status dikirim sebagai bool dari JSON
						if boolVal, ok := v.(bool); ok {
							fieldVal.SetBool(boolVal)
						}
					}
				}
				// udah ketemu field yang cocok, stop loop field, lanjut key berikutnya
				break
			}
		}
	}
}
