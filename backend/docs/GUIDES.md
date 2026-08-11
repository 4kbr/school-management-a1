# Panduan: Init, Config, dan Wiring Program

Dokumen ini menjelaskan bagaimana program backend ini di-init dari nol, bagaimana
konfigurasi dibaca, dan bagaimana semua komponen di-wire agar program bisa jalan.
Detail command eksekusi (cert, migrate, docker, dsb.) ada di [`command.md`](./command.md).

---

## 1. Alur boot program (ringkasan)

Semua dimulai dari entry point `cmd/api/main.go`. Urutan eksekusinya:

```
main() ──► godotenv.Load()                    # baca .env → env vars
    │
    ├──► sqlconnect.ConnectDb()               # cek koneksi DB (buka-tutup per request)
    │
    ├──► router.Router()                      # http.NewServeMux() + semua route
    │
    ├──► utils.ApplyMiddlewares(router, mw.SecurityHeaders)   # lapisi middleware
    │
    ├──► http.Server{ Addr, Handler, TLSConfig }             # server HTTP
    │
    └──► server.ListenAndServeTLS(cert, key)  # jalan, wajib HTTPS
```

Ringkasan tanggung jawab tiap bagian:

| Bagian | File | Tanggung jawab |
| --- | --- | --- |
| Entry point | `cmd/api/main.go` | Load env, cek DB, rakit router + middleware, start server |
| Router | `internal/api/router/router.go` | Mendaftarkan semua route ke `http.ServeMux` |
| Handler | `internal/api/handlers/*.go` | Terima HTTP request, panggil repository, balas response |
| Repository | `internal/repositories/sqlconnect/*.go` | Seluruh query DB + `ConnectDb()` |
| Middleware | `internal/api/middlewares/*.go` | Header keamanan, CORS, dsb. (bisa dilapis) |
| Utils | `pkg/utils/*.go` | Helper: apply middleware, error handler, validasi |
| Konfigurasi | `.env` (via godotenv) | Port, TLS cert, kredensial DB |

---

## 2. Prasyarat & tooling

| Tool | Fungsi | Cek versi |
| --- | --- | --- |
| Go (min. 1.22, project pakai 1.25) | Bahasa & runtime | `go version` |
| Docker + Docker Compose | Menjalankan MariaDB | `docker --version` |
| `air` | Hot-reload saat development | `air -v` |
| `swag` | Generate dokumentasi OpenAPI/Swagger | `swag --version` |
| `migrate` (golang-migrate) | Menjalankan migrasi DB | `migrate -version` |

Semua command di bawah dijalankan **dari folder `backend/`** (kecuali disebut lain).

---

## 3. Init project dari nol

> Bagian ini berguna kalau mulai project baru. Kalau repo sudah ada, cukup
> `go mod download` lalu lanjut ke section 4.

### 3.1 Inisialisasi modul

```bash
go mod init school-management-api
```

### 3.2 Struktur direktori

Buat struktur berikut (ini pola project ini):

```
backend/
├── cmd/api/main.go            # entry point & wiring DI
├── internal/
│   ├── api/
│   │   ├── handlers/          # HTTP handler + DTO mapping
│   │   ├── middlewares/       # security headers, cors, dsb.
│   │   └── router/            # routing setup
│   ├── models/                # entity / business model
│   │   └── dto/               # request/response object
│   └── repositories/
│       └── sqlconnect/        # query DB + ConnectDb()
├── migrations/                # file migrasi (golang-migrate)
├── docs/                      # dokumentasi + swagger generated
├── pkg/
│   └── utils/                 # helper reusable
├── .env.example               # template env (di-commit)
├── .env                       # env asli (di-ignore)
├── Makefile                   # otomasi command
└── go.mod / go.sum
```

### 3.3 Dependensi

```bash
# MySQL/MariaDB driver (wajib, database/sql tidak punya driver bawaan)
go get github.com/go-sql-driver/mysql

# Load .env ke env vars
go get github.com/joho/godotenv

# Validasi request body
go get github.com/go-playground/validator/v10

# Dokumentasi OpenAPI/Swagger
go get github.com/swaggo/swag
go get github.com/swaggo/http-swagger/v2
go get github.com/swaggo/files/v2
```

