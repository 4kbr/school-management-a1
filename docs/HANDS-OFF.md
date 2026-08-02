# HANDS-OFF — Session Handover

> Update tiap akhir sesi agar agent berikutnya langsung nyambung.

---

## Project Overview

School Management API — backend HTTP service in Go. Masih tahap awal (tutorial phase).

---

## Current State (2026-08-01)

### Files

```
school-management/
├── AGENTS.md            ← role, prinsip, stack, checklist
├── docs/
│   ├── GUIDES.md        ← step-by-step: nambah fitur, testing, dll
│   └── HANDS-OFF.md     ← ini, handover tiap sesi
└── backend/
    ├── cmd/api/main.go          ← entry point, mux + TLS + middleware chain
    ├── internal/
    │   ├── api/
    │   │   ├── handlers/        ← kosong (placeholder)
    │   │   └── middlewares/     ← CORS, SecurityHeaders, ResponseTime
    │   ├── models/              ← kosong (placeholder)
    │   └── repositories/        ← kosong (placeholder)
    ├── docs/
    │   └── command.md           ← dokumentasi command openssl, run server
    ├── .air.toml                ← live reload config
    ├── cert.pem                 ← self-signed cert (gitignored)
    ├── key.pem                  ← private key (gitignored)
    ├── openssl.cnf              ← config buat generate cert
    ├── go.mod
    └── .gitignore
```

### Latest Commit

```
d15f16f — add cors middleware
```

### Current Code

`cmd/api/main.go` — HTTP server dengan `net/http` stdlib + `http.NewServeMux()`. Route `/`, `/teachers/`, `/students/`, `/execs/`. Manual method switching pakai `switch r.Method` di dalam handler. Port hardcoded `:3000`, TLS enabled (self-signed cert), 3 middlewares wired: `ResponseTime → SecurityHeaders → CORS → mux`. Logging masih `fmt.Println`. Belum ada graceful shutdown.

---

## Architecture Decisions

| Keputusan | Status | Detail |
|-----------|--------|--------|
| **Router** | ✅ Done | `http.NewServeMux()` — tapi masih switch r.Method (TODO: refactor ke method routing) |
| **TLS** | ✅ Done | Self-signed cert untuk dev, MinVersion TLS 1.2 |
| **Middleware** | ✅ Done | Chain: `ResponseTime → SecurityHeaders → CORS → mux` |
| **Logging** | ⏳ Belum | Masih `fmt.Println`, rencana migrasi ke `log/slog` |
| **Config** | ⏳ Belum | `os.Getenv` + struct config (port, cert path, key path) |
| **Graceful Shutdown** | ⏳ Belum | Belum diimplementasi |
| **DB** | ⏳ Belum | `database/sql` + `sqlx` |
| **DI** | ⏳ Belum | Manual constructor injection (no framework) |
| **Layers** | ⏳ Belum | handler → service → repository (interface) → mysql (impl) |
| **Error handling** | ⏳ Belum | Sentinel errors + wrapping + handler mapping |

---

## Middleware yang Sudah Ada

| Middleware | File | Fungsi |
|------------|------|--------|
| `ResponseTime` | `internal/api/middlewares/response_time.go` | Hitung durasi request, set header `X-Response-Time`, log details |
| `SecurityHeaders` | `internal/api/middlewares/security_handlers.go` | 11 security headers (HSTS, nosniff, CORP, Cache-Control, dll) |
| `Cors` | `internal/api/middlewares/cors.go` | Handle CORS: allowed origins, methods, headers, credentials |

---

## Session History

### Sesi #1 — 2026-07-29
- Init project structure (cmd, internal, pkg)
- Basic routing: root, teachers (GET/POST/PUT/PATCH/DELETE), students, execs
- Created `AGENTS.md` (role, prinsip, stack, checklist)
- Created `docs/GUIDES.md` (guide nambah fitur, DI, naming, error handling, middleware, testing, graceful shutdown)
- Created `docs/HANDS-OFF.md` (ini)

### Sesi #2 — 2026-08-01
- Implementasi `http.NewServeMux()` (masih switch r.Method)
- Setup TLS + self-signed cert (`openssl.cnf`, `cert.pem`, `key.pem`)
- Tambah middleware: CORS, SecurityHeaders (11 header), ResponseTime
- Created `backend/docs/command.md` (dokumentasi openssl)
- Added `backend/openssl.cnf` (SAN localhost config)

---

## Pending / Next Actions

- [ ] Refactor ke method routing: `mux.HandleFunc("GET /teachers", ...)` — biar gak perlu switch r.Method
- [ ] Extract config ke struct + env vars (port, cert path, key path)
- [ ] Tambah middleware recovery (panic handler)
- [ ] Tambah middleware logging (pake `slog`, bukan `fmt.Println`)
- [ ] Tambah graceful shutdown (`signal.NotifyContext`)
- [ ] Bikin response helper di `pkg/response/` (JSON response seragam)
- [ ] Bikin domain entity pertama (Teacher?)
- [ ] Init repository layer (interface + mysql impl)
- [ ] Init service layer
- [ ] Init handler layer (refactor dari main.go ke internal/api/handlers/)

---

## Critical Notes

- Jangan commit sembarangan — tunggu explicit request dari user
- Ikuti `AGENTS.md` checklist tiap code review
- Ikuti `docs/GUIDES.md` untuk step nambah fitur
- Prioritas: idiomatic Go, stdlib first, jangan over-engineering
- User pake bahasa Indonesia campur Inggris — sesuaikan tone

---

## Quick Start (for next agent)

```bash
# read these first:
# - AGENTS.md (prinsip & stack)
# - docs/HANDS-OFF.md (status terkini)
# - docs/GUIDES.md (cara nambah fitur)

# run server (HTTPS):
cd backend && air
# atau
cd backend && go run ./cmd/api/
# server jalan di https://localhost:3000

# test:
curl -k https://localhost:3000/teachers
```
