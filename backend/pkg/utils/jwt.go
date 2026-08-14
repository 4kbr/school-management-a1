package utils

import (
	"context"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ContextKey adalah tipe khusus untuk key request context. Dipakai bikin key
// yang unique (public string biasa bisa bentrok antar package).
type ContextKey string

// Konstanta key untuk identitas user yang disimpan middleware JWT ke context.
const (
	ContextKeyClaims     ContextKey = "claims"
	ContextKeyRole      ContextKey = "role"
	ContextKeyUserID    ContextKey = "user_id"
	ContextKeyEmail     ContextKey = "email"
	ContextKeyExpiresAt ContextKey = "expires_at"
)

// ===== Konfigurasi JWT =====
// Nilai diambil dari env (ikuti pola config project: baca via os.Getenv).
const (
	// Nama cookie tempat token disimpan (di-set saat login, di-expire saat logout).
	AuthCookieName = "auth_token"
	// Umur default token (menit) dipakai kalau env JWT_EXPIRY_MIN tidak terbaca.
	defaultExpiryMin = 60
)

// Claims adalah struktur data yang disimpan di dalam token JWT.
// Sub/Email/Role = klaim kustom milik kita; RegisteredClaims = standar JWT
// (iss, exp, issued_at, dsb).
type Claims struct {
	Sub   int    `json:"sub"`
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

// jwtSecret membaca secret dari env JWT_SECRET.
func jwtSecret() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

// tokenExpiry membaca umur token (menit) dari env JWT_EXPIRY_MIN.
// Kalau tidak terbaca/parse gagal, pakai defaultExpiryMin.
func tokenExpiry() time.Duration {
	raw := os.Getenv("JWT_EXPIRY_MIN")
	if m, err := strconv.Atoi(raw); err == nil && m > 0 {
		return time.Duration(m) * time.Minute
	}
	return time.Duration(defaultExpiryMin) * time.Minute
}

// CreateToken membuat token JWT HS256 berisi identitas exec (id, email, role)
// + waktu kedaluwarsa. Token inilah yang dikirim ke client (cookie + JSON).
func CreateToken(execID int, email string, role string) (string, error) {
	// 1. Susun klaim token
	now := time.Now()
	claims := Claims{
		Sub:   execID,
		Email: email,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenExpiry())),
		},
	}

	// 2. Buat token dengan metode signing HS256 (symmetric key)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 3. Sign token memakai secret dari env
	tokenString, err := token.SignedString(jwtSecret())
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken memverifikasi signature + masa berlaku token, lalu mengembalikan
// klaim yang terkandung di dalamnya. Dipakai untuk membaca identitas pemegang token.
func ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	// 1. Parse + validasi token dengan secret yang sama seperti saat sign
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (interface{}, error) {
			// 2. Pastikan algoritma signing memang HS256 (anti algo-swapping)
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return jwtSecret(), nil
		},
	)
	if err != nil {
		return nil, err
	}

	// 3. Token valid → kembalikan klaim
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// GetClaims membaca identitas user (yang disimpan middleware JWT) dari request
// context. Mengembalikan claims + ok=false kalau belum ada status terautentikasi.
func GetClaims(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ContextKeyClaims).(*Claims)
	return claims, ok
}