CLI tambahan (install sekali, bukan dependency Go):
- `air` — hot reload: `go install github.com/air-verse/air@latest`
- `swag` — generator Swagger: `go install github.com/swaggo/swag/cmd/swag@latest`
- `migrate` — golang-migrate CLI (pakai package manager masing-masing OS)

---

## 4. Konfigurasi (.env)

### 4.1 Buat file env

```bash
cp .env.example .env
```

`.env` **tidak di-commit** (lihat `.gitignore`), sedangkan `.env.example` **di-commit**
sebagai template — supaya setup reproducible.

### 4.2 Variabel yang tersedia

| Variabel | Default | Dipakai di | Keterangan |
| --- | --- | --- | --- |
| `API_PORT` | `":3000"` | `main.go` | Port server (termasuk prefix `:`) |
| `TLS_CERT` | `"cert.pem"` | `main.go` | Path cert (relatif dari `backend/`) |
| `TLS_KEY` | `"key.pem"` | `main.go` | Path private key |
| `DB_USER` | `"app"` | `sql_config.go` | User DB |
| `DB_PASSWORD` | `"app_password"` | `sql_config.go` | Password DB |
| `DB_NAME` | `"school_management"` | `sql_config.go` | Nama database |
| `DB_PORT` | `3306` | `sql_config.go` | Port DB |
| `DB_HOST` | `127.0.0.1` | `sql_config.go` | Host DB |

Nilai memakai tanda kutip (`"app"`) — ini sengaja: aplikasi membaca via
`godotenv.Load()` yang otomatis membuang kutip. Makefile juga membuang kutip
dengan `$(subst ",,$(VAR))` saat membangun URL migrasi.

### 4.3 Cara env dibaca

Aplikasi tidak memakai struct config terpusat; env dibaca langsung via `os.Getenv`:

- `main.go` → `API_PORT`, `TLS_CERT`, `TLS_KEY`
- `internal/repositories/sqlconnect/sql_config.go` → `DB_USER`, `DB_PASSWORD`,
  `DB_NAME`, `DB_PORT`, `DB_HOST`

```go
// cmd/api/main.go
err := godotenv.Load()      // baca .env di working directory → env vars
if err != nil {
    panic(err)              // program TIDAK jalan kalau .env tidak ada
}
```

> Prinsip yang dipakai: **failing fast** — kalau `.env` tidak ada, langsung `panic`,
> biar kesalahan konfigurasi ketahuan sejak awal, bukan saat runtime.

---

## 5. Wiring program

### 5.1 `cmd/api/main.go` — entry point

Urutan di dalam `main()`:

1. **Load env**: `godotenv.Load()` → panic kalau `.env` tidak ada.
2. **Cek koneksi DB**: panggil `sqlconnect.ConnectDb()` (sekali saja, cek konfigurasi
   benar; koneksi sesungguhnya dibuka-tutup per request di repository).
3. **Baca port & TLS**: `os.Getenv("API_PORT")`, `os.Getenv("TLS_CERT")`,
   `os.Getenv("TLS_KEY")`. TLS di-minimunkan ke `tls.VersionTLS12`.
4. **Bangun router**: `router.Router()` → `http.ServeMux`.
5. **Lapis middleware**: `utils.ApplyMiddlewares(router, mw.SecurityHeaders)`.
   Saat ini baru `SecurityHeaders` yang aktif; middleware lain (CORS, RateLimiter,
   HPP, Compression, ResponseTime) masih di-comment — tinggal diaktifkan nanti
   dengan menambahkannya ke `ApplyMiddlewares`.
6. **Start server**: `http.Server{Addr, Handler, TLSConfig}` lalu
   `server.ListenAndServeTLS(cert, key)`.

```go
router := router.Router()
secureMux := utils.ApplyMiddlewares(router, mw.SecurityHeaders)

server := &http.Server{
    Addr:      port,
    Handler:   secureMux,
    TLSConfig: tlsConfig,
}
err = server.ListenAndServeTLS(cert, key)
```

### 5.2 `internal/repositories/sqlconnect/sql_config.go` — koneksi DB

`ConnectDb()` membaca `DB_*` dari env, menyusun connection string, lalu `sql.Open`.
Driver MySQL di-import blank (`_`) untuk registrasi side-effect.

