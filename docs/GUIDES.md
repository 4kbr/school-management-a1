# Guides

Dokumen ini berisi step-by-step untuk tugas umum selama development.

---

## 1. Menambah Fitur Baru (CRUD Example: Teachers)

Urutan bikin fitur dari atas ke bawah:

```
domain (entity)
  → repository (interface)
    → repository/mysql (implementation)
      → service (business logic)
        → handler (HTTP) + DTO
          → router (routing registration)
            → main.go (wiring DI)
```

### Langkah detail:

#### a. Domain Model — `internal/domain/teacher.go`

```go
package domain

type Teacher struct {
    ID        string
    Name      string
    Email     string
    Phone     string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

#### b. Repository Interface — `internal/repository/teacher.go`

```go
package repository

import (
    "context"
    "school-management-api/internal/domain"
)

type TeacherRepository interface {
    FindAll(ctx context.Context) ([]domain.Teacher, error)
    FindByID(ctx context.Context, id string) (*domain.Teacher, error)
    Create(ctx context.Context, teacher *domain.Teacher) error
    Update(ctx context.Context, teacher *domain.Teacher) error
    Delete(ctx context.Context, id string) error
}
```

#### c. Repository Implementation — `internal/repository/mysql/teacher.go`

```go
package mysql

import (
    "context"
    "database/sql"
    "school-management-api/internal/domain"
)

type teacherRepository struct {
    db *sql.DB
}

func NewTeacherRepository(db *sql.DB) *teacherRepository {
    return &teacherRepository{db: db}
}

func (r *teacherRepository) FindAll(ctx context.Context) ([]domain.Teacher, error) {
    // implementation
}
```

#### d. Service — `internal/service/teacher.go`

```go
package service

import (
    "context"
    "school-management-api/internal/domain"
    "school-management-api/internal/repository"
)

type TeacherService struct {
    repo repository.TeacherRepository
}

func NewTeacherService(repo repository.TeacherRepository) *TeacherService {
    return &TeacherService{repo: repo}
}

func (s *TeacherService) GetAllTeachers(ctx context.Context) ([]domain.Teacher, error) {
    return s.repo.FindAll(ctx)
}
```

#### e. Handler — `internal/handler/teacher.go`

```go
package handler

import (
    "encoding/json"
    "net/http"
    "school-management-api/internal/service"
)

type TeacherHandler struct {
    svc *service.TeacherService
}

func NewTeacherHandler(svc *service.TeacherService) *TeacherHandler {
    return &TeacherHandler{svc: svc}
}

func (h *TeacherHandler) HandleGetAll(w http.ResponseWriter, r *http.Request) {
    teachers, err := h.svc.GetAllTeachers(r.Context())
    if err != nil {
        // response error JSON
        return
    }
    json.NewEncoder(w).Encode(teachers)
}
```

#### f. Router — `internal/router/router.go`

```go
package router

import (
    "net/http"
    "school-management-api/internal/handler"
)

func New(h *handler.TeacherHandler) http.Handler {
    mux := http.NewServeMux()

    mux.HandleFunc("GET /teachers", h.HandleGetAll)
    mux.HandleFunc("POST /teachers", h.HandleCreate)
    mux.HandleFunc("GET /teachers/{id}", h.HandleGetByID)
    mux.HandleFunc("PUT /teachers/{id}", h.HandleUpdate)
    mux.HandleFunc("DELETE /teachers/{id}", h.HandleDelete)

    return mux
}
```

#### g. Wiring — `cmd/api/main.go`

```go
db := // init koneksi db
teacherRepo := mysql.NewTeacherRepository(db)
teacherSvc := service.NewTeacherService(teacherRepo)
teacherHdl := handler.NewTeacherHandler(teacherSvc)
r := router.New(teacherHdl)

