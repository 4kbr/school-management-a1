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
