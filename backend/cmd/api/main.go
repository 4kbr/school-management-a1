package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type user struct {
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
			responseForm := make(map[string]interface{})
			for key, values := range r.Form {
				responseForm[key] = values[0]
				fmt.Printf("form[%s] = %s\n", key, values[0])
				// for _, v := range values {
				// 	fmt.Printf("form[%s] = %s\n", key, v)
				// }
			}
			fmt.Println("processed response map:", responseForm)

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
			// var userInstance map[string]interface{}
			// atau explicit
			var userInstance user

			// bisa seperti ini
			// err := json.Unmarshal(body, &userInstance);
			// if err != nil {...}
			//
			// atau bisa ini
			if err := json.Unmarshal(body, &userInstance); err != nil {
				http.Error(w, "invalid json body", http.StatusBadRequest)
				return
			}
			fmt.Println("Unmarshal json payload into userInstance:", userInstance)
			fmt.Println("Receive user name as:", userInstance.Name)

			// for response
			responseJson := make(map[string]interface{})
			for key, values := range r.Form {
				responseJson[key] = values[0]
			}

			// unmarshal json
			err = json.Unmarshal(body, &responseJson)
			if err != nil {
				return
			}

			fmt.Println("Unmarshal json response", responseJson)

			// access the request, apa yang dibawa r ini
			fmt.Println("Body:", r.Body)                         //cth: Body: &{0x3e1be33a2018 <nil> <nil> false true {{} {0 0}} true false false 0x64dac0}
			fmt.Println("Form:", r.Form)                         //cth: Form: map[]
			fmt.Println("Header:", r.Header)                     //cth: Header: map[Accept:[*/*] Accept-Encoding:[gzip, deflate, br] Cache-Control:[no-cache] Connection:[keep-alive] Content-Length:[73] Content-Type:[application/json] ...
			fmt.Println("Context:", r.Context())                 //cth: Context: context.Background.WithValue(net/http context value http-server, *http.Server).WithValue(net/http context value local-addr, [::1]:3000).WithCancel.WithCancel
			fmt.Println("ContentLength:", r.ContentLength)       //cth: ContentLength: 73
			fmt.Println("Host:", r.Host)                         //cth: Host: localhost:3000
			fmt.Println("Method:", r.Method)                     //cth: Method: POST
			fmt.Println("Proto:", r.Proto)                       //cth: Proto: HTTP/1.1
			fmt.Println("RemoteAddr:", r.RemoteAddr)             //cth: RemoteAddr: [::1]:36618
			fmt.Println("RequestURI:", r.RequestURI)             //cth: RequestURI: /teachers
			fmt.Println("TLS:", r.TLS)                           //cth: TLS: <nil>
			fmt.Println("Trailer:", r.Trailer)                   //cth: Trailer: map[]
			fmt.Println("TransferEncoding:", r.TransferEncoding) //cth: TransferEncoding: []
			fmt.Println("URL:", r.URL)                           //cth: URL: /teachers
			fmt.Println("UserAgent:", r.UserAgent())             //cth: UserAgent: PostmanRuntime/7.53.0
			fmt.Println("URL Port():", r.URL.Port())             //cth: URL Port():

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
