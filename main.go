package main

import (
	"log"
	"net/http"
)

func main() {
	handler := http.Handler(NewServer())
	log.Fatal(http.ListenAndServe(":8000", handler))
}
