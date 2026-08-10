# List command proses development

## create certificate

```bash
openssl req -x509 -newkey rsa:2048 -nodes -keyout key.pem -out cert.pem -days 365 -config openssl.cnf
```

keterangan:

| Flag                  | Arti                                                                                                 |
| --------------------- | ---------------------------------------------------------------------------------------------------- |
| `req`                 | Membuat certificate request — dengan `-x509` langsung menghasilkan self-signed certificate tanpa CSR |
| `-x509`               | Output langsung certificate self-signed (bukan CSR). Cocok untuk development                         |
| `-newkey rsa:2048`    | Generate private key baru, RSA 2048-bit                                                              |
| `-nodes`              | _no DES_ — private key tanpa passphrase (biar gak diminta password tiap start server)                |
| `-keyout key.pem`     | Simpan private key ke file `key.pem`                                                                 |
| `-out cert.pem`       | Simpan certificate ke file `cert.pem`                                                                |
| `-days 365`           | Masa berlaku 365 hari                                                                                |
| `-config openssl.cnf` | Pakai konfigurasi dari `openssl.cnf` — isi cert & SAN otomatis, tanpa prompt interaktif              |

## kenapa pake openssl.cnf

- `prompt = no` — openssl gak tanya-tanya manual, semua field diisi dari file
- **SAN (Subject Alternative Name)** — browser modern (Chrome 58+) **menolak** certificate tanpa SAN, bahkan untuk `localhost`. `subjectAltName` di `openssl.cnf` membuat `localhost` & `127.0.0.1` valid
- `openssl.cnf` di-commit ke git (bukan rahasia) — jadi setup reproducible: siapapun bisa regenerate cert dengan command yang sama

## run server

```bash
# development (auto reload)
air

# manual
go run ./cmd/api/
```

## test server

```bash
# -k = skip certificate verification (cert self-signed), hanya untuk development
curl -k https://localhost:3000/teachers
```

## production notes

- Jangan pakai self-signed certificate di production — pakai certificate asli dari CA (mis. Let's Encrypt)
- Jangan pakai `curl -k` di production
- `key.pem` & `cert.pem` sudah di-ignore (lihat `backend/.gitignore`) — yang di-commit cuma `openssl.cnf` sebagai dokumentasi cara generate

## run database (Docker Compose)

Konfigurasi ada di `backend/docker-compose.yml`. Default env:

| Env                | Default             |
| ------------------ | ------------------- |
| `DB_ROOT_PASSWORD` | `root`              |
| `DB_NAME`          | `school_management` |
| `DB_USER`          | `app`               |
| `DB_PASSWORD`      | `app_password`      |

### Start database

```bash
# dari folder backend
docker compose up -d
```

### Cek status

```bash
docker compose ps
```

Tunggu sampai status `healthy` sebelum app Go konek (sekitar beberapa detik saat init pertama kali).

### Stop / start ulang

```bash
docker compose stop        # pause, data tetap
docker compose start       # start lagi
docker compose down        # hapus container (volume data TETAP ada)
```

### Reset total (hapus semua data)

```bash
docker compose down -v     # -v = hapus juga volume, data hilang permanen
```

## akses database via CLI

### Opsi A — via Docker (client dari dalam container, tanpa install apa-apa)

```bash
# masuk sesi interaktif mariadb
docker compose exec db mariadb -u app -papp_password school_management

# jalankan query langsung tanpa masuk shell
docker compose exec db mariadb -u app -papp_password school_management -e "SHOW DATABASES;"
docker compose exec db mariadb -u app -papp_password school_management -e "SHOW TABLES;"

# pakai user root (lebih penuh privileges)
docker compose exec db mariadb -u root -proot school_management
```

### Opsi B — langsung dari host (butuh client mysql/mariadb diinstall di Mac)

```bash
mariadb -h 127.0.0.1 -P 3306 -u app -papp_password school_management

#  atau pakai mysql cli jika mariadb belum ada
# mysql -h 127.0.0.1 -P 3306 -u app -papp_password school_management

```

Kredensial & DB awal sudah otomatis dibuat oleh compose di port `3306`.

## catatan

- Database & user dibuat sekali saat volume pertama kali di-init — ganti env name/kredensial sesudahnya hanya berlaku kalau volume di-reset (`docker compose down -v`).
- Jangan ubah kredensial via env setelah volume ada data — harus `docker compose down -v` dulu biar apply.

## Install mysql driver

`database/sql` tidak bisa jalan sendiri dan perlu install drivernya, salah satunya untuk mysql bisa pakai command ini

```bash
go get github.com/go-sql-driver/mysql
```

## Swagger / OpenAPI — cara menambah dokumentasi route baru

Dokumentasi di-generate otomatis dari anotasi comment di handler (pakai `swaggo/swag`). Jadi kalau menambah route baru, ikuti langkah ini.

### 1. Tambah anotasi di handler

Tulis comment `godoc` + anotasi `swag` **tepat di atas** function handler. Contoh pola dasar:

```go
// GetFooHandler godoc
// @Summary      Ringkasan singkat
// @Description  Penjelasan lebih panjang
// @Tags         teachers
// @Accept       json
// @Produce      json
// @Param        id    path  int                     true  "ID"
// @Param        body  body  dto.CreateTeacherRequest true "Request body"
// @Success      200   {object}  dto.TeacherListResponse
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /teachers/{id} [get]
func GetFooHandler(w http.ResponseWriter, r *http.Request) { ... }
```

Yang wajib:
- `@Summary`, `@Tags`, `@Router` — biar muncul di UI.
- `@Param` — untuk query/path/body/header parameter.
- `@Success`/`@Failure` — status code + tipe respons.
- `@Router` format: `path [method]`, method pakai lowercase (`get`, `post`, `put`, `patch`, `delete`).

Tipe `{object}` yang dipakai harus **bernama** (bukan anonymous struct) biar tergenerate rapi — makanya kita punya DTO di `internal/models/dto/` dan model di `internal/models/`.

### 2. Daftarkan route di `internal/api/router/router.go`

```go
mux.HandleFunc("GET /foo", handlers.GetFooHandler)
```

### 3. Generate ulang dokumentasi

```bash
# dari folder backend
make swagger
# atau manual
swag init -g cmd/api/main.go -o docs
```

Ini men-generate `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`.

### 4. Build & cek

```bash
go build ./...
```

Lalu jalankan server dan buka UI:

```bash
air
# buka https://localhost:3000/swagger/index.html
```

Cek route baru muncul di spec:
```bash
curl -sk https://localhost:3000/swagger/doc.json
```

### Catatan penting

- **`docs/` wajib di-generate sebelum build** — `router.go` meng-import `_ "school-management-api/docs"`, jadi kalau `docs/docs.go` belum ada, build gagal. Selalu `make swagger` setelah menambah/mengubah route.
- **Folder `docs/` di-commit** ke git (bukan di-ignore) supaya clone baru langsung bisa build.
- Setiap kali mengubah route/body/response, jalankan `make swagger` biar spec sinkron dengan kode.
- Akses UI lewat **HTTPS** karena server pakai `ListenAndServeTLS`.
- Anotasi `swag` untuk referensi lengkap: https://github.com/swaggo/swag

