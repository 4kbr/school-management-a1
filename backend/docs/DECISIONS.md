# Decision Records

File ini melacak keputusan teknis penting + alasannya (*why*), supaya kalau nanti kode di-refactor ulang, alasan di balik bentuk sekarang tidak hilang.

- Format: ADR singkat, ditambah berurutan (`ADR-00N`) mulai dari sekarang.
- Setiap entry: **Status / Tanggal / Context / Decision / Consequences**.
- Keputusan sesi awal (ADR-001..006) di-backfill dari catatan refactor awal.

---

## ADR-001: Refactor DB operation dari handler ke repository

- **Status:** Accepted
- **Tanggal:** 2026-08-09

### Context
Awalnya `internal/api/handlers/teachers.go` berisi semua query DB inline (SELECT, INSERT, UPDATE, DELETE, transaksi). Handler tebal dan mencampur tanggung jawab HTTP + SQL.

### Decision
Pindahkan seluruh operasi database ke `internal/repositories/sqlconnect/teachers_crud.go`. Handler tidak lagi mengakses DB secara langsung — cukup memanggil function repository.

### Consequences
- Handler lebih tipis dan fokus pada HTTP.
- Tanggung jawab terpisah jelas: repo = data, handler = request/response.
- Perlu batasan yang konsisten tentang apa yang boleh/kagak dilakukan handler vs repo.

---

## ADR-002: Existence-check 404 tetap di handler

- **Status:** Accepted
- **Tanggal:** 2026-08-09

### Context
Kita butuh membedakan "teacher tidak ditemukan" (HTTP 404) dengan error DB lain (500).

### Decision
Repository mengembalikan penanda *not-found* (`sql.ErrNoRows` awal, kemudian sentinel `ErrTeacherNotFound`). Handler yang melakukan cek & memetakan ke status code 404. Repository tetap fokus pada satu operasi CRUD, tidak mengurusi HTTP status.

### Consequences
- Repository murni operasi data; HTTP concern ada di handler.
- Handler sedikit lebih "cerdas" (tahu kapan harus 404).

---

## ADR-003: `ConnectDb()` dipindah ke dalam repository

- **Status:** Accepted
- **Tanggal:** 2026-08-09

### Context
Sebelumnya handler masih membuka dan menutup koneksi sendiri (`ConnectDb()` + `defer db.Close()`). Pola tutorial meng-call koneksi di dalam repository.

### Decision
Setiap function repository sendiri yang membuka koneksi (`ConnectDb()`) dan menutupnya (`defer db.Close()`). Function transaksi (batch) juga menangani `Begin`/`Commit`/`Rollback` penuh di dalamnya.

### Consequences
- Handler jadi murni HTTP (parse request → panggil repo → encode response).
- Setiap panggilan repo membuka 1 koneksi sendiri.
- Trade-off: PUT awalnya memanggil 2 repo function (`GetTeacherByID` + `UpdateTeacher`) = 2 koneksi → digabung jadi `PatchTeacherByID` (1 koneksi) untuk single patch.

---

## ADR-004: Query SQL + sorting/filter pindah ke repository

- **Status:** Accepted
- **Tanggal:** 2026-08-09

### Context
Membangun query (klausa `WHERE`, `ORDER BY`) dianggap concern SQL, bukan concern HTTP.

### Decision
`GetTeachers` menerima `url.Values` dan membangun query penuh di dalamnya. Helper `addSorting`, `addFilters`, `isValidSortField`, `isValidSortOrder` dipindahkan ke repository. Handler hanya meneruskan `r.URL.Query()`.

### Consequences
- Semua string SQL & building query ada di satu tempat (repository).
- Handler hanya parse param → panggil repo → encode response.
- Validasi field/order sorting (anti SQL injection) terkumpul di repo.

---

## ADR-005: Integrasi `utils.ErrorHandler` (Opsi B: sentinel + ErrorHandler)

- **Status:** Accepted
- **Tanggal:** 2026-08-09

### Context
`pkg/utils/error_handler.go` menyediakan `ErrorHandler(err, message)` yang log detail error ke stderr dan mengembalikan error baru berisi `message` saja. Penting: ia memakai `fmt.Errorf("%s", message)` — **tanpa `%w`** — sehingga error asli TIDAK terbawa (tidak bisa di-detect `errors.Is`).

### Decision
- Error umum (500) di repository: `return ... utils.ErrorHandler(err, "<pesan ramah>")` → log detail + return pesan sanitasi (tidak bocor detail DB ke response).
- Kasus 404/400: sentinel `ErrTeacherNotFound` / `ErrInvalidTeacherID` tetap dikembalikan mentah (tidak lewat `ErrorHandler`), supaya `errors.Is` di handler tetap bekerja.
- Handler: `errors.Is` untuk 404/400, selain itu `http.Error(w, err.Error(), 500)`.
- Logging ganda di handler (`fmt.Println`/`log.Println`) dihapus karena `ErrorHandler` sudah log.

### Consequences
- Status code tetap akurat (404/400/500).
- Response tetap aman (hanya pesan ramah, bukan detail teknis).
- Detail error teknis tetap tercatat di stderr.

---

## ADR-006: Dynamic INSERT via struct tags

- **Status:** Accepted
- **Tanggal:** 2026-08-09

### Context
Belajar cara kerja internal ORM: query INSERT dibuat dinamis dari tag struct. `generateInsertQuery` + `getStructValues` membaca tag `db`.

### Decision
Query kolom & placeholder serta nilai field di-generate dari tag `db` pada struct `models.Teacher`. Field `id` (auto-increment) di-skip agar tidak ikut INSERT.

