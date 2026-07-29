# HANDS-OFF — Session Handover

> Update tiap akhir sesi agar agent berikutnya langsung nyambung.

---

## Project Overview

School Management API — backend HTTP service in Go. Masih tahap awal (tutorial phase).

---

## Current State (2026-07-29)

### Files

```
school-management/
├── AGENTS.md            ← role, prinsip, stack, checklist
├── docs/
│   ├── GUIDES.md        ← step-by-step: nambah fitur, testing, dll
│   └── HANDS-OFF.md     ← ini, handover tiap sesi
└── backend/
    ├── cmd/api/main.go   ← entry point, routing basic
    ├── internal/         ← masih kosong
    ├── pkg/              ← masih kosong
    ├── go.mod
    └── .gitignore
```

### Latest Commit

```
e0deadf — basic routing http methods
```

### Current Code

`cmd/api/main.go` — HTTP server dengan `net/http` stdlib, route `/`, `/teachers`, `/students`, `/execs`. Manual method switching pakai `switch r.Method`. Port hardcoded `:3000`, no graceful shutdown, logging pake `fmt.Println`.

---

## Architecture Decisions

| Keputusan | Detail |
|-----------|--------|
| **Router** | Akan pake Go 1.22 `http.ServeMux` dgn method routing (built-in) |
| **Logging** | Rencana migrasi ke `log/slog` |
| **Config** | `os.Getenv` + struct config |
| **DB** | `database/sql` + `sqlx` |
| **DI** | Manual constructor injection (no framework) |
| **Layers** | handler → service → repository (interface) → mysql (impl) |
| **Error handling** | Sentinel errors + wrapping + handler mapping |

---

## Session History

### Sesi #1 — 2026-07-29
- Init project structure (cmd, internal, pkg)
- Basic routing: root, teachers (GET/POST/PUT/PATCH/DELETE), students, execs
- Created `AGENTS.md` (role, prinsip, stack, checklist)
- Created `docs/GUIDES.md` (guide nambah fitur, DI, naming, error handling, middleware, testing, graceful shutdown)
- Created `docs/HANDS-OFF.md` (ini)

---

## Pending / Next Actions

Tidak ada task spesifik yang pending. Next sesi bisa mulai refactor `main.go`:
- [ ] Migrasi ke `http.NewServeMux` + method routing
- [ ] Extract config ke struct + env vars
- [ ] Tambah middleware (recovery, logging)
- [ ] Tambah graceful shutdown
- [ ] Init `internal/config/`, `internal/handler/middleware/`
- [ ] Bikin response helper di `pkg/response/`
- [ ] Bikin domain entity pertama (Teacher?)

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

# run server:
cd backend && go run ./cmd/api/
```
