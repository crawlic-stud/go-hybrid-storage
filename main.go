package main

import (
	"fmt"
	"hybrid-storage/handlers"
	fileHandlers "hybrid-storage/handlers/file_handlers"
	"net/http"

	"github.com/rs/cors"
)

func main() {
	fmt.Println("Starting server on http://localhost:8000")

	handler := http.NewServeMux()

	handler.HandleFunc("GET /", handlers.Root)

	// handlers for files
	handler.HandleFunc("POST /files", fileHandlers.UploadFile)
	handler.HandleFunc("GET /files", fileHandlers.GetAllFilesHandler)
	handler.HandleFunc("GET /files/{id}", fileHandlers.GetFileHandler)
	handler.HandleFunc("PUT /files/{id}", fileHandlers.ReplaceFileHandler)
	handler.HandleFunc("DELETE /files/{id}", fileHandlers.DeleteFileHandler)

	// handlers for metadata
	handler.HandleFunc("GET /files/{id}/metadata", fileHandlers.GetMetadataHandler)
	handler.HandleFunc("PATCH /files/{id}/metadata", fileHandlers.ReplaceMetadataHandler)

	corsConfig := cors.New(cors.Options{
		AllowedHeaders:   []string{"Origin", "Authorization", "Accept", "Content-Type"},
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "PUT"},
		AllowCredentials: true,
	})

	corsHandler := corsConfig.Handler(handler)

	http.ListenAndServe(
		":8000",
		corsHandler,
	)
}