http.ListenAndServe(":3000", r)
```

---

## 2. Dependency Injection / Wiring Pattern

Manual DI (no framework). Constructor injection via `main.go`:

```
db → repo → service → handler → router → server
```

Setiap layer terima dependency-nya lewat constructor. Kalau test tinggal ganti implementasi interface-nya.

---

## 3. Convention Penamaan

| Elemen         | Convention              | Contoh                                      |
| -------------- | ----------------------- | ------------------------------------------- |
| File           | `snake_case.go`         | `teacher_handler.go`                        |
| Interface      | `TeacherRepository`     | —                                           |
| Struct impl    | lowercase + nama domain | `teacherRepository`, `teacherService`       |
| Constructor    | `New` prefix            | `NewTeacherRepository`, `NewTeacherService` |
| Handler method | `Handle` prefix         | `HandleGetAll`, `HandleCreate`              |
| Request DTO    | `CreateTeacherRequest`  | —                                           |
| Response DTO   | `TeacherResponse`       | —                                           |

---

## 4. Error Handling Pattern

### Sentinel errors di domain

```go
var ErrTeacherNotFound = errors.New("teacher not found")
var ErrDuplicateEmail = errors.New("duplicate email")
```

### Wrap di layer service

```go
if err != nil {
    return fmt.Errorf("service.Teacher.GetAll: %w", err)
}
```

### Map error ke HTTP di handler

```go
switch {
case errors.Is(err, domain.ErrTeacherNotFound):
    http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
default:
    http.Error(w, `{"error":"internal"}`, http.StatusInternalServerError)
}
```

---

## 5. Middleware Pattern

Middleware adalah `func(next http.Handler) http.Handler`.

```go
func RequestLogger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        slog.Info("request", "method", r.Method, "path", r.URL.Path)
        next.ServeHTTP(w, r)
    })
}
```

Gunakan di router:

```go
mux = middleware.RequestLogger(mux)
mux = middleware.Recovery(mux)
mux = middleware.CORS(mux)
```

---

## 6. Testing Pattern

### Table-driven test untuk service

```go
func TestTeacherService_GetAll(t *testing.T) {
    tests := []struct {
        name    string
        mock    func(repo *mocks.TeacherRepository)
        wantLen int
        wantErr bool
    }{
        // cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // arrange
            // act
            // assert
        })
    }
}
```

### Handler test dengan httptest

```go
func TestTeacherHandler_HandleGetAll(t *testing.T) {
    req := httptest.NewRequest("GET", "/teachers", nil)
    rec := httptest.NewRecorder()
    handler.HandleGetAll(rec, req)
    assert.Equal(t, http.StatusOK, rec.Code)
}
```

---

## 7. Adding Feature Checklist

Copy-paste dan centang tiap kali nambah fitur:

```
□ domain model (internal/domain/)
□ repository interface (internal/repository/)
□ repository implementation (internal/repository/mysql/)
□ service layer (internal/service/)
□ handler + DTO (internal/handler/)
□ routing (internal/router/)
□ wiring di main.go
□ unit test (service + handler)
```

---

## 8. Graceful Shutdown Pattern

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

server := &http.Server{Addr: ":3000", Handler: r}

go func() {
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatal(err)
    }
}()

<-ctx.Done()
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
server.Shutdown(shutdownCtx)
```

---

## 9. Live Reload with `air`

`air` — tool yang otomatis rebuild & restart server saat ada file berubah.

### Install

```bash
go install github.com/air-verse/air@latest
```

### Init config di `backend/`

```bash
cd backend
air init
```

Bikin file `.air.toml` di `backend/`. Isi minimal:

```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  bin = "tmp/api"
  cmd = "go build -o tmp/api ./cmd/api/"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_error = true

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
```

### Run

```bash
cd backend
air
```

Server otomatis restart tiap kali `.go` file berubah.

### Tips

- File `.air.toml` path relative dari `backend/`, sesuaikan `cmd` dan `bin` dengan struktur project.
- Kalau ada folder baru yang perlu di-watch, tambah ke `include_dir`.
- `kill_delay` bisa naikin kalau graceful shutdown butuh waktu.
- Tambah `backend/tmp/` ke `.gitignore`.

---

## 10. Defer — Penggunaan & Best Practice

### Apa itu `defer`

`defer` menjadwalkan eksekusi function **setelah fungsi sekitarnya selesai** (return, panic, atau selesai normal). Biasanya untuk cleanup.

### Sifat penting

- Args di-**evaluasi di tempat** `defer` ditulis, bukan saat dieksekusi.
- Dieksekusi LIFO (_last in, first out_) — kebalikan urutan defer.
- Tetap jalan meskipun fungsi panic.

### Use case di project real

#### a. Tutup request body (`io.ReadCloser`)

```go
body, err := io.ReadAll(r.Body)
if err != nil {
    http.Error(w, "failed to read body", http.StatusBadRequest)
    return
}
defer r.Body.Close()
```

PENTING: `r.Body` otomatis di-close Go setelah handler return (Go 1.25+), tapi tetap eksplisit safest practice.

#### b. Tutup rows hasil SQL query

