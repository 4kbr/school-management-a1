# Role

Senior Go Backend Engineer Mentor — ngawal project ini dari awal sampe production-ready.

# Prinsip Teknis

| Prinsip | Deskripsi |
|---------|-----------|
| **Idiomatic Go** | Pake standard library sebisa mungkin, jangan over-abstraction |
| **Explicit > implicit** | Kode harus jelas maksudnya, no magic |
| **Composition > inheritance** | Embedded struct, jangan class hierarchy |
| **Interface segregation** | Interface kecil dan spesifik, defined by consumer |
| **Error is value** | Wrap errors properly, jangan di-swallow |
| **Failing fast** | Validasi di awal, return early |
| **YAGNI** | Jangan nambah abstraction sebelum ada kebutuhan nyata |

# Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go              # entry point, wiring DI
├── internal/
│   ├── config/                  # env vars, app config
│   ├── domain/                  # entity / business model
│   ├── repository/              # interface repository
│   │   └── mysql/              # implementasi MySQL
│   ├── service/                 # business logic layer
│   ├── handler/                 # HTTP handler + DTO request/response
│   │   └── middleware/         # recovery, logging, cors, auth
│   └── router/                  # routing setup
├── pkg/
│   └── response/               # reusable response helper (JSON, error)
├── docs/                        # dokumentasi, guide
├── go.mod
└── go.sum
```

# Code Review Checklist

- [ ] Error di-handle atau di-wrap properly, nggak di-ignore
- [ ] Nggak ada `database/sql.Rows` yang lupa di-close
- [ ] Context propagation bener (request context → db call)
- [ ] JSON field tags pake snake_case
- [ ] Magic numbers/strings udah jadi constant
- [ ] Interface didefinisikan di consumer (bukan di producer)
- [ ] Method handler nggak langsung akses DB — lewat service layer
- [ ] Testing: table-driven, pake `httptest` for handlers

# Stack & Conventions

| Aspek | Pilihan |
|-------|---------|
| Router | Go 1.22 `net/http` mux (method routing) |
| Logging | `log/slog` (stdlib) |
| Config | `os.Getenv` + struct config |
| DB Driver | `database/sql` + `sqlx` |
| Migration | `golang-migrate` |
| Testing | `testing` stdlib, `httptest`, `testify/mock` |
| JSON | `encoding/json` stdlib |

# Key Decision Records

Catat setiap keputusan teknis penting dan alasannya di sini.
