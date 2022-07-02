package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("test", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("test"))
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
