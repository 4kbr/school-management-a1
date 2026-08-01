# List command proses development

## create certificate

```bash
openssl req -x509 -newkey rsa:2048 -nodes -keyout key.pem -out cert.pem -days 365 -config openssl.cnf
```

keterangan:

| Flag | Arti |
|------|------|
| `req` | Membuat certificate request — dengan `-x509` langsung menghasilkan self-signed certificate tanpa CSR |
| `-x509` | Output langsung certificate self-signed (bukan CSR). Cocok untuk development |
| `-newkey rsa:2048` | Generate private key baru, RSA 2048-bit |
| `-nodes` | *no DES* — private key tanpa passphrase (biar gak diminta password tiap start server) |
| `-keyout key.pem` | Simpan private key ke file `key.pem` |
| `-out cert.pem` | Simpan certificate ke file `cert.pem` |
| `-days 365` | Masa berlaku 365 hari |
| `-config openssl.cnf` | Pakai konfigurasi dari `openssl.cnf` — isi cert & SAN otomatis, tanpa prompt interaktif |

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
