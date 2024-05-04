package main

import (
	"log"
	"masspay/internal/app"
	"net/http"
)

func main() {
	router := app.NewRouter()
	log.Println("Starting server on :9000")
	if err := http.ListenAndServe(":9000", router); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
