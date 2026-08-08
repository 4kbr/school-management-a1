package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	// === 1. method based routing: path sama, method beda ===
	mux.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "list items\n")
	})

	mux.HandleFunc("POST /items/create", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "item created\n")
	})

	// === 2. wildcard in pattern — path parameter: {name} = tepat satu segment ===
	mux.HandleFunc("GET /teachers/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "teacher id: %s\n", r.PathValue("id"))
	})

	// === 3. multiple path parameters: /path1/{param1}/path2/{param2} ===
	mux.HandleFunc("GET /teachers/{id}/students/{sid}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "teacher %s, student %s\n", r.PathValue("id"), r.PathValue("sid"))
	})

	// === 4. wildcard "...": tangkap sisa path (bisa multi-segment) ===
	// /files/a → a ; /files/a/b/c → a/b/c — WAJIB di segment paling akhir
	mux.HandleFunc("GET /files/{path...}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "file path: %s\n", r.PathValue("path"))
	})

	// === 5. wildcard di tengah path — VALID (asal single segment) ===
	// "/users/john/profile" → name=john
	mux.HandleFunc("GET /users/{name}/profile", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "profile: %s\n", r.PathValue("name"))
	})

	// === 6. {$} exact match: /items cuma match /items, gak match /items/ ===
	mux.HandleFunc("GET /items/{$}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "items exact match\n")
	})

	// === 7. root ===
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "root\n")
	})

	// ================================================================
	// CONTOH ERROR — uncomment satu-satu buat lihat panic saat register
	// ================================================================

	// a. "/{param2}/path2" — KATANYA error, padahal VALID kalau sendirian.
	//    Wildcard single segment boleh di posisi mana pun.
	//    Yang jadi error: DUA pattern begini sekaligus → konflik.
	// mux.HandleFunc("GET /{a}/x", func(w http.ResponseWriter, r *http.Request) {})
	// mux.HandleFunc("GET /b/{c}", func(w http.ResponseWriter, r *http.Request) {})
	// → panic: conflicting patterns /{a}/x and /b/{c} (request /b/x match dua-duanya)

	// b. wildcard {...} bukan di segment terakhir
	// mux.HandleFunc("GET /{a...}/b", func(w http.ResponseWriter, r *http.Request) {})
	// → panic: wildcard must be in final segment

	// c. dua wildcard {...}
	// mux.HandleFunc("GET /x/{a...}/{b}", func(w http.ResponseWriter, r *http.Request) {})
	// → panic: too many wildcards

	// d. malformed: brace gak ditutup / nama wildcard kosong
	// mux.HandleFunc("GET /a/{b", func(w http.ResponseWriter, r *http.Request) {})
	// → panic: pattern has malformed or empty wildcard

	// Catatan: "/items" (tanpa slash) vs "/items/" (dengan slash) itu pattern BEDA.
	// "/items/" tanpa {$} = subtree — request /items bakal 307 redirect ke /items/
	// → makanya dulu Postman "manggil" 2x.

	fmt.Println("modern routes running on :9090")
	fmt.Println("tes:")
	fmt.Println("  curl http://localhost:9090/items")
	fmt.Println("  curl -X POST http://localhost:9090/items/create")
	fmt.Println("  curl http://localhost:9090/teachers/123")
	fmt.Println("  curl http://localhost:9090/teachers/123/students/9")
	fmt.Println("  curl http://localhost:9090/files/a/b/c")
	fmt.Println("  curl http://localhost:9090/users/john/profile")
	http.ListenAndServe(":9090", mux)
}
