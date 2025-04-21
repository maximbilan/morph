package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/morph/internal/app"
)

const port = 8080

func main() {
	log.Println("Starting morph on port", port)

	http.HandleFunc("/handler", app.Handle)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}
