package middlewares

import (
	"fmt"
	"net/http"
	"strings"
)

// SecurityHeaders menambahkan header keamanan ke setiap response.
// Catatan: mayoritas header ini dirancang untuk web page, yang paling
// relevan untuk REST API adalah nosniff + HSTS.
func SecurityHeaders(next http.Handler) http.Handler {
	fmt.Println("HPP middleware...")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("SecurityHeaders middleware being returned...")
		// Matikan DNS prefetch — privasi: lokasi link tidak bocor via DNS lookup.
		w.Header().Set("X-DNS-Prefetch-Control", "off")

		// Anti clickjacking: halaman kita dilarang di-frame oleh situs lain.
		w.Header().Set("X-Frame-Options", "DENY")

		// Legacy: filter XSS bawaan browser (era IE/WebKit) — deprecated
		// di browser modern, tapi tetap aman dikirim untuk yang lama.
		w.Header().Set("X-XSS-Protection", "1;mode=block")

		// PENTING untuk API: larang MIME sniffing — browser tidak boleh
		// menebak content-type dari isi body (mis. JSON dibaca sebagai HTML).
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// HSTS: browser diwajibkan pakai HTTPS selama max-age detik (2 tahun).
		// Hanya berpengaruh kalau server dipanggil lewat HTTPS.
		w.Header().Set("Strict-Transport-Security", "max-age=63072000;includeSubDomains;preload")

		// Batasi sumber resource (script, img, dll) hanya dari origin sendiri.
		// Pengecualian: halaman Swagger UI (/swagger/) butuh inline style/script
		// + aset dari CDN (unpkg/jsdelivr), jadi diberi CSP yang lebih longgar.
		if strings.HasPrefix(r.URL.Path, "/swagger/") {
			w.Header().Set("Content-Security-Policy",
				"default-src 'self'; "+
					"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://unpkg.com https://cdn.jsdelivr.net; "+
					"style-src 'self' 'unsafe-inline' https://unpkg.com https://cdn.jsdelivr.net; "+
					"img-src 'self' data:; "+
					"font-src 'self' data: https://unpkg.com https://cdn.jsdelivr.net; "+
					"connect-src 'self'")
		} else {
			w.Header().Set("Content-Security-Policy", "default-src 'self'")
		}

		// Browser tidak mengirim header Referer saat request dari halaman kita.
		w.Header().Set("Referrer-Policy", "no-referrer")

		// Hapus header "Server: Go-http-server/1.1" bawaan Go — jangan expose
		// detail stack ke client. Nilai nil pada Header map = header tidak ditulis.
		w.Header()["Server"] = nil

		// X-Powered-By tidak dikirim Go secara default (itu bawaan Express).
		// Kalau mau hapus total, pakai nilai nil, bukan string kosong —
		// string kosong tetap menulis baris header kosong di response.
		w.Header()["X-Powered-By"] = nil

		// Legacy (era Flash/PDF): "none" = dokumen kita tidak boleh dijadikan
		// target cross-domain policy oleh produk lain.
		w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")

		// Data API tidak boleh di-cache browser/proxy (bisa berisi data sensitif).
		w.Header().Set("Cache-Control", "no-store")

		// CORP: resource boleh di-load dari origin mana pun (cross-origin),
		// karena API nanti dipanggil frontend dari domain berbeda.
		// Bedanya dengan CORS: ini isolasi di level resource (anti
		// side-channel attack), bukan permission untuk akses.
		w.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")

		next.ServeHTTP(w, r)

		fmt.Println("SecurityHeaders middleware ends...")
	})
}