```go
rows, err := db.QueryContext(ctx, "SELECT * FROM teachers")
if err != nil {
    return fmt.Errorf("query: %w", err)
}
defer rows.Close()

for rows.Next() {
    // scan
}
```

Tanpa `defer rows.Close()`, connection pool bocor — koneksi DB gak balik ke pool.

#### c. Rollback transaction kalau error

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return fmt.Errorf("begin tx: %w", err)
}
defer tx.Rollback() // no-op kalau sudah Commit

if err := doSomething(); err != nil {
    return err // otomatis rollback via defer
}

return tx.Commit()
```

Pattern Standar Go: `defer tx.Rollback()` — kalau `Commit` sukses, `Rollback` jadi no-op. Kalau error, rollback otomatis.

#### d. Tutup koneksi database pas shutdown

```go
func main() {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // ... server jalan
}
```

#### e. Release mutex

```go
mu.Lock()
defer mu.Unlock()
```

#### f. Cancel context

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

---

### Perangkap (common mistakes)

| Mistake                                                            | Kenapa salah                                             | Benar                                                 |
| ------------------------------------------------------------------ | -------------------------------------------------------- | ----------------------------------------------------- |
| `defer` di dalam loop                                              | Numpuk n gak release sampe fungsi selesai                | Eksekusi langsung, jangan defer                       |
| `defer rows.Close()` setelah `for rows.Next()`                     | Rows ditutup sebelum dipake (kalau ada return)           | Taruh `defer` langsung setelah `db.Query`             |
| `defer f()` dengan arg berubah                                     | Args di-evaluasi saat defer ditulis, bukan saat eksekusi | Passing pointer atau closure jika perlu nilai terbaru |
| `defer` di `main()` untuk resource yang harus release sebelum exit | Cocok kok, tapi pastikan tau timing-nya                  | OK asal sesuai urutan                                 |

---

## 11. Linting & Formatting (gofmt, go vet, golangci-lint)

Buat yang familiar dengan TypeScript: Go punya padanan ESLint & Prettier, dan yang keren sebagian udah jadi bagian dari ecosystem standar — bukan dependency eksternal.

### 11.1. Perbandingan dengan TypeScript

| TypeScript       | Go              | Keterangan                                                     |
| ---------------- | --------------- | -------------------------------------------------------------- |
| Prettier         | `gofmt`         | Formatter resmi bawaan Go. Zero config, otomatis jalan di save |
| ESLint (dasar)   | `go vet`        | Static analysis bawaan Go. Deteksi bug pola umum               |
| ESLint (lengkap) | `golangci-lint` | Aggregator puluhan linter. Configurable, mirip plugin ESLint   |
| —                | `goimports`     | Formatter + auto-manage import (sort, hapus unused)            |

Poin penting yang beda dari TypeScript:

- **`gofmt` gak bisa dikonfigurasi** — beda dengan Prettier yang punya `tabWidth`, `singleQuote`, dll. Ini intentional: konsistensi format seluruh ecosystem lebih penting daripada preferensi pribadi.
- **`go vet`** itu linter, bukan formatter — memeriksa runtime correctness, bukan style.
- **`golangci-lint`** (eksternal, community) — kalau mau kontrol level ala ESLint rules, ini yang paling deket. Umum dipakai di CI.
- Gak ada config file yang wajib — `gofmt` + `go vet` cukup jalan sebagai baseline di project mana pun.

### 11.2. `gofmt` — formatter bawaan (si "Prettier"-nya Go)

**Cara kerja:** `gofmt` parse source code jadi **AST** (Abstract Syntax Tree) — bukan operasi regex — lalu mengeluarkan output kanonikal. Makanya konsisten di semua kode Go, termasuk yang ditulis orang lain. Kalau syntax error, dia menolak output (tidak memformat setengah-setengah).

**Perintah dasar:**

```bash
gofmt -l .          # list file yang formatnya belum sesuai (dry run)
gofmt -d file.go    # tampilkan diff format sebelum/sesudah (review dulu)
gofmt -w file.go    # write — tulis langsung hasil format ke file
go fmt ./...        # alias `gofmt -l -w` untuk semua package di folder
```

**Editor integration:** extension Go di VS Code/Goland otomatis jalanin `gofmt` tiap save. Kalau mau cek tanpa editor:

```bash
gofmt -l ./...
```

Kalau output kosong, semua file udah terformat.

### 11.3. `goimports` — formatter + import management

**Cara kerja:** sama seperti `gofmt`, plus memanage blok `import` — mengurutkan sesuai standard grouping (stdlib, third-party, internal project), dan menghapus import yang tidak terpakai.

**Install & pakai:**

```bash
go install golang.org/x/tools/cmd/goimports@latest