```go
func ConnectDb() (*sql.DB, error) {
    dbUser := os.Getenv("DB_USER")
    // ...
    connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
        dbUser, dbPassword, dbHost, dbPort, dbName)
    db, err := sql.Open("mysql", connectionString)
    if err != nil {
        return nil, err
    }
    return db, nil
}
```

> Setiap function repository memanggil `ConnectDb()` sendiri dan `defer db.Close()`.
> Jadi tidak ada satu `*sql.DB` global yang di-share antar request.

### 5.3 `internal/api/router/router.go` — daftar route

`Router()` membuat `http.NewServeMux()` lalu mendaftarkan route dengan
**method routing** Go 1.22 (`"GET /teachers"`, `"POST /teachers/{id}"`, dsb.).
Import blank `_ "school-management-api/docs"` diperlukan agar Swagger UI jalan
(`docs/` harus di-generate via `make swagger`).

```go
mux.HandleFunc("GET /teachers", handlers.GetTeachersHandler)
mux.HandleFunc("GET /teachers/{id}", handlers.GetOneTeacherByIdHandler)
mux.Handle("GET /swagger/", httpSwagger.WrapHandler)
```

### 5.4 `pkg/utils/middlewares_util.go` — cara melapis middleware

`Middleware` adalah fungsi yang menerima `http.Handler` dan mengembalikan
`http.Handler`. `ApplyMiddlewares` melapisnya secara berurutan.

```go
type Middleware func(http.Handler) http.Handler

func ApplyMiddlewares(handler http.Handler, middlewares ...Middleware) http.Handler {
    for _, middleware := range middlewares {
        handler = middleware(handler)
    }
    return handler
}
```

### 5.5 `internal/api/middlewares/security_headers.go` — contoh middleware

Setiap middleware berbentuk `func Nama(next http.Handler) http.Handler`. Logika
berada di dalam `http.HandlerFunc`: set header/side-effect sebelum
`next.ServeHTTP(w, r)`, dan bisa ada aksi setelahnya.

```go
func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        // ...
        next.ServeHTTP(w, r)
    })
}
```

---

## 6. Menjalankan program

```bash
# 1) Start database (MariaDB via Docker), tunggu sampai healthy
make dockerdev-up

# 2) Apply migrasi
make migrate-up

# 3) Jalankan server
make dev        # hot-reload (air) — untuk development
# atau
make run        # go run ./cmd/api/ — tanpa hot-reload

# 4) Test
curl -k https://localhost:3000/teachers
```

Cek dokumentasi API interaktif di `https://localhost:3000/swagger/index.html`.
Detail tentang cert, migrate, dan command lain: lihat [`command.md`](./command.md).

---

## 7. Referensi file kunci

| File | Peran dalam wiring |
| --- | --- |
| `cmd/api/main.go` | Merakit semua komponen (env → DB → router → middleware → server) |
| `internal/repositories/sqlconnect/sql_config.go` | `ConnectDb()` — buat koneksi dari env |
| `internal/repositories/sqlconnect/teachers_crud.go` | Contoh repository (query DB) |
| `internal/repositories/sqlconnect/students_crud.go` | Repository students |
| `internal/api/router/router.go` | Pendaftaran route |
| `internal/api/handlers/*.go` | Handler per resource |
| `internal/api/middlewares/security_headers.go` | Contoh middleware |
| `pkg/utils/middlewares_util.go` | `ApplyMiddlewares` |
| `pkg/utils/error_handler.go` | `ErrorHandler` — sanitasi error |
| `pkg/utils/validation.go` | Validasi + `DecodeJSON` |
| `.env.example` | Template konfigurasi |

---

## 8. Menambah API baru (langkah demi langkah)

Alur ini mengikuti pola yang sudah dipakai di resource `teachers` dan `students`.
Contoh di bawah memakai resource fiktif **`execs`** (satu entitas, tanpa relasi
khusus) — ikuti pola yang persis. Jumlah langkah: **8**.

```
Model → Migration → DTO → Repository → Handler → Router → Swagger → Verifikasi
```

### 8.1 Model — `internal/models/<entity>.go`

Buat struct entity. **Wajib ada tag `db`** (dipakai helper query dinamis) dan
tag `json` (format response).

