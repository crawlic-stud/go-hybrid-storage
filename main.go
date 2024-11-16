package main

import (
	"fmt"
	"hybrid-storage/handlers"
	"hybrid-storage/utils"
	"net/http"
)

func main() {
	fmt.Println("Starting server on localhost:8000")

	http.HandleFunc("/", utils.HttpHandler(handlers.Root, "GET"))
	http.HandleFunc("/files", utils.HttpHandler(handlers.UploadFile, "POST"))

	// mux handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/files/{id}", utils.HttpHandler(handlers.GetFilesHandler, "GET"))

	http.ListenAndServe(":8000", nil)
}