goimports -l .          # list file yang belum rapi
goimports -w file.go    # format + rapikan import
```

Banyak editor bisa di-set pakai `goimports` sebagai formatter pengganti `gofmt` di save.

### 11.4. `go vet` — linter bawaan (static analysis)

**Cara kerja:** `go vet` menganalisis AST + type info, lalu melaporkan konstruksi yang mencurigakan — bukan sekadar style. Contoh yang dia deteksi:

- Argument `Printf` tidak cocok dengan format string
- `unreachable code` (kode setelah `return`)
- `copy` dengan slice yang overlap
- Struct literal yang lupa field (pada assignment ke struct kosong)

**Perintah:**

```bash
go vet ./...
```

Jalankan sebelum commit, atau sisipkan di CI. Ini baseline yang wajib lewat di project mana pun.

### 11.5. `golangci-lint` — aggregator linter ala ESLint (opsional)

**Cara kerja:** satu binary yang menjalankan banyak linter sekaligus (paralel), lalu menampilkan hasil gabungan. Konfigurasi lewat file `.golangci.yml` — mirip `.eslintrc`.

**Install:**

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

**Init config minimal** — bikin `.golangci.yml` di root project (`backend/`):

```yaml
run:
  timeout: 5m

linters:
  enable:
    - errcheck # error harus di-handle atau di-ignore eksplisit
    - govet # go vet, otomatis jalan juga
    - ineffassign # deteksi assignment yang tidak efektif
    - staticcheck # linter populer dengan banyak rule
    - unused # deteksi kode yang tidak dipakai
```

**Perintah:**

```bash
golangci-lint run ./...
golangci-lint run ./... --fix   # auto-fix yang bisa diperbaiki
```

Catatan: `errcheck` sangat relevan dengan project ini — sesuai prinsip "Error is value", error nggak boleh di-swallow.

### 11.6. Alur kerja sehari-hari

```
menulis kode → editor format otomatis (gofmt/goimports)
            → go vet ./...       (baseline, wajib)
            → gofmt -l ./...     (pastikan output kosong)
            → golangci-lint run ./...  (kalau sudah setup, sebelum commit)
```

**Cheat sheet:**

| Kebutuhan                     | Perintah                        |
| ----------------------------- | ------------------------------- |
| Format semua file di package  | `go fmt ./...`                  |
| Cek file yang belum terformat | `gofmt -l ./...`                |
| Static analysis               | `go vet ./...`                  |
| Full lint                     | `golangci-lint run ./...`       |
| Full lint + auto-fix          | `golangci-lint run ./... --fix` |

---

## 12. Database Migration (golang-migrate)

Versioned SQL migration untuk schema. Struktur data berubah seiring waktu — migrasi menjaga perubahan itu tercatat, reproducible, dan bisa rollback.

### 12.1. Konsep

- Setiap perubahan schema = satu pasang file: `xxx.up.sql` (maju) dan `xxx.down.sql` (mundur).
- File diurutkan berdasarkan nomor: `000001_`, `000002_`, dst — diterapkan berurutan.
- Tool mencatat versi yang sudah jalan di tabel `schema_migrations` di DB — migrasi tidak dijalankan dua kali.
- **Aturan emas: jangan pernah mengedit file migrasi yang sudah di-apply.** Kalau butuh perubahan, buat file migrasi baru. Edit file lama = DB dan kode drift, down migration rusak.

```
backend/migrations/
├── 000001_create_teachers.up.sql
├── 000001_create_teachers.down.sql
├── 000002_add_students.up.sql
└── 000002_add_students.down.sql
```

### 12.2. Install CLI

```bash
go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Flag `-tags 'mysql'` **wajib** — tanpa itu driver MySQL/MariaDB tidak ikut ter-bundle ke binary. (MariaDB kompatibel dengan driver `mysql`.)

### 12.3. Inisiasi pertama

Pastikan DB sudah jalan (bikin DB-nya otomatis oleh docker compose — lihat `backend/docker-compose.yml`, variabel `MARIADB_DATABASE`):

```bash
cd backend
docker compose up -d db
```

Buat file migrasi pertama:

```bash
migrate create -ext sql -dir migrations -seq create_teachers
```

keterangan syntax:

- `-ext sql` → file migrasi pakai SQL, bukan Go.
- `-dir migrations` → folder tempat file migrasi.
- `-seq` → pakai nomor urut (sequence) untuk nama file, bukan timestamp.
- `create_teachers` → nama deskriptif untuk migrasi ini.

Hasilnya dua file kosong:

```
migrations/
├── 000001_create_teachers.up.sql
└── 000001_create_teachers.down.sql
```

Isi `000001_create_teachers.up.sql`:

```sql
CREATE TABLE IF NOT EXISTS teachers (
    id          INT AUTO_INCREMENT PRIMARY KEY,
    first_name  VARCHAR(100) NOT NULL,
    last_name   VARCHAR(100) NOT NULL,
    class       VARCHAR(50),
    subject     VARCHAR(100),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

Isi `000001_create_teachers.down.sql`:

```sql
DROP TABLE IF EXISTS teachers;
```

Sesuaikan kolom dengan domain model di `internal/models/` (contoh: `internal/models/teacher.go`).

### 12.4. Run up (apply migrasi)

```bash
migrate -path migrations -database 'mysql://app:app_password@tcp(127.0.0.1:3306)/school_management' up
```

Yang terjadi:

- Buat tabel `schema_migrations` (kalau belum ada) — catat versi `1`.
- Jalankan `000001_create_teachers.up.sql`.
- Dipanggil lagi → output `no change`, tidak jalan ulang.

### 12.5. Down (rollback)

```bash
# mundur 1 step (hapus tabel teachers)
migrate -path migrations -database 'mysql://app:app_password@tcp(127.0.0.1:3306)/school_management' down 1

# mundur semua migrasi
migrate -path migrations -database '...' down
```

`down 1` menjalankan `000001_create_teachers.down.sql` dan menurunkan versi di `schema_migrations`.

### 12.6. Cek status

```bash
migrate -path migrations -database 'mysql://app:app_password@tcp(127.0.0.1:3306)/school_management' version   # versi aktif sekarang
migrate -path migrations -database 'mysql://app:app_password@tcp(127.0.0.1:3306)/school_management' status    # semua file, applied / pending
```

### 12.7. Tambah table baru

Jangan sentuh `000001_` yang sudah jalan. Buat file baru:

```bash
migrate create -ext sql -dir migrations -seq add_students
```

```
migrations/
├── 000001_create_teachers.up.sql     # sudah applied, jangan diubah
├── 000001_create_teachers.down.sql
├── 000002_add_students.up.sql        # file baru
└── 000002_add_students.down.sql
```

Isi `000002_add_students.up.sql` (CREATE TABLE `students`), `.down.sql` (DROP), lalu:

```bash
migrate -path migrations -database 'mysql://app:app_password@tcp(127.0.0.1:3306)/school_management' up
```

Versi naik ke `2`. Langkah sama persis untuk tambah kolom, index, dll — selalu file migrasi baru.

### 12.8. Format DSN

MariaDB dipanggil lewat driver `mysql`. Format DSN untuk `-database`:

```
mysql://<user>:<password>@tcp(<host>:<port>)/<dbname>
```

Nilai default cocok dengan `backend/.env.example` dan `docker-compose.yml`:

```bash
# .env.example
DB_USER="app"
DB_PASSWORD="app_password"
DB_NAME="school_management"
DB_PORT=3306
DB_HOST=127.0.0.1
```

DSN lengkap: `mysql://app:app_password@tcp(127.0.0.1:3306)/school_management`. Bisa didefinisikan sebagai variabel shell agar tidak ketik ulang tiap command:

```bash
export DB_URL='mysql://app:app_password@tcp(127.0.0.1:3306)/school_management'
migrate -path migrations -database "$DB_URL" up
```

### 12.9. Opsional — run migrasi otomatis saat startup

Alternatif untuk produksi: embed file migrasi ke binary dengan `go:embed`, jalankan sebelum server start. Binary jadi self-contained — tidak butuh CLI migrate di server.

```go
//go:embed migrations/*.sql
var migrationsFS embed.FS

m, err := iofs.New(migrationsFS, "migrations")
if err != nil {
    return fmt.Errorf("load migrations: %w", err)
}
if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
    return fmt.Errorf("apply migrations: %w", err)
}
```

Catatan: untuk app multi-instance (beberapa replica jalan bareng), migrasi lebih aman dijalankan sebagai step terpisah di CI sebelum deploy, bukan saat tiap instance start.
