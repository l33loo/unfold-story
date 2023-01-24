package main

import (
	"log"
	"net/http"
)

func main() {
	fileServer := http.FileServer(http.Dir("./public"))
	http.Handle("/", fileServer)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