### Consequences & bug yang ditemukan
- Model awalnya cuma punya tag `json`, tidak ada tag `db` → semua tag terbaca kosong, query jadi `INSERT INTO teachers () VALUES ()`.
- Solusi: tambahkan tag `db` pada model (mis. `db:"first_name"`).
- Bug lanjutan: skip-id inkonsisten — `generateInsertQuery` skip `"id"`, `getStructValues` skip `"id,omitempty"` → id ikut ter-append → `expected 5 arguments, got 6`.
- Perbaikan: samakan logika dengan `strings.TrimSuffix(tag, ",omitempty")` lalu skip `== "id"`, sehingga jumlah values = jumlah placeholder.

---

## ADR-007: Data validation best practice (DTO + `go-playground/validator`)

- **Status:** Accepted
- **Tanggal:** 2026-08-09

### Context
Sebelumnya request langsung di-decode ke `models.Teacher` tanpa validasi. Tutorial sesi ini menawarkan 2 cara (manual lalu best practice); dipilih langsung best practice.

### Decision
- Gunakan library `github.com/go-playground/validator/v10` (validator standar de-facto di Go).
- Buat request DTO terpisah di `internal/models/dto/teacher_dto.go`:
  - `CreateTeacherRequest` & `UpdateTeacherRequest` — semua field `required`.
  - `PatchTeacherRequest` — field bertipe pointer (`*string`) + tag `omitempty`, sehingga `nil` = field tidak dikirim (partial update).
- Validator dibuat singleton di `pkg/utils/validation.go` (`validator.New()` dipakai sekali, bukan per-request), dengan helper `ValidateStruct` (single) dan `ValidateSlice` (batch, prefix field pakai `[i].`).
- Error validasi → response 400 JSON terstruktur: `{"status":"error","errors":[{"field":"...","message":"..."}]}`.

### Consequences
- Kontrak request ter-decouple dari model persistensi.
- Type-safe (struct), bukan `map[string]interface{}` yang manual & rawan typo.
- Bisa beda rule per endpoint (Create/PUT full, PATCH partial).
- Repository PATCH tetap memakai `map` + `applyUpdates`; DTO dikonversi lewat helper `patchReqToMap` (hanya field non-nil yang masuk).

---

## ADR-008: Decode type-mismatch sebagai error terstruktur (`DecodeJSON`)

- **Status:** Accepted
- **Tanggal:** 2026-08-09

### Context
Kalau client mengirim tipe yang salah (mis. `class` diisi angka padahal field `string`), `encoding/json` gagal di tahap *decode* dan mengembalikan `*json.UnmarshalTypeError`. Sebelumnya ini di-handle sebagai teks polos `invalid request body`, tidak konsisten dengan error validasi yang terstruktur.

### Decision
- Tambahkan `utils.DecodeJSON(body, dst)` yang:
  - decode sukses → `nil`;
  - deteksi `*json.UnmarshalTypeError` → `ValidationError` terstruktur (field + pesan seperti `must be string, got number`);
  - error JSON lain → `ValidationError` generik.
- Semua handler DTO (POST/PUT/PATCH batch/PATCH single) memakai `DecodeJSON`.

### Consequences
- Error decode & error validasi jadi konsisten (sama-sama 400 JSON terstruktur).
- Client mendapat pesan yang jelas per field, bukan teks polos.

### Catatan perbaikan bug yang ditemukan saat implementasi
- `toSnake("ID")` diperbaiki jadi `"id"` (bukan `"i_d"`) — underscore hanya ditambah sebelum huruf besar yang didahului huruf kecil (mengenali akronim).
- Single PATCH: `req.ID` di-set dari URL path (bukan body), karena DTO shared memakai `required` pada `ID`.
- Repo `PatchTeachers` membaca `update["id"].(int)` (nilai dari DTO), bukan `.(float64)` (nilai dari JSON decode).

---

## ADR-009: Dokumentasi OpenAPI + Swagger UI (`swaggo/swag`)

- **Status:** Accepted
- **Tanggal:** 2026-08-10

### Context
Ganti Postman manual (mis. saat pindah device) butuh dokumentasi API yang self-contained dan interaktif. Dipilih `swaggo/swag` karena itu standar de-facto & common use di Go: spec OpenAPI di-generate otomatis dari anotasi kode (tidak perlu nulis YAML manual yang gampang tidak sinkron).

### Decision
- Tambahkan dependency: `github.com/swaggo/swag`, `github.com/swaggo/http-swagger/v2`, `github.com/swaggo/files/v2`, plus CLI `swag` (untuk regen).
- Info global API ditaruh di `cmd/api/main.go` via anotasi `@title`, `@version`, `@description`, `@BasePath /`, `@schemes https`.
- Setiap handler (8 endpoint teachers + stubs root/students/execs) diberi anotasi `@Summary/@Tags/@Param/@Success/@Router` dsb.
- Tambah response DTO bernama (`dto.TeacherListResponse`, `TeacherDeleteResponse`, `TeacherDeleteOneResponse`) biar skema respons terdokumentasi rapi.
- Route Swagger UI diregister di `router.go`: `mux.Handle("GET /swagger/", httpSwagger.WrapHandler)` + `import _ "school-management-api/docs"`.
- Generate spec via `swag init -g cmd/api/main.go -o docs` → `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`. Ditambah target `make swagger`.

### Consequences
- UI interaktif di `https://localhost:3000/swagger/index.html` — bisa "Try it out" langsung.
- `router.go` meng-import `_ "school-management-api/docs"`, jadi **`docs/` wajib di-generate (`make swagger`) sebelum build**. Folder `docs/` di-commit supaya clone baru langsung bisa build.
- Setiap perubahan route/body/response harus diikuti `make swagger` biar spec tetap sinkron.
- Akses lewat HTTPS karena server memakai `ListenAndServeTLS`.
