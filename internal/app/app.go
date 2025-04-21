package app

import (
	"log"
	"net/http"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling...")
}