```go
package models

type Exec struct {
	ID        int    `json:"id,omitempty" db:"id"`
	FirstName string `json:"first_name,omitempty" db:"first_name"`
	LastName  string `json:"last_name,omitempty" db:"last_name"`
	Email     string `json:"email,omitempty" db:"email"`
	Class     string `json:"class,omitempty" db:"class"`
}
```

Catatan:
- Field `id` tetap diberi tag `db:"id"` tapi **tidak ikut INSERT** (helper otomatis
  mengecualikan field bernama `id`).
- Kolom yang tidak dipakai di aplikasi (mis. `created_at`, `updated_at`) tidak
  perlu dimasukkan ke struct — cukup ada di migration.

### 8.2 Migration — `migrations/00000N_create_<entity>.{up,down}.sql`

Satu pasang file: `.up.sql` (membuat tabel) dan `.down.sql` (rollback).
Penomoran harus berurutan dari migrasi terakhir.

```sql
-- migrations/000003_create_execs.up.sql
CREATE TABLE IF NOT EXISTS execs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    first_name VARCHAR(255) NOT NULL,
    last_name  VARCHAR(255) NOT NULL,
    email      VARCHAR(255) UNIQUE NOT NULL,
    class      VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_email (email)
) AUTO_INCREMENT = 100;
```

```sql
-- migrations/000003_create_execs.down.sql
DROP TABLE IF EXISTS execs;
```

Apply / rollback / cek versi:

```bash
make migrate-up
make migrate-down   # hanya kalau perlu rollback
make migrate-version
```

> Kalau tabel punya `FOREIGN KEY` (seperti `students.class → teachers.class`),
> tambahkan validasi di repository (langkah 8.4) biar error-nya bersih, bukan
> error FK mentah — lihat `classExists`/`ErrClassNotFound` di `students_crud.go`.

### 8.3 DTO — `internal/models/dto/<entity>_dto.go` + `<entity>_response.go`

Buat file request DTO baru (satu struct per method), mengikuti `teacher_dto.go` /
`student_dto.go`. Pola validasi: create & update **full required**, patch pakai
**pointer** (`*string`) agar field nil = tidak dikirim client.

```go
package dto

type CreateExecRequest struct {
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,max=100"`
	Email     string `json:"email" validate:"required,email"`
	Class     string `json:"class" validate:"required,max=50"`
}

// UpdateExecRequest: sama seperti Create, untuk PUT (full replace).

type PatchExecRequest struct {
	ID        int     `json:"id" validate:"required"`
	FirstName *string `json:"first_name" validate:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name" validate:"omitempty,max=100"`
	Email     *string `json:"email" validate:"omitempty,email"`
	Class     *string `json:"class" validate:"omitempty,max=50"`
}
```

Tambah response DTO di file `internal/models/dto/<entity>_response.go`
(per domain, biar Swagger menampilkan skema bernama):

```go
type ExecListResponse struct {
	Status string         `json:"status"`
	Count  int            `json:"count"`
	Data   []models.Exec  `json:"data"`
}

type ExecDeleteResponse struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
	IDs    []int  `json:"ids"`
}

