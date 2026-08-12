package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// ===== Konstanta parameter Argon2id =====
// Nilai ini harus disimpan BERSAMA hash agar bisa diverifikasi ulang.
// Dipakai di format PHC string: $argon2id$v=19$m=<memory>,t=<time>,p=<threads>$<salt>$<hash>
const (
	argon2Time    uint32 = 1         // jumlah iterasi
	argon2Memory  uint32 = 64 * 1024 // memori dalam KiB (64 MB) — mahal untuk brute-force
	argon2Threads uint8  = 4         // paralelisme (cores)
	argon2KeyLen  uint32 = 32        // panjang hash output (byte)
)

// HashPassword mengenkripsi password plaintext dengan Argon2id lalu
// mengembalikan hash dalam format PHC string. Format ini menyimpan semua
// parameter + salt + hash, sehingga VerifyPassword bisa membaca ulang nilainya.
func HashPassword(password string) (string, error) {
	// 1. Generate salt acak 16 byte — unik tiap hash agar hash sama password beda.
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// 2. Hitung hash Argon2id dari password + salt + parameter di atas.
	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	// 3. Encode salt & hash ke base64 (tanpa padding '=' biar string bersih).
	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	hashB64 := base64.RawStdEncoding.EncodeToString(hash)

	// 4. Susun PHC string standar; kolom "v=19" = versi Argon2 yang dipakai.
	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argon2Memory, argon2Time, argon2Threads, saltB64, hashB64,
	)

	return encoded, nil
}

// VerifyPassword memeriksa apakah password cocok dengan hash PHC yang tersimpan.
// Mengembalikan true bila cocok, false bila tidak. Tidak pernah panic.
func VerifyPassword(encoded, password string) (bool, error) {
	// 1. Pecah PHC string menjadi bagian-bagian yang dipisah '$'.
	//    [0]="", [1]=argon2id, [2]=v=19, [3]=m=..,t=..,p=.., [4]=salt, [5]=hash
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid encoded hash format")
	}

	// 2. Validasi nama algoritma harus argon2id (bukan argon2i/dll).
	if parts[1] != "argon2id" {
		return false, errors.New("unsupported hash algorithm")
	}

	// 3. Baca versi (bagian [2], format "v=19").
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, errors.New("invalid argon2 version")
	}

	// 4. Baca parameter memori, waktu, thread (bagian [3], format "m=..,t=..,p=..").
	var memory uint32
	var time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false, errors.New("invalid argon2 parameters")
	}

	// 5. Decode salt & hash tersimpan dari base64.
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, errors.New("invalid salt encoding")
	}
	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, errors.New("invalid hash encoding")
	}

	// 6. Hitung ulang hash dari password dengan parameter yang SAMA seperti saat create.
	computedHash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(decodedHash)))

	// 7. Bandingkan dengan ConstantTimeCompare — waktu selalu sama (anti timing attack).
	if subtle.ConstantTimeCompare(computedHash, decodedHash) == 1 {
		return true, nil
	}
	return false, nil
}
