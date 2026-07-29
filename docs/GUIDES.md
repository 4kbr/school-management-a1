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

| Elemen | Convention | Contoh |
|--------|-----------|--------|
| File | `snake_case.go` | `teacher_handler.go` |
| Interface | `TeacherRepository` | — |
| Struct impl | lowercase + nama domain | `teacherRepository`, `teacherService` |
| Constructor | `New` prefix | `NewTeacherRepository`, `NewTeacherService` |
| Handler method | `Handle` prefix | `HandleGetAll`, `HandleCreate` |
| Request DTO | `CreateTeacherRequest` | — |
| Response DTO | `TeacherResponse` | — |

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
```