type ExecDeleteOneResponse struct {
	Status string `json:"status"`
	ID     int    `json:"id"`
}
```

### 8.4 Repository — `internal/repositories/sqlconnect/<entity>_crud.go`

Salin pola `students_crud.go` dan ganti menjadi `execs`. Yang harus ada:

1. **Sentinel error** di bagian atas file:
   ```go
   var ErrExecNotFound = errors.New("exec not found")
   var ErrInvalidExecID = errors.New("invalid exec id in update")
   ```
2. **8 fungsi** dengan pola sama:
   | Fungsi | Metode |
   | --- | --- |
   | `GetExecs(values url.Values)` | GET list (filter + sort) |
   | `GetExecByID(id)` | GET by id (404 via `ErrExecNotFound`) |
   | `CreateExecs([]models.Exec)` | POST batch (isi ID dari `LastInsertId`) |
   | `UpdateExec(exec)` | PUT (full update) |
   | `PatchExecByID(id, updates)` | PATCH single (get + apply + update) |
   | `PatchExecs([]map[string]interface{})` | PATCH batch (transaksi) |
   | `DeleteExecByID(id)` | DELETE single (return `rowsAffected`) |
   | `DeleteExecs([]int)` | DELETE batch (transaksi, all-or-nothing) |
3. **Helper duplikat** yang disesuaikan nama + kolom: `addExecFilters`,
   `addExecSorting`, `isValidExecSortField`, `generateExecInsertQuery`,
   `getExecStructValues`, `applyExecUpdates`.
   - Tabel pada query: `execs`.
   - Field filter/sort: sesuaikan kolom entity (`first_name`, `last_name`,
     `email`, `class`).
   - `applyExecUpdates`: skip key `"id"`, type-assertion aman per `reflect.Kind`.

> Pola ini sengaja **diduplikasi** per entity, bukan dibuat generic — sesuai
> prinsip YAGNI & explicit > implicit (lihat ADR-010 di `DECISIONS.md`).

### 8.5 Handler — `internal/api/handlers/<entity>.go`

Salin pola `students.go`, ganti nama + model/DTO + fungsi repo. Perlu:

1. **8 handler** dengan anotasi Swagger di atas tiap function (`@Summary`,
   `@Tags`, `@Param`, `@Success`, `@Router`). Contoh:
   ```go
   // GetExecsHandler godoc
   // @Summary      List execs
   // @Tags         execs
   // @Produce      json
   // @Success      200 {object} dto.ExecListResponse
   // @Router       /execs [get]
   func GetExecsHandler(w http.ResponseWriter, r *http.Request) { ... }
   ```
2. **Helper mapping** di bawah file: `createExecReqToModel`,
   `updateExecReqToModel`, `patchExecReqToMap`.
3. **Error mapping** dengan `errors.Is` pada sentinel:
   - `ErrExecNotFound` → `404 "exec not found"`
   - `ErrInvalidExecID` → `400`
   - sentinel khusus (mis. FK) → `400` dengan pesan sesuai
   - selain itu → `500`
4. `writeValidationError` **tidak perlu dibuat ulang** — sudah ada di package
   `handlers` (dipakai semua resource).
5. Pola single PATCH: `req.ID = id` dari `r.PathValue("id")` sebelum validasi.

### 8.6 Router — `internal/api/router/router.go`

Daftarkan route dengan method routing Go 1.22, lalu hapus handler placeholder lama:

```go
mux.HandleFunc("GET /execs", handlers.GetExecsHandler)
mux.HandleFunc("POST /execs", handlers.AddExecHandler)
mux.HandleFunc("PATCH /execs", handlers.PatchExecsHandler)
mux.HandleFunc("DELETE /execs", handlers.DeleteExecsHandler)

mux.HandleFunc("GET /execs/{id}", handlers.GetOneExecByIdHandler)
mux.HandleFunc("PUT /execs/{id}", handlers.UpdateOneExecByIdHandler)
mux.HandleFunc("PATCH /execs/{id}", handlers.PatchOneExecByIdHandler)
mux.HandleFunc("DELETE /execs/{id}", handlers.DeleteOneExecByIdHandler)
```

### 8.7 Swagger

```bash
make swagger
```

Wajib dijalankan setelah ada perubahan route/handler — `router.go` meng-import
`_ "school-management-api/docs"`, jadi `docs/` harus selalu ter-generate sebelum
build (detail: ADR-009 + section Swagger di `command.md`).

### 8.8 Verifikasi

```bash
go build ./... && go vet ./...
make dev

# cek route baru muncul di spec OpenAPI
curl -sk https://localhost:3000/swagger/doc.json | grep -i execs

# test CRUD (sesuaikan data valid, termasuk class yang sudah ada di teachers kalau ada FK)
curl -sk -X POST https://localhost:3000/execs \
  -H "Content-Type: application/json" \
  -d '[{"first_name":"A","last_name":"B","email":"a@b.com","class":"9-A"}]'
curl -sk https://localhost:3000/execs
```

### Checklist

- [ ] Model punya tag `json` + `db`
- [ ] Migration `.up.sql` + `.down.sql` berurutan, sudah `make migrate-up`
- [ ] DTO request + response lengkap, tag `validate` sesuai (patch pakai pointer)
- [ ] Repository punya sentinel error + 8 fungsi + helper duplikat
- [ ] Handler: anotasi Swagger + error mapping via `errors.Is`
- [ ] Router method-routing + placeholder lama dihapus
- [ ] `make swagger` sudah jalan, spec update
- [ ] `go build ./... && go vet ./...` lulus
- [ ] Test manual CRUD via curl / Swagger UI
