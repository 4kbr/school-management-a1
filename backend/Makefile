# Load env vars dari .env (DB creds, dll) supaya tersedia di recipe.
# Catatan: TIDAK pakai `export` — karena aplikasi membaca .env sendiri lewat
# godotenv.Load() (yang otomatis membuang kutip). Kalau kita export, nilai
# berkutip ("app", ":3000", "cert.pem") jadi env var sungguhan dan mengalahkan
# parsing godotenv → bikin server gagal. Tanpa export, masalah itu hilang.
include .env

# Nilai .env berbentuk "app" (dengan kutip) — buang kutip untuk membangun URL migrasi.
DB_USER_C := $(subst ",,$(DB_USER))
DB_PASSWORD_C := $(subst ",,$(DB_PASSWORD))
DB_HOST_C := $(subst ",,$(DB_HOST))
DB_PORT_C := $(subst ",,$(DB_PORT))
DB_NAME_C := $(subst ",,$(DB_NAME))

# URL koneksi MySQL/MariaDB untuk golang-migrate, dibangun dari .env.
MIGRATE := migrate -path migrations -database 'mysql://$(DB_USER_C):$(DB_PASSWORD_C)@tcp($(DB_HOST_C):$(DB_PORT_C))/$(DB_NAME_C)'

.PHONY: dev run migrate-up migrate-down

# Jalankan dengan hot-reload (air)
dev:
	air

# Jalankan server biasa
run:
	go run ./cmd/api/

# Apply semua migrasi naik
migrate-up:
	$(MIGRATE) up

# Rollback semua migrasi turun
migrate-down:
	$(MIGRATE) down
