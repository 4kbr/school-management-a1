package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	mw "school-management-api/internal/api/middlewares"
	"school-management-api/internal/api/router"
	"school-management-api/internal/repositories/sqlconnect"
	"school-management-api/pkg/utils"

	"github.com/joho/godotenv"
)

// handler teacher - END

// @title School Management API
// @version 1.0
// @description REST API untuk manajemen sekolah (teachers, dst). Dokumentasi ini di-generate otomatis dari anotasi kode oleh swag (swaggo/swag).
// @BasePath /
// @schemes https
func main() {

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	_, err = sqlconnect.ConnectDb()
	if err != nil {
		fmt.Println("Error---:", err)
		panic(err)
	}

	port := os.Getenv("API_PORT")

	// TLS: MinVersion TLS 1.2 — versi lama (SSLv3, TLS 1.0/1.1) sudah insecure
	// cert.pem & key.pem: self-signed, cuma untuk development.
	// Path dari .env (relatif ke CWD backend/). Cara generate: lihat backend/docs/command.md
	cert := os.Getenv("TLS_CERT")
	key := os.Getenv("TLS_KEY")
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// // init rate limiter
	// rl := mw.NewRateLimiter(10, time.Minute)

	// // setup hpp
	// hppOptions := mw.HPPOptions{
	// 	CheckQuery:                  true,
	// 	CheckBody:                   true,
	// 	CheckBodyOnlyForContentType: "application/json",
	// 	Whitelist:                   []string{"allowedParam"},
	// }

	// secureMux :=
	// 	mw.Cors(
	// 		mw.SecurityHeaders(
	// 			rl.Middleware(
	// 				mw.ResponseTime(
	// 					mw.Hpp(hppOptions)(
	// 						mw.Compression(
	// 							mux,
	// 						),
	// 					),
	// 				),
	// 			),
	// 		),
	// 	)
	// secureMux := ApplyMiddlewares(
	// 	mux, mw.Hpp(hppOptions), mw.Compression, mw.SecurityHeaders, mw.ResponseTime, rl.Middleware, mw.Cors,
	// )

	router := router.Router()
	secureMux := utils.ApplyMiddlewares(router, mw.SecurityHeaders)
	// create custom server
	server := &http.Server{
		Addr:    port,
		Handler: secureMux,
		// Handler: mw.ResponseTime(mw.SecurityHeaders(mw.Cors(mux))),
		// Handler:   mw.Cors(mux),
		TLSConfig: tlsConfig,
	}

	fmt.Println("server is running on port:", port)
	err = server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("error starting the server", err)
	}
}
