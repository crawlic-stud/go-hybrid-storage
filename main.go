package main

import (
	"fmt"
	"hybrid-storage/handlers"
	fileHandlers "hybrid-storage/handlers/file_handlers"
	"log"
	"net/http"
	"time"

	"github.com/rs/cors"
)

type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		lrw := &LoggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default status code
		}

		start := time.Now()
		next.ServeHTTP(lrw, r)
		log.Printf("%d %s %s %s in %v", lrw.statusCode, r.Method, r.URL.Path, r.URL.Query().Encode(), time.Since(start))
	})
}

func main() {
	portServe := ":8008"
	fmt.Println("Starting server on http://localhost" + portServe)

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

	loggingHandler := LoggingMiddleware(handler)
	corsHandler := corsConfig.Handler(loggingHandler)

	// client := &http.Client{
	// 	Timeout: 10 * time.Second,
	// 	Transport: &http.Transport{
	// 		MaxIdleConns:        100,
	// 		MaxIdleConnsPerHost: 100,
	// 	},
	// }

	http.ListenAndServe(
		portServe,
		corsHandler,
	)
}
