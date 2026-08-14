package middlewares

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"school-management-api/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
)

// JWTMiddleware mem-verifikasi token autentikasi dari request dan menyimpan
// identitas user (claims) ke request context.
// Token diambil dari cookie `auth_token` (set saat login), dengan fallback
// header `Authorization: Bearer <token>` untuk client non-browser / testing.
func JWTMiddleware(next http.Handler) http.Handler {
	fmt.Println("-------------------- JWT Middleware --------------------")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("++++++++++++ Inside JWT Middleware")

		// 1. Ambil token: prioritas cookie, fallback Bearer header
		tokenString, err := extractToken(r)
		if err != nil {
			http.Error(w, "Authorization Header Missing", http.StatusUnauthorized)
			return
		}

		// 2. Parse + verify token (signature, masa berlaku, algo).
		//    ParseToken (pkg/utils) sudah membungkus jwt/v5 + secret dari env.
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			// Beri pesan berbeda utk expired vs malformed (sesuai pola tutorial)
			if errors.Is(err, jwt.ErrTokenExpired) {
				http.Error(w, "Token Expired", http.StatusUnauthorized)
				return
			} else if errors.Is(err, jwt.ErrTokenMalformed) ||
				errors.Is(err, jwt.ErrTokenSignatureInvalid) {
				http.Error(w, "Token Invalid", http.StatusUnauthorized)
				return
			}
			utils.ErrorHandler(err, "jwt parse")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// 3. Simpan identitas user ke context supaya handler bisa membacanya
		//    via utils.GetClaims(ctx). Ikut menyimpan key individual mirip
		//    tutorial (role/expiresAt/userId) + email.
		ctx := context.WithValue(r.Context(), utils.ContextKeyClaims, claims)
		ctx = context.WithValue(ctx, utils.ContextKeyRole, claims.Role)
		ctx = context.WithValue(ctx, utils.ContextKeyUserID, claims.Sub)
		ctx = context.WithValue(ctx, utils.ContextKeyEmail, claims.Email)
		if claims.ExpiresAt != nil {
			ctx = context.WithValue(ctx, utils.ContextKeyExpiresAt, claims.ExpiresAt.Time)
		}

		// 4. Lanjut ke handler berikutnya
		next.ServeHTTP(w, r.WithContext(ctx))
		fmt.Println("Sent Response from JWT Middleware")
	})
}

// extractToken mengambil token dari cookie `auth_token` atau header Bearer.
func extractToken(r *http.Request) (string, error) {
	// Cookie (diprioritaskan — cara client browser mengirim token)
	if c, err := r.Cookie(utils.AuthCookieName); err == nil && c.Value != "" {
		return c.Value, nil
	}

	// Fallback: Authorization: Bearer <token>
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer "), nil
	}

	return "", errors.New("token not found")
}