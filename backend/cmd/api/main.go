package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type User struct {
	Name string `json:"name"`
	Age  int16  `json:"age"`
	City string `json:"city"`
}

func main() {
	// todo: nanti pindahkan ke .env
	port := ":3000"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello root route"))
		fmt.Println("hello root route")
	})
	http.HandleFunc("/teachers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Write([]byte("hello GET Method on teachers route"))
			fmt.Println("hello GET Method on teachers route")
		case http.MethodPost:
			// parse form data (necessary for x-www-form-urlencoded)
			// r.Form is available only after calling ParseForm
			err := r.ParseForm()
			if err != nil {
				http.Error(w, "failed to parse form", http.StatusBadRequest)
				return
			}

			// iterate over parsed form data to inspect the values
			// r.Form is map[string][]string — supports duplicate keys
			response := make(map[string]interface{})
			for key, values := range r.Form {
				response[key] = values[0]
				fmt.Printf("form[%s] = %s\n", key, values[0])
				// for _, v := range values {
				// 	fmt.Printf("form[%s] = %s\n", key, v)
				// }
			}
			fmt.Println("processed response map:", response)

			// RAW body (json)
			// baca raw body, cocok untuk Content-Type: application/json
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read body", http.StatusBadRequest)
				return
			}
			// pakai defer untuk close function
			defer r.Body.Close()

			fmt.Println("isi RAW body:", body)
			fmt.Println("isi string RAW body:", string(body))

			// // untuk menyimpan jsonnya bisa seperti ini
			// var jsonData map[string]interface{}
			// atau explicit
			var jsonData User
			if err := json.Unmarshal(body, &jsonData); err != nil {
				http.Error(w, "invalid json body", http.StatusBadRequest)
				return
			}
			fmt.Println("json body:", jsonData)

			w.Write([]byte("hello POST Method on teachers route"))
		case http.MethodPut:
			w.Write([]byte("hello PUT Method on teachers route"))
			fmt.Println("hello PUT Method on teachers route")
		case http.MethodPatch:
			w.Write([]byte("hello PATCH Method on teachers route"))
			fmt.Println("hello PATCH Method on teachers route")
		case http.MethodDelete:
			w.Write([]byte("hello DELETE Method on teachers route"))
			fmt.Println("hello DELETE Method on teachers route")
		default:
			w.Write([]byte("hello others Method on teachers route"))
			fmt.Println("hello others Method on teachers route")
		}
	})
	http.HandleFunc("/students", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello students route"))
		fmt.Println("hello students route")
	})
	http.HandleFunc("/execs", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello execs route"))
		fmt.Println("hello execs route")
	})

	fmt.Println("server is running on port:", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalln("error starting the server", err)
	}
}
