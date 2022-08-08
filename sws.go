package main

import (
	"log"
	"net/http"
	"net/http/httputil"
)

// simple http web server

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		b, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Printf("error: %v\n", err)
		}
		log.Printf("%s\n", b)
		w.Write([]byte("OK"))
	})

	log.Println("listening on :5050")
	log.Fatal(http.ListenAndServe(":5050", nil))
}
