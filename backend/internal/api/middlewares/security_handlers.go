package middlewares

import "net/http"

// SecurityHeaders menambahkan header keamanan ke setiap response.
// Catatan: mayoritas header ini dirancang untuk web page, yang paling
// relevan untuk REST API adalah nosniff + HSTS.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// Browser tidak mengirim header Referer saat request dari halaman kita.
		w.Header().Set("Referrer-Policy", "no-referrer")

		// Remove atau set nama library atau technology yang digunakan sebagai backend
		w.Header().Set("X-Powered-By", "")

		next.ServeHTTP(w, r)
	})
}
